package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
)

var retryClient *retryablehttp.Client

var delay atomic.Int64

func getHttpClient() *retryablehttp.Client {
	if retryClient == nil {
		retryClient = retryablehttp.NewClient()
		retryClient.RetryMax = 3
	}
	return retryClient
}
func GetOrderStatus(ctx context.Context,
	baseUrl string,
	ch chan model.Order,
	ch_out chan model.Order,
	wg *sync.WaitGroup) {

	log, err := logger.GetLogger()
	if err != nil {
		logger.Warn(err)
	}

	client := getHttpClient()

	defer wg.Done()

	log.Info("Starting orders sync thread")
	for {
		//Проверяем, не выставлен ли таймаут
		if delay.Load() > 0 {
			timer := time.NewTimer(time.Duration(delay.Load()) * time.Second)
			select {
			case <-timer.C:
				delay.Store(0)
				continue
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return
			}
		}

		order, ok := <-ch
		if !ok {
			return
		}

		url := fmt.Sprintf("%s/api/orders/%s", baseUrl, order.Number)

		request, err := retryablehttp.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.WarnMsg(err)
		}
		request.Header.Set("Content-Type", "application/json")

		resp, err := log.DoRequestWithLog(client, request)

		if err != nil {
			log.WarnMsg(err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			dec := json.NewDecoder(resp.Body)
			oldStt := order.Status
			err := dec.Decode(&order)
			if err == nil {
				log.Info(fmt.Sprintf("Order processed: %s\nStatus: %s\nAccrual: %f",
					order.Number, order.Status, order.Accrual))
				// Обновляем только в случае изменения статуса
				if oldStt != order.Status {
					ch_out <- order
				}
			} else {
				logger.Debug("Error while processing order: ", err)
			}
		} else if resp.StatusCode == http.StatusNoContent {
			if order.UploadedAt.Add(time.Hour * 24).Before(time.Now()) {
				log.Warn(fmt.Sprintf("24 hours pass, but order %s is not still added", order.Number))
				order.Status = model.INVALID
				ch_out <- order
			}
		} else if resp.StatusCode == http.StatusTooManyRequests {
			delayVal, err := strconv.Atoi(resp.Header.Get("Retry-After"))
			if err != nil {
				delayVal = 60
			}
			delay.Store(int64(delayVal))
			log.WarnMsg(fmt.Sprintf("Too many requests, wait for %d seconds", delayVal))
		}
	}
}
