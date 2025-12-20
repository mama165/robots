package tests

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"robots/internal/conf"
	"robots/pkg/events"
	"robots/pkg/robot"
	"robots/pkg/workers"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEnd_WritesOutputFile vérifie que le secret est écrit dans le fichier à la fin du processus.
func TestEndToEnd_WritesOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "result.txt")
	secret := "hello world."

	cfg := conf.Config{
		NbrOfRobots:            5,
		Secret:                 secret,
		OutputFile:             outputFile,
		BufferSize:             50,
		EndOfSecret:            ".",
		PercentageOfLost:       0,
		PercentageOfDuplicated: 0,
		DuplicatedNumber:       1,
		MaxAttempts:            5,
		GossipTime:             50 * time.Millisecond,
		QuietPeriod:            200 * time.Millisecond,
		Timeout:                3 * time.Second,
		MetricInterval:         100 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	sm := robot.SecretManager{Config: cfg}
	robots := sm.CreateRobots(strings.Fields(cfg.Secret))
	eventsCh := make(chan events.Event, 100)

	// Start workers
	for _, r := range robots {
		go workers.NewMergeSecretWorker(slog.Default(), r, eventsCh).Run(ctx)
		go workers.NewProcessSummaryWorker(slog.Default(), r, robots, eventsCh).Run(ctx)
		go workers.NewStartGossipWorker(cfg, slog.Default(), r, robots, eventsCh).Run(ctx)
		go workers.NewConvergenceDetectorWorker(cfg, slog.Default(), r, eventsCh).Run(ctx)
	}

	// Attend que le fichier soit créé
	require.Eventually(t, func() bool {
		_, err := os.Stat(outputFile)
		return err == nil
	}, 2*time.Second, 50*time.Millisecond, "Le fichier de sortie devrait être écrit")

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	require.Equal(t, secret, string(content))
}

// TestDoesNotWriteFileBeforeQuiescence vérifie que le fichier n'est pas écrit si le timeout est avant la quiescence
func TestDoesNotWriteFileBeforeQuiescence(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "result.txt")

	cfg := conf.Config{
		NbrOfRobots:            5,
		Secret:                 "hello world.",
		OutputFile:             outputFile,
		EndOfSecret:            ".",
		BufferSize:             50,
		PercentageOfLost:       0,
		PercentageOfDuplicated: 0,
		GossipTime:             20 * time.Millisecond,
		QuietPeriod:            5 * time.Second,        // volontairement long
		Timeout:                500 * time.Millisecond, // plus court que QuietPeriod
		MetricInterval:         100 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	sm := robot.SecretManager{Config: cfg}
	robots := sm.CreateRobots(strings.Fields(cfg.Secret))

	eventsCh := make(chan events.Event, 100)
	logger := slog.New(slog.NewTextHandler(nil, nil))

	// Start workers
	for _, r := range robots {
		go workers.NewMergeSecretWorker(logger, r, eventsCh).Run(ctx)
		go workers.NewProcessSummaryWorker(logger, r, robots, eventsCh).Run(ctx)
		go workers.NewStartGossipWorker(cfg, logger, r, robots, eventsCh).Run(ctx)
		go workers.NewConvergenceDetectorWorker(cfg, logger, r, eventsCh).Run(ctx)
	}

	<-ctx.Done()

	_, err := os.Stat(outputFile)
	assert.True(t, errors.Is(err, fs.ErrNotExist), "Le fichier ne devrait pas être écrit avant la quiescence")
}

// TestWinnerChannelBlocksFileWrite vérifie que la coordination des robots empêche l'écriture multiple
func TestWinnerChannelBlocksFileWrite(t *testing.T) {
	ass := assert.New(t)

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "secret.txt")

	cfg := conf.Config{
		OutputFile:  outputFile,
		QuietPeriod: 1 * time.Second,
		EndOfSecret: ".",
	}

	logger := slog.Default()
	eventsCh := make(chan events.Event, 100)

	word := "Hidden."

	r1 := &robot.Robot{
		ID:            1,
		SecretParts:   []robot.SecretPart{{Word: word}},
		LastUpdatedAt: time.Now().Add(-2 * time.Second),
	}
	r2 := &robot.Robot{
		ID:            2,
		SecretParts:   []robot.SecretPart{{Word: word}},
		LastUpdatedAt: time.Now().Add(-2 * time.Second),
	}

	w1 := workers.NewConvergenceDetectorWorker(cfg, logger, r1, eventsCh)
	w2 := workers.NewConvergenceDetectorWorker(cfg, logger, r2, eventsCh)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go w1.Run(ctx)
	go w2.Run(ctx)

	<-ctx.Done()

	content, err := os.ReadFile(outputFile)
	ass.NoError(err)
	ass.Equal(word, string(content))
}
