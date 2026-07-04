package main

import (
	"net/http"

	"github.com/e-l-l-a-r/gophermart/internal/compressor"
	"github.com/e-l-l-a-r/gophermart/internal/config"
	"github.com/e-l-l-a-r/gophermart/internal/handler"
	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/e-l-l-a-r/gophermart/internal/repository"
)

func main() {
	if err := run(); err != nil {
		logger.Fatal(err)
	}
}

func run() error {
	conf := config.GetConfig()
	log, err := logger.InitLogger(conf.LogLevel)

	if err != nil {
		return err
	}
	conf.Print()

	storage, err := repository.InitSqlStorage(conf.DbConnString)
	if err != nil {
		log.Error("Can't connect to database " + err.Error())
		return err
	}

	err = storage.DoMigrate()
	if err != nil {
		log.Error("Can't migrate database " + err.Error())
		return err
	}

	router := handler.GetRouter(storage)
	err = http.ListenAndServe(conf.Address, compressor.GzipHandle(log.LogHandle(router)))
	if err != nil {
		return err
	}
	return nil
}
