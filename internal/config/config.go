package config

import (
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/e-l-l-a-r/gophermart/internal/logger"
	"github.com/spf13/pflag"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	LogLevel       string `env:"LOG_LEVEL"`
	DbConnString   string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func (conf *Config) Print() {
	log, err := logger.GetLogger()

	if err == nil {
		log.InfoMsg("==================================")
		log.InfoMsg("Running server on ", conf.Address)
		log.InfoMsg("LogLevel ", conf.LogLevel)
		log.InfoMsg("DB Connegtion dtring: ", conf.DbConnString)
		log.InfoMsg("Address of accrual: ", conf.AccrualAddress)
		log.InfoMsg("==================================")
	}
}

func parseFlags() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Usage = func() {
		logger.Info(pflag.CommandLine.Output(), "GopherMart server\nUsage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
	}

	pflag.Parse()
}

func GetConfig() (result Config) {
	var flagRunAddr = pflag.StringP("address", "a", "localhost:8080",
		"address and port to run server")
	var flagLogLevel = pflag.StringP("log-level", "l", "Info",
		"log level, may be Debug, Info (default), Warning, Error")
	var flagDbConnSrting = pflag.StringP("db-conn-string", "d", "",
		"database connection string")
	var flagAccuralAddr = pflag.StringP("key", "k", "",
		"address and port to connect to accural server")

	err := env.Parse(&result)

	if err != nil {
		logger.Warn(err)
	}

	parseFlags()

	if result.Address == "" {
		result.Address = *flagRunAddr
	}
	if result.LogLevel == "" {
		result.LogLevel = *flagLogLevel
	}
	if result.DbConnString == "" {
		result.DbConnString = *flagDbConnSrting
	}
	if result.AccrualAddress == "" {
		result.AccrualAddress = *flagAccuralAddr
	}

	return
}
