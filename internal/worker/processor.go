package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/model"
)

var retryClient *retryablehttp.Client

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
		order, ok := <-ch
		if !ok {
			return
		}

		data, err := json.Marshal(order)
		if err != nil {
			continue
		}

		url := fmt.Sprintf("%s/api/orders/%s", baseUrl, order.Number)

		request, err := retryablehttp.NewRequest(http.MethodGet, url, bytes.NewReader(data))
		if err != nil {
			log.WarnMsg(err)
		}
		request.Header.Set("Content-Type", "application/json")

		result, err := logger.ExecuteWithRetry(func(args ...interface{}) (interface{}, error) {
			return log.DoRequestWithLog(client, request)
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
				err := dec.Decode(&order)
				if err == nil {
					log.Info(fmt.Sprintf("Order processed: %s\nStatus: %s\nAccrual: %f",
						order.Number, order.Status, order.Accrual))
					// Обновляем только в случае изменения статуса
					if oldStt != order.Status {
						ch_out <- order
					}
				} else {
					log.WarnMsg("Error while processing order: ", err)
				}
			} else if resp.StatusCode == http.StatusNoContent {
				if order.UploadedAt.Add(time.Hour * 24).Before(time.Now()) {
					log.Warn(fmt.Sprintf("24 hours pass, but order %s is not still added", order.Number))
					order.Status = model.INVALID
					ch_out <- order
				}
			} else if resp.StatusCode == http.StatusTooManyRequests {
				delay, err := strconv.Atoi(resp.Header.Get("Retry-After"))
				if err != nil {
					delay = 60
				}
				log.WarnMsg(fmt.Sprintf("Too many requests, wait for %d seconds", delay))
				timer := time.NewTimer(time.Duration(delay) * time.Second)
				select {
				case <-timer.C:
					continue
				case <-ctx.Done():
					if !timer.Stop() {
						<-timer.C
					}
					return
				}
			} else {
				log.Warn("Order processing error: " + order.Number)
			}
		}

	}
}
