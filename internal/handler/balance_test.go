package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestBalanceReq(t *testing.T) {
	logger.InitLogger("info")

	mockStorage := new(repository.MockStorage)
	handler := balanceReq()

	t.Run("success", func(t *testing.T) {
		user := model.NewUser("testuser")
		user.SetBalance(123.45, 10)

		req := httptest.NewRequest("GET", "/api/user/balance", nil)
		ctx := context.WithValue(req.Context(), "user", user)
		ctx = repository.ContextWithStorage(ctx, mockStorage)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "application/json", resp.Header().Get("Content-Type"))

		var balance model.Balance
		err := json.NewDecoder(resp.Body).Decode(&balance)
		assert.NoError(t, err)
		assert.Equal(t, 123.45, balance.Current)
		assert.Equal(t, 10, balance.Withdrawn)
	})
}
