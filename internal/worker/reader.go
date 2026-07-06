package worker

import (
	"context"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

const getOrdersCnt = 10

func DoWork(storage repository.Storage, doneCh chan struct{}, baseUrl string) {
	ch := make(chan model.Order, getOrdersCnt)
	chOut := make(chan model.Order, getOrdersCnt)
	log, err := logger.GetLogger()
	if err != nil {
		logger.Warn(err)
	}

	// Горутина gолучает из БД список заказов для обработки и отправляет их в канал
	go func() {
		ctx := repository.ContextWithStorage(context.Background(), storage)
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
				for _, item := range orders {
					ch <- item
				}
			}
		}
		ctx.Done()
	}()

	go GetOrderStatus(baseUrl, ch, chOut)

	go UpdOrderData(storage, chOut)

	return
}
