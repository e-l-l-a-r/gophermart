package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/model"
)

func incorrectReqest(resp http.ResponseWriter, _ *http.Request) {
	http.Error(resp, "Incorrect API", http.StatusBadRequest)
}

func okRequest(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("OK"))
}

func serverError(msg string, resp http.ResponseWriter) {
	http.Error(resp, "Interval ERROR", http.StatusInternalServerError)
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
func CheckAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		session, err := req.Cookie("session")
		if err != nil {
			http.Error(resp, "Unauthorized", http.StatusUnauthorized)
			return
		}
		user, err := model.GetUserBySession(req.Context(), session.Value)

		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
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
