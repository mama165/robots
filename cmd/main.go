package main

import (
	"context"
	"fmt"
	"github.com/Netflix/go-env"
	"github.com/mama165/sdk-go/logs"
	"os/signal"
	"robots/internal/conf"
	rb "robots/internal/robot"
	sp "robots/internal/supervisor"
	"robots/pkg/workers"
	"sync"
	"syscall"
)

func main() {
	var config conf.Config
	if _, err := env.UnmarshalFromEnviron(&config); err != nil {
		panic(err)
	}
	log := logs.GetLoggerFromString(config.LogLevel)
	if err := validateEnvVariables(config); err != nil {
		log.Error(err.Error())
		panic(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT) // Handle CTRL+C
	defer stop()
	secretManager := rb.SecretManager{Config: config, Log: log}
	secret := secretManager.SplitSecret(config.Secret)
	robots := secretManager.CreateRobots(secret)
	winner := make(chan rb.Robot)
	waitGroup := sync.WaitGroup{}
	supervisor := sp.NewSupervisor(ctx, cancel, &waitGroup, log)

	// Running two goroutines for each robot to start
	for _, robot := range robots {
		supervisor.
			Add(workers.NewProcessSummaryWorker(config, log, robot, robots).WithName("summary worker")).
			Add(workers.NewUpdateWorker(config, log, robot).WithName("update worker")).
			Add(workers.NewSuperviseRobotWorker(config, log, robot, winner).WithName("supervise robot worker")).
			Add(workers.NewStartGossipWorker(config, log, robot, robots).WithName("start gossip worker"))
	}
	// Only the winner goroutine handle the writing
	supervisor.Add(workers.NewWriteSecretWorker(config, log, winner).WithName("write secret worker"))
	supervisor.Run()

	// Wait for the context cancellation (timeout or CTRL+C)
	<-ctx.Done()
	log.Info("Stopping supervisor...")
	supervisor.Stop()
}

func validateEnvVariables(config conf.Config) error {
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
