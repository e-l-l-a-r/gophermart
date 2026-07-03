package handler

import (
	"encoding/json"
	"net/http"

	"github.com/e-l-l-a-r/gophermart/internal/model"
)

func balanceReq() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user := ctx.Value("user").(*model.User)

		resp.Header().Set("Content-Type", "application/json")

		// сериализуем ответ сервера
		enc := json.NewEncoder(resp)
		if err := enc.Encode(user.GetBalance()); err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}
