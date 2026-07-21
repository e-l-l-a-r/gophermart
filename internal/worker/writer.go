package worker

import (
	"context"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
)

func UpdOrderData(ctx context.Context, ch chan model.Order) {
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

		if err := order.SaveToDb(ctx); err != nil {
			log.WarnMsg("error saving order to db: ", err)
		}
	}

}
