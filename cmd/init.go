package main

import "github.com/kdv2001/onlySubscription/pkg/config"

type configValues struct {
	TelegramToken string `env:"TELEGRAM_TOKEN" json:"telegram_token"`
	PostgresDSN   string `env:"DATABASE_DSN" json:"database_dsn"`
}

const configPath = "./deploy/values.json"

func initFlags() (*configValues, error) {
	v := &configValues{}
	err := config.UnmarshalJSONFile(v, configPath)
	if err != nil {
		return nil, err
	}

	return v, nil
}
