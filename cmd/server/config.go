package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	"log/slog"
)

type Config struct {
	AddrServer    string `json:"addr_server,omitempty"`
	AddrDatabase  string `json:"database_dsn,omitempty"`
	LogLevel      string `json:"log_level,omitempty"`
	SecretKey     string `json:"key,omitempty"`
	CryptoKey     string `json:"crypto_key,omitempty"`
	TrustedSubnet string `json:"trusted_subnet,omitempty"`
}

// GetConfig() получить конфиг сервера
func GetConfig() (*Config, error) {
	cfg := &Config{}

	parseFlags(cfg)
	parseConfig(cfg)
	parseEnv(cfg)

	if cfg.AddrServer == "" {
		cfg.AddrServer = "localhost:8080"
	}

	if cfg.AddrDatabase == "" {
		cfg.AddrDatabase = "host=localhost user=echo9et password=123321 dbname=echo9et sslmode=disable"
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	_, _, err := net.SplitHostPort(cfg.AddrServer)
	if err != nil {
		return nil, fmt.Errorf("неверный адрес сервера: %w", err)
	}

	return cfg, nil
}

// parseFlags Чтение переданных флагов
func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.AddrServer, "a", "", "server and port to run server")
	flag.StringVar(&cfg.AddrDatabase, "d", "", "address to postgres base")
	flag.StringVar(&cfg.LogLevel, "l", "", "log level")
	flag.StringVar(&cfg.SecretKey, "k", "", "secret key for encryption")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "privat key")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "sunbnet clients")
	flag.Parse()
}

// parseEnv Чтение перемменных окружения
func parseEnv(cfg *Config) {
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.AddrServer = envRunAddr
	}

	if envDatabaseAddr := os.Getenv("DATABASE_DSN"); envDatabaseAddr != "" {
		cfg.AddrDatabase = envDatabaseAddr
	}

	if envRunLogLVL := os.Getenv("LOG_LVL"); envRunLogLVL != "" {
		cfg.LogLevel = envRunLogLVL
	}

	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		cfg.SecretKey = envSecretKey
	}

	if envCryptoKey := os.Getenv("CRYPTO-KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
	}
}

// parseConfig Чтение JSON-конфига
func parseConfig(cfg *Config) {
	fileData, err := os.ReadFile("config")
	if err != nil {
		slog.Warn("Не удалось открыть конфигурационный файл", "error", err)
		return
	}

	tmpCfg := Config{}
	err = json.Unmarshal(fileData, &tmpCfg)
	if err != nil {
		slog.Warn("Ошибка при разборе JSON конфига", "error", err)
		return
	}

	if flag.Lookup("a").Value.String() == "" {
		cfg.AddrServer = tmpCfg.AddrServer
	}
	if flag.Lookup("d").Value.String() == "" {
		cfg.AddrDatabase = tmpCfg.AddrDatabase
	}
	if flag.Lookup("l").Value.String() == "" {
		cfg.LogLevel = tmpCfg.LogLevel
	}
	if flag.Lookup("k").Value.String() == "" {
		cfg.SecretKey = tmpCfg.SecretKey
	}
	if flag.Lookup("crypto-key").Value.String() == "" {
		cfg.CryptoKey = tmpCfg.CryptoKey
	}
	if flag.Lookup("trusted_subnet").Value.String() == "" {
		cfg.TrustedSubnet = tmpCfg.TrustedSubnet
	}
}
