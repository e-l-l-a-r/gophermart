package handler

import (
	"github.com/e-l-l-a-r/gophermart/internal/repository"
	"github.com/go-chi/chi/v5"
)

func GetRouter(storage repository.Storage) *chi.Mux {
	rtr := chi.NewRouter()

	rtr.Use(WithStorage(storage))

	rtr.Get("/", incorrectReqest)
	rtr.Post("/", incorrectReqest)

	rtr.Post("/api/user/register", registerReq())
	rtr.Post("/api/user/login", loginReq())

	rtr.Post("/api/user/orders", CheckAuth(addOrderReq()))
	rtr.Get("/api/user/orders", CheckAuth(getOrdersReq()))
	rtr.Get("/api/user/balance", CheckAuth(balanceReq()))

	rtr.Post("/api/user/balance/withdraw", CheckAuth(balanceWithdrawReq()))

	rtr.Get("/api/user/withdrawals", CheckAuth(getwithdrawalsReq()))

	return rtr
}
