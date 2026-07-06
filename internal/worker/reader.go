package worker

import (
	"context"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

const getOrdersCnt = 10

func DoWork(storage repository.Storage, doneCh chan struct{}, baseUrl string) {
	ch := make(chan model.Order, 1)
	chOut := make(chan model.Order, getOrdersCnt)
	log, err := logger.GetLogger()
	if err != nil {
		logger.Warn(err)
	}

	// Горутина gолучает из БД список заказов для обработки и отправляет их в канал
	go func() {
		ctx := repository.ContextWithStorage(context.Background(), storage)
		log.Info("Starting orders reader thread")
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
					time.Sleep(time.Duration(500) * time.Millisecond)
				} else {
					for _, item := range orders {
						ch <- item
					}
					if len(orders) < getOrdersCnt {
						// Если данных не полная пачка, делаем небольшую паузу, чтоб не долбиться в БД
						time.Sleep(time.Duration(3) * time.Second)
					}
				}
			}
		}
		ctx.Done()
	}()

	go GetOrderStatus(baseUrl, ch, chOut)

	go UpdOrderData(storage, chOut)

	return
}
