package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
)

func getAuthData(req *http.Request) (auth model.AuthData, err error) {

	dec := json.NewDecoder(req.Body)

	if err = dec.Decode(&auth); err != nil {
		err = logger.NewTracedError("Incorrect data", err)
		logger.Info("cannot decode request JSON body", err)
		return
	}

	if !auth.IsValid() {
		err = fmt.Errorf("некорректный запрос")
	}

	return
}
func registerReq() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		auth, err := getAuthData(req)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := model.AddUser(ctx, auth)
		if err != nil {
			if _, ok := errors.AsType[*logger.TracedError](err); ok {
				http.Error(resp, "Пользователь с таким логином уже существует", http.StatusConflict)
			}
			return
		}
		http.SetCookie(resp, getCookeForUser(*user))
		resp.Write([]byte("OK"))
		return
	}
}

func loginReq() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		auth, err := getAuthData(req)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := model.LoginUser(ctx, auth)

		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		if !user.IsOk() {
			http.Error(resp, "Неверный логин или пароль", http.StatusUnauthorized)
			return
		}

		http.SetCookie(resp, getCookeForUser(*user))
		resp.Write([]byte("OK"))
		return

	}
}
