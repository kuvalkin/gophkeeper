package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address               string `env:"ADDRESS"`
	DatabaseDSN           string `env:"DATABASE_DSN"`
	DatabaseTimeout       time.Duration
	TokenExpirationPeriod time.Duration
}

func Resolve() (*Config, error) {
	conf := &Config{
		Address:     "localhost:8080",
		DatabaseDSN: "",
		// hardcoded for now
		DatabaseTimeout:       5 * time.Second,
		TokenExpirationPeriod: time.Hour * 24 * 30,
	}

	err := parseEnv(conf)
	if err != nil {
		return nil, fmt.Errorf("error parsing env: %w", err)
	}

	err = checkRequiredFields(conf)
	if err != nil {
		return nil, fmt.Errorf("required fields: %w", err)
	}

	return conf, nil
}

func parseEnv(conf *Config) error {
	if err := env.Parse(conf); err != nil {
		return fmt.Errorf("parse environment variables: %w", err)
	}

	err := validateServerAddress(conf.Address)
	if err != nil {
		return fmt.Errorf("invalid server address: %w", err)
	}

	return nil
}

func checkRequiredFields(conf *Config) error {
	if conf.Address == "" {
		return errors.New("run address is required")
	}

	if conf.DatabaseDSN == "" {
		return errors.New("database DSN is required")
	}

	return nil
}

func validateServerAddress(address string) error {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return errors.New("need address in a form host:port")
	}

	if err := validateIP(parts[0]); err != nil {
		return fmt.Errorf("invalid ip address: %w", err)
	}

	if err := validatePort(parts[1]); err != nil {
		return fmt.Errorf("invalid port in address: %w", err)
	}

	return nil
}

func validateIP(ipString string) error {
	if ipString == "localhost" {
		return nil
	}

	ip := net.ParseIP(ipString)
	if ip == nil {
		return fmt.Errorf("could not parse ip: %v", ipString)
	}

	return nil
}

func validatePort(portString string) error {
	port, err := strconv.Atoi(portString)
	if err != nil {
		return fmt.Errorf("could not parse port: %w", err)
	}

	if port < 0 || port > 65535 {
		return fmt.Errorf("port out of range: %d", port)
	}

	return nil
}
