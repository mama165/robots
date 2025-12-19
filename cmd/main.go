package main

import (
	"context"
	"os"
	"os/signal"
	"robots/internal/conf"
	"robots/pkg/errors"
	"robots/pkg/events"
	"robots/pkg/robot"
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
	domainEvent := make(chan events.Event, config.BufferSize)
	telemetryEvent := make(chan events.Event, config.BufferSize)
	secretManager := robot.SecretManager{Config: config}
	secret := secretManager.SplitSecret(config.Secret)
	robots := secretManager.CreateRobots(secret)
	winner := make(chan *robot.Robot, 1)
	// ⚠️ Buffer will receive a lot of events
	// ⚠️ Message can be lost
	waitGroup := sync.WaitGroup{}
	supervisor := workers.NewSupervisor(ctx, cancel, &waitGroup, log)
	counter := events.NewCounter()
	once := &sync.Once{}

	file, err := os.Create(config.OutputFile)
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}
	defer file.Close()

	// Only few workers run for each robot
	for _, r := range robots {
		supervisor.Add(
			workers.NewProcessSummaryWorker(log, r, robots, domainEvent).WithName("summary worker"),
			workers.NewMergeSecretWorker(log, r, domainEvent).WithName("update worker"),
			workers.NewConvergenceDetectorWorker(config, log, r, winner, once, file).WithName("convergence detector worker"),
			workers.NewStartGossipWorker(config, log, r, robots, domainEvent).WithName("start gossip worker"),
			workers.NewQuiescenceDetectorWorker(config, log, r, domainEvent, 0).WithName("quiescence worker"),
		)
	}
	// One worker is responsible for writing the secret
	// One worker to handle the events
	supervisor.Add(
		workers.NewChannelCapacityWorker(config, log, domainEvent).WithName("channel capacity worker"),
		workers.NewSnapshotWorker(config, log, telemetryEvent).WithName("snapshot worker"),
		workers.NewEventFanout(log, domainEvent, telemetryEvent).Add(
			events.NewInvariantViolationHandler(log, counter),
			events.NewMessageDuplicatedHandler(log, counter),
			events.NewMessageReceivedHandler(log, counter),
			events.NewMessageReorderedHandler(log, counter),
			events.NewMessageSentHandler(log, counter),
			events.NewWorkerRestartedAfterPanicHandler(log, counter),
			events.NewChannelCapacityHandler(log, config.LowCapacityThreshold),
			events.NewQuiescenceDetectorHandler(log),
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
