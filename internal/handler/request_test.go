package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckAuth(t *testing.T) {
	logger.InitLogger("info")

	mockStorage := new(repository.MockStorage)

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*model.User)
		w.Write([]byte(user.GetLogin()))
	})

	handler := CheckAuth(innerHandler)

	t.Run("authorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		ctx := repository.ContextWithStorage(req.Context(), mockStorage)
		req = req.WithContext(ctx)
		req.AddCookie(&http.Cookie{Name: "session", Value: "valid-session"})
		resp := httptest.NewRecorder()

		mockStorage.On("GetUserBySession", mock.Anything, "valid-session").Return(repository.UserData{
			Login:     "testuser",
			Current:   100.0,
			Withdrawn: 0,
		}, nil).Once()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "testuser", resp.Body.String())
		mockStorage.AssertExpectations(t)
	})

	t.Run("unauthorized no cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("unauthorized invalid session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		ctx := repository.ContextWithStorage(req.Context(), mockStorage)
		req = req.WithContext(ctx)
		req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-session"})
		resp := httptest.NewRecorder()

		mockStorage.On("GetUserBySession", mock.Anything, "invalid-session").Return(repository.UserData{}, &repository.ErrNoData{}).Once()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		mockStorage.AssertExpectations(t)
	})
}
