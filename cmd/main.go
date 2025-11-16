package main

import (
	"fmt"
	"github.com/Netflix/go-env"
	"log/slog"
	"math/rand"
	"os"
	"robots/internal/robot"
	"time"
)

var (
	config robot.Config
	winner chan robot.Robot
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if _, err := env.UnmarshalFromEnviron(&config); err != nil {
		panic(err)
	}
	log := NewLogger(config.LogLevel)
	if err := validateEnvVariables(); err != nil {
		log.Info(err.Error())
		panic(err)
	}

	// Add a timeout for the overall execution
	timeout := time.After(config.Timeout)

	secretManager := robot.SecretManager{Config: config, Log: log}
	secret := secretManager.SplitSecret(config.Secret)
	robots := secretManager.CreateRobots(secret)
	winner = make(chan robot.Robot)

	// Running one goroutine for each robot to start
	for _, r := range robots {
		go secretManager.StartRobot(r, winner)
	}

	for {
		size := len(robots)
		sender := rand.Intn(size)
		receiver := rand.Intn(size)
		if sender == receiver {
			continue
		}
		secretManager.ExchangeMessage(*robots[sender], *robots[receiver])
		select {
		case r := <-winner:
			// A winner has been found
			if err := secretManager.WriteSecret(r.BuildSecret()); err != nil {
				panic(err)
			}
			log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", r.ID, config.OutputFile))
			for _, robotChan := range robots {
				// Properly close the channel
				close(robotChan.Inbox)
			}
			return
		case <-timeout:
			log.Info(fmt.Sprintf("Timeout after %s", config.Timeout))
			return
		default:
			time.Sleep(10 * time.Millisecond) // To avoid the 100% of CPU
		}
	}
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

func validateEnvVariables() error {
	if config.NbrOfRobots <= 0 {
		return fmt.Errorf("number of robots should be positive : %d", config.NbrOfRobots)
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
