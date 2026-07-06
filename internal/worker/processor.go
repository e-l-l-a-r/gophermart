package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
)

func GetOrderStatus(baseUrl string, ch chan model.Order, ch_out chan model.Order) {

	log, err := logger.GetLogger()
	if err != nil {
		logger.Warn(err)
	}

	client := http.Client{
		Timeout: time.Second * 1, // интервал ожидания: 1 секунда
	}

	defer close(ch_out)

	log.Info("Starting orders sync thread")
	for {
		order, ok := <-ch
		if !ok {
			return
		}

		data, err := json.Marshal(order)
		if err != nil {
			continue
		}

		url := fmt.Sprintf("http://%s/api/orders/%s", baseUrl, order.Number)

		request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			log.WarnMsg(err)
		}
		request.Header.Set("Content-Type", "application/json")

		result, err := logger.ExecuteWithRetry(func(args ...interface{}) (interface{}, error) {
			return log.DoRequestWithLog(&client, request)
		})

		if err != nil {
			log.WarnMsg(err)
			continue
		}

		if result != nil {
			resp := result.(*http.Response)
			if resp.StatusCode == http.StatusOK {
				dec := json.NewDecoder(resp.Body)
				oldStt := order.Status
				if err := dec.Decode(&order); err != nil {
					log.Info(fmt.Sprintf("Order processed: %s\nStatus: %s\nAccrual: %f",
						order.Status, order.Number, order.Accrual))
					// Обновляем только в случае изменения статуса
					if oldStt != order.Status {
						ch_out <- order
					}
				}
			} else if resp.StatusCode == http.StatusNoContent {
				order.Status = model.INVALID
				ch_out <- order
			} else if resp.StatusCode == http.StatusTooManyRequests {
				delay, err := strconv.Atoi(resp.Header.Get("Retry-After"))
				if err != nil {
					delay = 60
				}
				log.WarnMsg(fmt.Sprintf("Too many requests, wait for %d seconds", delay))
				time.Sleep(time.Duration(delay) * time.Second)
			} else {
				log.Warn("Order processing error: " + order.Number)
			}
		}

	}
}
