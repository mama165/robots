package main

import (
	"context"
	"os/signal"
	"robots/internal/conf"
	rb "robots/internal/robot"
	sp "robots/internal/supervisor"
	"robots/pkg/errors"
	"robots/pkg/events"
	"robots/pkg/workers"
	"sync"
	"syscall"

	"github.com/Netflix/go-env"
	"github.com/mama165/sdk-go/logs"
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

	baseCtx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(baseCtx, config.Timeout)
	ctx, stop := signal.NotifyContext(timeoutCtx, syscall.SIGINT) // Handle CTRL+C
	defer cancel()
	defer stop()
	secretManager := rb.SecretManager{Config: config}
	secret := secretManager.SplitSecret(config.Secret)
	robots := secretManager.CreateRobots(secret)
	winner := make(chan rb.Robot)
	// ⚠️ Buffer will receive a lot of events
	// ⚠️ Message can be lost
	event := make(chan events.Event, config.BufferSize)
	waitGroup := sync.WaitGroup{}
	supervisor := sp.NewSupervisor(ctx, cancel, &waitGroup, log)
	counter := events.NewCounter()

	// Only few workers run for each robot
	for _, robot := range robots {
		supervisor.Add(
			workers.NewProcessSummaryWorker(log, robot, robots, event).WithName("summary worker"),
			workers.NewMergeSecretWorker(log, robot, event).WithName("update worker"),
			workers.NewConvergenceDetectorWorker(config, log, robot, winner).WithName("convergence detector worker"),
			workers.NewStartGossipWorker(config, log, robot, robots, event).WithName("start gossip worker"),
			workers.NewQuiescenceDetectorWorker(config, log, robot, event).WithName("quiescence worker"),
		)
	}
	// One worker is responsible for writing the secret
	// One worker to handle the events
	supervisor.Add(
		workers.NewMetricWorker(config, log, event).WithName("channel capacity worker"),
		workers.NewDispatcher(log, event).Add(
			events.NewInvariantViolationProcessor(log, counter),
			events.NewMessageDuplicatedProcessor(log),
			events.NewMessageReceivedProcessor(log, counter),
			events.NewMessageReorderedProcessor(log, counter),
			events.NewMessageSentProcessor(log, counter),
			events.NewWorkerRestartedAfterPanicProcessor(log, counter),
			events.NewChannelCapacityProcessor(log, config.LowCapacityThreshold),
			events.NewQuiescenceDetectorProcessor(log),
		).WithName("metric worker"),
	)
	supervisor.Run()

	// Wait for the context cancellation (timeout or CTRL+C)
	<-ctx.Done()
	log.Info("Stopping supervisor...")
	supervisor.Stop()
}

// TODO Ajouter les validations restantes
func validateEnvVariables(config conf.Config) error {
	if config.NbrOfRobots < 2 {
		return errors.ErrNumberOfRobots
	}
	if config.BufferSize <= 0 {
		return errors.ErrNegativeBufferSize
	}
	if config.PercentageOfLost < 0 {
		return errors.ErrNegativePercentageOfLost
	}
	if config.PercentageOfDuplicated < 0 {
		return errors.ErrNegativePercentageOfDuplicated
	}
	if config.DuplicatedNumber < 0 {
		return errors.ErrNegativeDuplicatedNumber
	}
	if config.MaxAttempts <= 0 {
		return errors.ErrNegativeMaxAttempts
	}
	if config.MetricInterval <= 0 {
		return errors.ErrNegativeMetricInterval
	}
	return nil
}
