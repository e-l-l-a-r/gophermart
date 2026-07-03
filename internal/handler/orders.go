package handler

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/model"
)

func addOrderReq() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		user := ctx.Value("user").(*model.User)

		defer req.Body.Close()

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(resp, "failed to read request body", http.StatusInternalServerError)
			return
		}

		order := model.Order{
			Number:     string(body),
			Status:     model.NEW,
			UploadedAt: time.Now(),
		}

		if !order.IsValid() {
			http.Error(resp, "invalid order number", http.StatusUnprocessableEntity)
			return
		}

		err = order.AddToUser(ctx, user)

		if err != nil {
			if _, ok := errors.AsType[*model.ErrAlreadyExists](err); ok {
				resp.Write([]byte("Данный заказ уже был добавлен ранее"))
				return
			}
			if _, ok := errors.AsType[*model.ErrAddedByAnotherUser](err); ok {
				http.Error(resp, "Данный заказ уже был добавлен другим пользователем", http.StatusConflict)
				return
			}
			http.Error(resp, "Неизвестная ошибка", http.StatusInternalServerError)
		}

		resp.WriteHeader(http.StatusAccepted)
		resp.Write([]byte("Заказ успешно добавлен"))

	}
}
