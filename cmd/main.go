package main

import (
	"context"
	"fmt"
	"github.com/Netflix/go-env"
	"log/slog"
	"os"
	"os/signal"
	"robots/internal/robot"
)

func main() {
	var config robot.Config
	if _, err := env.UnmarshalFromEnviron(&config); err != nil {
		panic(err)
	}
	log := NewLogger(config.LogLevel)
	if err := validateEnvVariables(config); err != nil {
		log.Info(err.Error())
		panic(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt) // Handle CTRL+C
	defer stop()
	secretManager := robot.SecretManager{Config: config, Log: log}
	secret := secretManager.SplitSecret(config.Secret)
	robots := secretManager.CreateRobots(secret)
	secretManager.FindSecret(ctx, robots)
}

// NewLogger Build a logger
// Fallback as INFO by default
func NewLogger(logLevel string) *slog.Logger {
	var level slog.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

func validateEnvVariables(config robot.Config) error {
	if config.NbrOfRobots < 2 {
		return fmt.Errorf("number of robots should be at least 2")
	}
	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer size should be positive : %d", config.BufferSize)
	}
	if config.PercentageOfLost < 0 {
		return fmt.Errorf("percentage of lost should be positive : %d", config.PercentageOfLost)
	}
	if config.PercentageOfDuplicated < 0 {
		return fmt.Errorf("percentage of lost should be positive : %d", config.PercentageOfDuplicated)
	}
	if config.DuplicatedNumber < 0 {
		return fmt.Errorf("duplicated number should be positive : %d", config.DuplicatedNumber)
	}
	if config.MaxAttempts <= 0 {
		return fmt.Errorf("max attempts should be positive : %d", config.MaxAttempts)
	}
	return nil
}
