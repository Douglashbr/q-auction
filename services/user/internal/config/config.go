// Package config lê a configuração do serviço a partir de variáveis de ambiente.
package config

import "os"

type Config struct {
	Port     string
	Service  string
	Instance string // hostname do container; identifica a réplica atrás do LB
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	instance, err := os.Hostname()
	if err != nil {
		instance = "unknown"
	}

	return Config{Port: port, Service: "user", Instance: instance}
}
