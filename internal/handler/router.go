package handler

import (
	"github.com/e-l-l-a-r/gophermart/internal/repository"
	"github.com/go-chi/chi/v5"
)

var (
	routesInstalled bool
	rtr             *chi.Mux
)

func GetRouter(storage repository.Storage) *chi.Mux {
	if routesInstalled {
		return rtr // Return existing router
	}
	routesInstalled = true
	rtr = chi.NewRouter()

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
