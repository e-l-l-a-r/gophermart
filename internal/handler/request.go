package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

func incorrectReqest(resp http.ResponseWriter, _ *http.Request) {
	http.Error(resp, "Incorrect API", http.StatusBadRequest)
}

func okRequest(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("OK"))
}

func serverError(msg string, resp http.ResponseWriter) {
	http.Error(resp, "Internal ERROR "+msg, http.StatusInternalServerError)
}
func getCookeForUser(user model.User) *http.Cookie {
	return &http.Cookie{
		Name:     "session",
		Value:    user.GetSessionKey(),
		Expires:  time.Now().Add(time.Minute * 10),
		Path:     "/",
		HttpOnly: true,
	}
}
func WithStorage(storage repository.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := repository.ContextWithStorage(r.Context(), storage)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		session, err := req.Cookie("session")
		if err != nil {
			http.Error(resp, "Unauthorized", http.StatusUnauthorized)
			return
		}
		user, err := model.GetUserBySession(req.Context(), session.Value)

		if err != nil {
			if _, ok := errors.AsType[*repository.ErrNoData](err); ok {
				http.Error(resp, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			logger.Err(err.Error())
			return
		}

		if user == nil {
			http.Error(resp, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userCtx := context.WithValue(req.Context(), "user", user)
		next(resp, req.WithContext(userCtx))
	}
}
