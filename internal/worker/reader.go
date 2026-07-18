package worker

import (
	"context"
	"sync"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

const getOrdersCnt = 10

func DoWork(storage repository.Storage, doneCh chan struct{}, baseUrl string) *sync.WaitGroup {
	ch := make(chan model.Order, 1)
	chOut := make(chan model.Order, getOrdersCnt)
	log, err := logger.GetLogger()
	if err != nil {
		logger.Warn(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ctx = repository.ContextWithStorage(ctx, storage)

	// Горутина gолучает из БД список заказов для обработки и отправляет их в канал
	go func() {
		log.Info("Starting orders reader thread")
		defer cancel()
		for {
			select {
			case <-doneCh:
				close(ch)
				return
			default:
				log.Info("ReadData")
				orders, err := model.GetOrdersToProcess(ctx, getOrdersCnt)
				if err != nil {
					log.Error(err.Error())
					continue
				}
				if len(orders) == 0 {
					// Если данных нет, делаем небольшую паузу, чтоб не долбиться в БД
					timer := time.NewTimer(time.Duration(500) * time.Millisecond)
					select {
					case <-timer.C:
						continue
					case <-doneCh:
						close(ch)
						if !timer.Stop() {
							<-timer.C
						}
						return
					}
				} else {
					for _, item := range orders {
						select {
						case ch <- item:
						case <-doneCh:
							close(ch)
							return
						}
					}
					if len(orders) < getOrdersCnt {
						// Если данных не полная пачка, делаем небольшую паузу, чтоб не долбиться в БД
						timer := time.NewTimer(time.Duration(3) * time.Second)
						select {
						case <-timer.C:
							continue
						case <-doneCh:
							close(ch)
							if !timer.Stop() {
								<-timer.C
							}
							return
						}
					}
				}
			}
		}
	}()

	wg := &sync.WaitGroup{}

	n := 5
	wg.Add(n)
	for i := 0; i < n; i++ {
		go GetOrderStatus(ctx, baseUrl, ch, chOut, wg)
	}

	go func() {
		wg.Wait()
		close(chOut)
	}()

	go UpdOrderData(ctx, chOut)

	return wg
}
