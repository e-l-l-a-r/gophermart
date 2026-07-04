package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandlers(t *testing.T) {
	logger.InitLogger("info")

	mockStorage := new(repository.MockStorage)

	t.Run("Register success", func(t *testing.T) {
		authData := model.AuthData{
			Login:    "testuser",
			Password: "testpassword",
		}
		body, _ := json.Marshal(authData)
		req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
		ctx := repository.ContextWithStorage(req.Context(), mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		mockStorage.On("AddUser", mock.Anything, mock.Anything).Return(nil).Once()
		mockStorage.On("CreateSession", mock.Anything, mock.Anything).Return("test-session-uuid", nil).Once()

		registerReq().ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Header().Get("Set-Cookie"), "session=test-session-uuid")
		mockStorage.AssertExpectations(t)
	})

	t.Run("Login success", func(t *testing.T) {
		authData := model.AuthData{
			Login:    "testuser",
			Password: "testpassword",
		}
		body, _ := json.Marshal(authData)
		req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
		ctx := repository.ContextWithStorage(req.Context(), mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		mockStorage.On("CreateSession", mock.Anything, mock.Anything).Return("login-session-uuid", nil).Once()

		loginReq().ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Header().Get("Set-Cookie"), "session=login-session-uuid")
		mockStorage.AssertExpectations(t)
	})

	t.Run("Login unauthorized", func(t *testing.T) {
		authData := model.AuthData{
			Login:    "testuser",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(authData)
		req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
		ctx := repository.ContextWithStorage(req.Context(), mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		mockStorage.On("CreateSession", mock.Anything, mock.Anything).Return("", &repository.ErrNoData{}).Once()

		loginReq().ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
		mockStorage.AssertExpectations(t)
	})
}
