package tests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/pkg/events"
	"robots/pkg/workers"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerGossip_Robustness(t *testing.T) {
	ass := assert.New(t)

	logger := slog.Default()

	cfg := conf.Config{
		NbrOfRobots:            6,
		Secret:                 "Hidden beneath the old oak tree, golden coins patiently await discovery.",
		BufferSize:             10,
		EndOfSecret:            ".",
		PercentageOfLost:       50, // simulate loss
		PercentageOfDuplicated: 50, // simulate duplication
		DuplicatedNumber:       2,
		MaxAttempts:            5,
		GossipTime:             100 * time.Millisecond,
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
		PercentageOfLost:       20,
		PercentageOfDuplicated: 20,
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
	event := make(chan events.Event, 100)
	once := &sync.Once{}
	buffer := bytes.Buffer{}

	// start workers
	for _, r := range robots {
		go workers.NewMergeSecretWorker(slog.Default(), r, event).Run(ctx)
		go workers.NewProcessSummaryWorker(slog.Default(), r, robots, event).Run(ctx)
		go workers.NewStartGossipWorker(cfg, slog.Default(), r, robots, event).Run(ctx)
		go workers.NewConvergenceDetectorWorker(cfg, slog.Default(), r, make(chan *robot.Robot, 1), once, &buffer).Run(ctx)
	}

	// ASSERTION
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		content, err := io.ReadAll(&buffer)
		require.NoError(c, err)
		require.Equal(c, secret, string(content))
	}, 2*time.Second, 50*time.Millisecond)
}

func TestDoesNotWriteFileBeforeQuiescence(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "result.txt")

	cfg := conf.Config{
		NbrOfRobots: 5,
		Secret:      "hello world.",
		OutputFile:  outputFile,
		EndOfSecret: ".",

		BufferSize:             50,
		PercentageOfLost:       0,
		PercentageOfDuplicated: 0,

		GossipTime:     20 * time.Millisecond,
		QuietPeriod:    5 * time.Second,        // ⚠️ volontairement long
		Timeout:        500 * time.Millisecond, // ⛔ plus court que QuietPeriod
		MetricInterval: 100 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	sm := robot.SecretManager{Config: cfg}
	robots := sm.CreateRobots(strings.Fields(cfg.Secret))

	eventsCh := make(chan events.Event, 100)
	winnerCh := make(chan *robot.Robot, 1)
	once := &sync.Once{}
	buffer := bytes.Buffer{}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Start workers
	for _, r := range robots {
		go workers.NewMergeSecretWorker(logger, r, eventsCh).Run(ctx)
		go workers.NewProcessSummaryWorker(logger, r, robots, eventsCh).Run(ctx)
		go workers.NewStartGossipWorker(cfg, logger, r, robots, eventsCh).Run(ctx)
		go workers.NewConvergenceDetectorWorker(cfg, logger, r, winnerCh, once, &buffer).Run(ctx)
	}

	// Attendre la fin du contexte (timeout)
	<-ctx.Done()

	// ⛔ ASSERTION PRINCIPALE : le fichier ne doit PAS exister
	_, err := os.Stat(outputFile)

	if err == nil {
		t.Fatalf("output file should NOT be written before quiet period")
	}

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("unexpected error while checking output file: %v", err)
	}
}

func TestMainTimeoutPreventsFileWrite(t *testing.T) {
	outputFile := t.TempDir() + "/secret.txt"
	cfg := conf.Config{
		QuietPeriod: 5 * time.Second,
		Timeout:     100 * time.Millisecond, // comme dans le main
		OutputFile:  outputFile,
	}

	r := &robot.Robot{
		ID:            0,
		SecretParts:   []robot.SecretPart{{Index: 0, Word: "hello"}},
		LastUpdatedAt: time.Now(),
	}

	winner := make(chan *robot.Robot, 1)
	once := &sync.Once{}
	log := slog.Default()
	buffer := bytes.Buffer{}

	worker := workers.NewConvergenceDetectorWorker(cfg, log, r, winner, once, &buffer)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	go worker.Run(ctx)

	<-ctx.Done()

	_, err := os.Stat(outputFile)
	if err == nil {
		t.Fatal("Le fichier a été écrit, mais il ne devrait pas l'être avant la quiescence")
	}
}

func TestFileNotWrittenIfTimeoutBeforeQuietPeriod(t *testing.T) {
	outputFile := t.TempDir() + "/secret.txt"
	cfg := conf.Config{
		QuietPeriod: 5 * time.Second,
		Timeout:     1 * time.Second, // timeout < quietPeriod
		OutputFile:  outputFile,
	}

	// Simuler des robots actifs
	robots := make([]robot.Robot, 6)
	winner := make(chan *robot.Robot, 1)
	once := &sync.Once{}
	buffer := bytes.Buffer{}
	log := slog.Default()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	for i := 0; i < len(robots); i++ {
		w := workers.NewConvergenceDetectorWorker(cfg, log, &robots[i], winner, once, &buffer)
		go w.Run(ctx)
	}

	<-ctx.Done()

	if _, err := os.Stat(outputFile); err == nil {
		t.Fatal("Le fichier a été écrit, mais il ne devrait pas l'être !")
	}
}

func TestWinnerChannelBlocksFileWrite(t *testing.T) {
	ass := require.New(t)
	outputFile := "test_secret.txt"
	_ = os.Remove(outputFile)

	winnerChan := make(chan *robot.Robot) // unbuffered, pour simuler blocage
	cfg := conf.Config{
		OutputFile:  outputFile,
		QuietPeriod: 1 * time.Second,
		EndOfSecret: ".",
	}
	log := slog.Default()
	once := &sync.Once{}
	buffer := bytes.Buffer{}
	word := "Hidden."

	// Robot 1 trouve le secret
	r1 := &robot.Robot{
		ID:            1,
		SecretParts:   []robot.SecretPart{{Word: word}},
		LastUpdatedAt: time.Now().Add(-2 * time.Second),
	}
	w1 := workers.NewConvergenceDetectorWorker(cfg, log, r1, winnerChan, once, &buffer)

	// Robot 2 arrive après
	r2 := &robot.Robot{
		ID:            2,
		SecretParts:   []robot.SecretPart{{Word: word}},
		LastUpdatedAt: time.Now().Add(-2 * time.Second),
	}

	w2 := workers.NewConvergenceDetectorWorker(cfg, log, r2, winnerChan, once, &buffer)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Lancer le worker 1 dans une goroutine (bloque sur channel)
	go func() { _ = w1.Run(ctx) }()

	// Lancer le worker 2 dans une goroutine
	go func() { _ = w2.Run(ctx) }()

	<-ctx.Done()

	// Then buffer is not empty
	secret, err := io.ReadAll(&buffer)
	ass.Equal(word, string(secret))
	ass.NoError(err)
}
