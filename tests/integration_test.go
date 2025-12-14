package tests

import (
	"context"
	"log/slog"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/pkg/events"
	"robots/pkg/workers"
	"testing"
	"time"

	"github.com/mama165/sdk-go/logs"
	"github.com/stretchr/testify/assert"
)

func TestWorkerGossip_Robustness(t *testing.T) {
	ass := assert.New(t)

	logLevel := slog.LevelDebug
	logger := logs.GetLoggerFromLevel(logLevel)

	cfg := conf.Config{
		NbrOfRobots:            6,
		Secret:                 "Hidden beneath the old oak tree, golden coins patiently await discovery.",
		OutputFile:             "",
		BufferSize:             10,
		EndOfSecret:            ".",
		PercentageOfLost:       50, // simulate loss
		PercentageOfDuplicated: 50, // simulate duplication
		DuplicatedNumber:       2,
		MaxAttempts:            5,
		Timeout:                0,
		QuietPeriod:            0,
		GossipTime:             100 * time.Millisecond,
		LogLevel:               logLevel.String(),
	}

	words := []string{"hello", "world."}
	sm := robot.SecretManager{Config: cfg}
	robots := sm.CreateRobots(words)
	event := make(chan events.Event, 100)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// When workers are started manually
	for _, r := range robots {
		r := r // Capture local variable for goroutine
		go func() {
			err := workers.NewMergeSecretWorker(logger, r, event).Run(ctx)
			ass.NoError(err)
		}()
		go func() {
			err := workers.NewProcessSummaryWorker(logger, r, robots, event).Run(ctx)
			ass.NoError(err)
		}()
		go func() {
			err := workers.NewStartGossipWorker(cfg, logger, r, robots, event).Run(ctx)
			ass.NoError(err)
		}()
	}

	timeout := 2 * time.Second
	interval := 50 * time.Millisecond

	for _, r := range robots {
		ass.Eventually(func() bool {
			return r.IsSecretCompleted(cfg.EndOfSecret)
		}, timeout, interval, "Robot %d should have completed the secret", r.ID)
	}
}
