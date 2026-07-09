package model

import (
	"context"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (w *Withdraw) Add(ctx context.Context, userNm string) error {
	storage, ok := repository.FromContext(ctx)
	if !ok {
		return logger.NewTracedError("storage not found in context", nil)
	}

	err := storage.AddNewWithdraw(ctx, w.Order, userNm, w.Sum)

	if err != nil {
		return logger.NewTracedError("error adding withdraw: ", err)
	}

	return nil
}

func GetWithdrawalsForUser(ctx context.Context, userNm string) ([]Withdraw, error) {
	storage, ok := repository.FromContext(ctx)
	if !ok {
		return nil, logger.NewTracedError("storage not found in context", nil)
	}

	data, err := storage.GetWithdrawalsList(ctx, userNm)

	if err != nil {
		return nil, logger.NewTracedError("error getting orders: ", err)
	}

	res := make([]Withdraw, 0, len(data))

	for i := 0; i < len(data); i++ {
		if err == nil {
			withdraw := Withdraw{
				data[i].Number,
				data[i].Sum,
				data[i].Processed,
			}
			res = append(res, withdraw)
		}
	}

	return res, nil
}
