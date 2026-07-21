package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

func balanceReq() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user := ctx.Value("user").(*model.User)

		resp.Header().Set("Content-Type", "application/json")

		// сериализуем ответ сервера
		enc := json.NewEncoder(resp)
		if err := enc.Encode(user.GetBalance()); err != nil {
			http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logger.Err(err.Error())
			return
		}

	}
}

func balanceWithdrawReq() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user := ctx.Value("user").(*model.User)
		defer req.Body.Close()

		dec := json.NewDecoder(req.Body)

		var withdraw model.Withdraw

		if err := dec.Decode(&withdraw); err != nil {
			err = logger.NewTracedError("Incorrect data", err)
			logger.Debug("cannot decode request JSON body", err)
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		// Проверяем валидность номера заказа
		order := model.Order{Number: withdraw.Order}
		if !order.IsValid() {
			http.Error(resp, "invalid order number", http.StatusUnprocessableEntity)
			return
		}

		err := withdraw.Add(ctx, user.GetLogin())
		if err != nil {
			if _, ok := errors.AsType[*repository.ErrNoData](err); ok {
				http.Error(resp, "insufficient funds", http.StatusPaymentRequired)
				return
			}
			http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logger.Err(err.Error())
			return
		}

		resp.Write([]byte("Бонусы успешно списаны"))

	}
}

func getwithdrawalsReq() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		user := ctx.Value("user").(*model.User)

		defer req.Body.Close()

		orders, err := model.GetWithdrawalsForUser(ctx, user.GetLogin())

		if err != nil {
			http.Error(resp, "Неизвестная ошибка", http.StatusInternalServerError)
			return
		}

		if len(orders) == 0 {
			resp.WriteHeader(http.StatusNoContent)
			return
		}

		resp.Header().Set("Content-Type", "application/json")

		// сериализуем ответ сервера
		enc := json.NewEncoder(resp)
		if err := enc.Encode(orders); err != nil {
			http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logger.Err(err.Error())
		}
	}
}
