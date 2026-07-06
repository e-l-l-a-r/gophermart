package worker

import (
	"context"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

func UpdOrderData(storage repository.Storage, ch chan model.Order) {
	ctx := repository.ContextWithStorage(context.Background(), storage)

	log, err := logger.GetLogger()
	if err != nil {
		logger.Warn(err)
	}

	log.Info("Starting orders update thread")
	for {
		order, ok := <-ch
		if !ok {
			return
		}

		order.SaveToDb(ctx)
	}

	ctx.Done()
}
