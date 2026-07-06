package repository

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) AddUser(ctx context.Context, user itfUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockStorage) CreateSession(ctx context.Context, user itfUser) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetUserBySession(ctx context.Context, session string) (UserData, error) {
	args := m.Called(ctx, session)
	return args.Get(0).(UserData), args.Error(1)
}

func (m *MockStorage) AddNewOrder(ctx context.Context, orderNum string, userNm string) (string, error) {
	args := m.Called(ctx, orderNum, userNm)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetOrdesList(ctx context.Context, userNm string, lim_stt int, count int) ([]OrderData, error) {
	args := m.Called(ctx, userNm, lim_stt, count)
	return args.Get(0).([]OrderData), args.Error(1)
}

func (m *MockStorage) UpdOrderData(ctx context.Context, orderNum string, orserStt int, accrual float64) error {
	args := m.Called(ctx, orderNum, orserStt, accrual)
	return args.Error(0)
}

func (m *MockStorage) Close() {
	m.Called()
}
