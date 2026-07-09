package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddOrderReq(t *testing.T) {
	logger.InitLogger("info")

	mockStorage := new(repository.MockStorage)

	handler := addOrderReq()

	t.Run("success", func(t *testing.T) {
		user := model.NewUser("testuser")
		req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678903"))
		ctx := context.WithValue(req.Context(), "user", user)
		ctx = repository.ContextWithStorage(ctx, mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		mockStorage.On("AddNewOrder", mock.Anything, "12345678903", "testuser").Return("", nil).Once()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusAccepted, resp.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("invalid order number", func(t *testing.T) {
		user := model.NewUser("testuser")
		req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678904")) // Invalid Luhn
		ctx := context.WithValue(req.Context(), "user", user)
		ctx = repository.ContextWithStorage(ctx, mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	})

	t.Run("already added by same user", func(t *testing.T) {
		user := model.NewUser("testuser")
		req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678903"))
		ctx := context.WithValue(req.Context(), "user", user)
		ctx = repository.ContextWithStorage(ctx, mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		mockStorage.On("AddNewOrder", mock.Anything, "12345678903", "testuser").Return("testuser", nil).Once()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "уже был добавлен ранее")
		mockStorage.AssertExpectations(t)
	})

	t.Run("already added by another user", func(t *testing.T) {
		user := model.NewUser("testuser")
		req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString("12345678903"))
		ctx := context.WithValue(req.Context(), "user", user)
		ctx = repository.ContextWithStorage(ctx, mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		mockStorage.On("AddNewOrder", mock.Anything, "12345678903", "testuser").Return("anotheruser", nil).Once()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusConflict, resp.Code)
		mockStorage.AssertExpectations(t)
	})
}
