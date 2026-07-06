package repository

import (
	"context"
	"time"
)

type itfUser interface {
	GetLogin() string
	GetAuthKey() string
}

type UserData struct {
	Login     string
	Current   float64
	Withdrawn int
}

type OrderData struct {
	Number   string
	Status   int
	Accrual  float64
	Uploaded time.Time
}

type Storage interface {
	AddUser(ctx context.Context, user itfUser) error
	CreateSession(ctx context.Context, user itfUser) (session string, err error)
	GetUserBySession(ctx context.Context, session string) (data UserData, err error)
	AddNewOrder(ctx context.Context, orderNum string, userNm string) (username string, err error)
	GetOrdesList(ctx context.Context, userNm string, lim_stt int, count int) (data []OrderData, err error)
	UpdOrderData(ctx context.Context, orderNum string, orserStt int, accrual float64) error
	Close()
}

type storageKey struct{}

func ContextWithStorage(ctx context.Context, s Storage) context.Context {
	return context.WithValue(ctx, storageKey{}, s)
}

func FromContext(ctx context.Context) (Storage, bool) {
	s, ok := ctx.Value(storageKey{}).(Storage)
	return s, ok
}
