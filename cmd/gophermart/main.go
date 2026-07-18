package main

import (
	"net/http"
	"sync"

	"github.com/e-l-l-a-r/gophermart/internal/compressor"
	"github.com/e-l-l-a-r/gophermart/internal/config"
	"github.com/e-l-l-a-r/gophermart/internal/handler"
	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
	"github.com/e-l-l-a-r/gophermart/internal/worker"
)

func main() {
	conf := config.GetConfig()
	_, err := logger.InitLogger(conf.LogLevel)

	if err != nil {
		logger.Fatal(err)
	}

	conf.Print()

	storage, err := repository.InitSqlStorage(conf.DbConnString)

	if err != nil {
		logger.Fatal("Can't connect to database " + err.Error())
	}

	err = storage.DoMigrate()
	if err != nil {
		logger.Fatal("Can't migrate database " + err.Error())
	}

	doneCh, wg, err := runWorker(conf, storage)

	defer close(doneCh)

	if err != nil {
		logger.Fatal(err)
	}
	if err := runHttpServer(conf, storage); err != nil {
		logger.Fatal(err)
	}
	wg.Wait()

}

func runHttpServer(conf config.Config, storage repository.Storage) error {
	log, _ := logger.GetLogger()

	router := handler.GetRouter(storage)
	err := http.ListenAndServe(conf.Address, compressor.GzipHandle(log.LogHandle(router)))
	if err != nil {
		return err
	}

	return nil
}

func runWorker(conf config.Config, storage repository.Storage) (doneCh chan struct{}, wg *sync.WaitGroup, err error) {

	doneCh = make(chan struct{})

	wg = worker.DoWork(storage, doneCh, conf.AccrualAddress)

	return

}
