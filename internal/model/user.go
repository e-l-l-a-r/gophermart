package model

import (
	"context"

	"github.com/e-l-l-a-r/gophermart/internal/crypto"
	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

type AuthData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	session  string
}

func (a *AuthData) IsValid() bool {
	if a.Login == "" || a.Password == "" {
		return false
	}
	return true
}

type User struct {
	auth        AuthData
	orders      []Order
	withdrawals []Withdraw
	balance     Balance
}

func (u *User) GetLogin() string {
	return u.auth.Login
}

func (u *User) GetAuthKey() string {
	return crypto.GetAuthKey(u.auth.Login, u.auth.Password)
}

func (u *User) GetSessionKey() string {
	return u.auth.session
}

func (u *User) IsOk() bool {
	return u.auth.session != ""
}

func (u *User) GetBalance() Balance {
	return u.balance
}

func NewUser(login string) *User {
	return &User{
		auth: AuthData{
			Login: login,
		},
	}
}

func (u *User) SetBalance(current float64, withdrawn int) {
	u.balance = Balance{
		Current:   current,
		Withdrawn: withdrawn,
	}
}

func AddUser(ctx context.Context, data AuthData) (*User, error) {
	user := User{
		auth: data,
	}

	storage, ok := repository.FromContext(ctx)
	if !ok {
		return nil, logger.NewTracedError("storage not found in context", nil)
	}

	err := storage.AddUser(ctx, &user)

	if err != nil {
		return nil, logger.NewTracedError("error adding user: ", err)
	}

	session, _ := storage.CreateSession(ctx, &user)

	user.auth.session = session

	return &user, nil

}

func LoginUser(ctx context.Context, data AuthData) (*User, error) {
	user := User{
		auth: data,
	}

	storage, ok := repository.FromContext(ctx)
	if !ok {
		return nil, logger.NewTracedError("storage not found in context", nil)
	}

	session, err := storage.CreateSession(ctx, &user)

	if err != nil {
		return nil, logger.NewTracedError("error creating session: ", err)
	}

	user.auth.session = session

	return &user, nil
}

func GetUserBySession(ctx context.Context, session string) (*User, error) {

	storage, ok := repository.FromContext(ctx)
	if !ok {
		return nil, logger.NewTracedError("storage not found in context", nil)
	}

	userData, err := storage.GetUserBySession(ctx, session)
	if err != nil {
		return nil, logger.NewTracedError("error getting user data: ", err)
	}

	if userData.Login == "" {
		return nil, nil
	}

	return &User{
		auth: AuthData{
			Login:    userData.Login,
			Password: "***",
			session:  session,
		},
		balance: Balance{
			Current:   userData.Current,
			Withdrawn: userData.Withdrawn,
		},
	}, nil
}
