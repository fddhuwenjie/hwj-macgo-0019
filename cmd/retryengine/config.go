package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr      string
	DataDir         string
	BudgetWindow    time.Duration
	ClockSkew       time.Duration
	MaxConcurrent   int
	ShutdownTimeout time.Duration
}

func LoadConfig() (Config, error) {
	cfg := Config{
		ListenAddr:      "127.0.0.1:0",
		DataDir:         "./data",
		BudgetWindow:    24 * time.Hour,
		ClockSkew:       100 * time.Millisecond,
		MaxConcurrent:   16,
		ShutdownTimeout: 5 * time.Second,
	}

	if v := os.Getenv("HWJ_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("HWJ_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv("HWJ_BUDGET_WINDOW"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("parse HWJ_BUDGET_WINDOW: %w", err)
		}
		cfg.BudgetWindow = d
	}
	if v := os.Getenv("HWJ_CLOCK_SKEW"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("parse HWJ_CLOCK_SKEW: %w", err)
		}
		cfg.ClockSkew = d
	}
	if v := os.Getenv("HWJ_MAX_CONCURRENT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("parse HWJ_MAX_CONCURRENT: %w", err)
		}
		cfg.MaxConcurrent = n
	}
	if v := os.Getenv("HWJ_SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("parse HWJ_SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = d
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.BudgetWindow <= 0 {
		return fmt.Errorf("budget window must be positive")
	}
	if c.ClockSkew < 0 {
		return fmt.Errorf("clock skew must not be negative")
	}
	if c.MaxConcurrent <= 0 {
		return fmt.Errorf("max concurrent must be positive")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}
	if c.DataDir == "" {
		return fmt.Errorf("data dir must not be empty")
	}
	return nil
}
