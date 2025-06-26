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
	AddrServer string `json:"addr_server,omitempty"`
	LogLevel   string `json:"log_level,omitempty"`
	SecretKey  string `json:"key,omitempty"`
	CryptoKey  string `json:"crypto_key,omitempty"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// GetConfig конструктор конфига
func GetConfig() (*Config, error) {
	cfg := &Config{}

	parseFlags(cfg)
	parseConfig(cfg)

	if cfg.AddrServer == "" {
		cfg.AddrServer = "localhost:8080"
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	_, _, err := net.SplitHostPort(cfg.AddrServer)
	if err != nil {
		return nil, fmt.Errorf("неверный адрес сервера: %w", err)
	}

	if cfg.Username == "" || cfg.Password == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("не введен пользователь, пароль или секретный ключ")
	}

	return cfg, nil
}

// parseFlags Чтение переданных флагов
func parseFlags(cfg *Config) {
	flag.StringVar(&cfg.AddrServer, "a", "", "server and port to run server")
	flag.StringVar(&cfg.LogLevel, "l", "", "log level")
	flag.StringVar(&cfg.SecretKey, "k", "", "secret key for encryption")
	flag.StringVar(&cfg.CryptoKey, "cert", "", "path to trusted CA certificate")
	flag.StringVar(&cfg.Username, "user", "", "username")
	flag.StringVar(&cfg.Password, "pass", "", "password")
	flag.Parse()
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
	if flag.Lookup("l").Value.String() == "" {
		cfg.LogLevel = tmpCfg.LogLevel
	}
	if flag.Lookup("k").Value.String() == "" {
		cfg.SecretKey = tmpCfg.SecretKey
	}
	if flag.Lookup("crypto-key").Value.String() == "" {
		cfg.CryptoKey = tmpCfg.CryptoKey
	}
	if flag.Lookup("user").Value.String() == "" {
		cfg.Username = tmpCfg.Username
	}
	if flag.Lookup("password").Value.String() == "" {
		cfg.Password = tmpCfg.Password
	}
}
