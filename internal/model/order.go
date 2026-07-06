package model

import (
	"context"
	"fmt"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

const (
	NEW        = "NEW"
	PROCESSING = "PROCESSING"
	INVALID    = "INVALID"
	PROCESSED  = "PROCESSED"
)

var OrderStatuses = [...]string{NEW, PROCESSING, INVALID, PROCESSED}

var OrderStatusMap = map[string]int{
	NEW:        0,
	PROCESSING: 1,
	INVALID:    2,
	PROCESSED:  3,
}

func getStatusAsInt(status string) (int, error) {
	val, ok := OrderStatusMap[status]
	if !ok {
		return -1, fmt.Errorf("unknown status: %s", status)
	}
	return val, nil
}

func getStatusAsString(status int) (string, error) {
	if status < 0 || status >= len(OrderStatuses) {
		return "", fmt.Errorf("invalid status code: %d", status)
	}
	return OrderStatuses[status], nil
}

type (
	Order struct {
		Number     string    `json:"number"`
		Status     string    `json:"status"`
		Accrual    float64   `json:"accrual"`
		UploadedAt time.Time `json:"uploaded_at"`
	}
	ErrAlreadyExists struct {
		logger.TracedError
	}
	ErrAddedByAnotherUser struct {
		logger.TracedError
	}
)

func (o *Order) IsValid() bool {
	// Удаляем пробелы, если есть
	cleaned := make([]byte, 0, len(o.Number))
	for i := 0; i < len(o.Number); i++ {
		if o.Number[i] != ' ' {
			cleaned = append(cleaned, o.Number[i])
		}
	}

	// Проверяем отсутствие символов, помимо цифр
	for i := 0; i < len(cleaned); i++ {
		if cleaned[i] < '0' || cleaned[i] > '9' {
			return false
		}
	}

	// Проверяем номер алгоритмом Луна
	sum := 0
	double := false

	for i := len(cleaned) - 1; i >= 0; i-- {
		digit := int(cleaned[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	if sum%10 == 0 {
		//Если номер валидный, то перепишем значение на очищенное от лишних символов
		o.Number = string(cleaned)
		return true
	}

	return false
}

func (o *Order) AddToUser(ctx context.Context, u *User) error {
	storage, ok := repository.FromContext(ctx)
	if !ok {
		return logger.NewTracedError("storage not found in context", nil)
	}

	username, err := storage.AddNewOrder(ctx, o.Number, u.auth.Login)
	if err != nil {
		return logger.NewTracedError("error getting user data: ", err)
	}

	if username == "" {
		return nil
	}

	if username != u.auth.Login {
		return &ErrAddedByAnotherUser{
			*logger.NewTracedError("order already added by another user",
				fmt.Errorf("")),
		}
	}

	return &ErrAlreadyExists{
		*logger.NewTracedError("order already exists", fmt.Errorf("")),
	}
}

func (o *Order) SaveToDb(ctx context.Context) error {
	storage, ok := repository.FromContext(ctx)
	if !ok {
		return logger.NewTracedError("storage not found in context", nil)
	}

	stt, err := getStatusAsInt(o.Status)
	if err != nil {
		return logger.NewTracedError("bad status value", err)
	}
	err = storage.UpdOrderData(ctx, o.Number, stt, o.Accrual)

	if err != nil {
		return logger.NewTracedError("error getting orders: ", err)
	}

	return nil
}

func GetOrdersForUser(ctx context.Context, userNm string) ([]Order, error) {
	storage, ok := repository.FromContext(ctx)
	if !ok {
		return nil, logger.NewTracedError("storage not found in context", nil)
	}

	data, err := storage.GetOrdesList(ctx, userNm, len(OrderStatusMap), 0)

	if err != nil {
		return nil, logger.NewTracedError("error getting orders: ", err)
	}

	res := make([]Order, len(data))

	for i := 0; i < len(data); i++ {
		stt, err := getStatusAsString(data[i].Status)
		if err == nil {
			order := Order{
				data[i].Number,
				stt,
				data[i].Accrual,
				data[i].Uploaded,
			}
			res = append(res, order)
		}
	}

	return res, nil
}

func GetOrdersToProcess(ctx context.Context, count int) ([]Order, error) {
	storage, ok := repository.FromContext(ctx)
	if !ok {
		return nil, logger.NewTracedError("storage not found in context", nil)
	}

	data, err := storage.GetOrdesList(ctx, "", OrderStatusMap[PROCESSING], count)

	if err != nil {
		return nil, logger.NewTracedError("error getting orders: ", err)
	}

	res := make([]Order, 0, len(data))

	for i := 0; i < len(data); i++ {

		stt, err := getStatusAsString(data[i].Status)
		if err == nil {
			order := Order{
				data[i].Number,
				stt,
				data[i].Accrual,
				data[i].Uploaded,
			}
			res = append(res, order)
		}
	}

	return res, nil
}
