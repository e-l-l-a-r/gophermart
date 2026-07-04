package repository

import (
	"context"
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

type Storage interface {
	AddUser(ctx context.Context, user itfUser) error
	CreateSession(ctx context.Context, user itfUser) (session string, err error)
	GetUserBySession(ctx context.Context, session string) (data UserData, err error)
	AddNewOrder(ctx context.Context, orderNum string, userNm string) (username string, err error)
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
