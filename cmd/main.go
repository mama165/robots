package main

import (
	"fmt"
	"github.com/Netflix/go-env"
	"log"
	"math/rand"
	"os"
	"robots/internal/robot"
	"strings"
	"time"
)

var (
	config robot.Config
	done   chan int
)

func main() {
	rand.Seed(time.Now().UnixNano())
	_, err := env.UnmarshalFromEnviron(&config)
	if err != nil {
		panic(err)
	}
	if config.NbrOfRobots <= 0 {
		panic(fmt.Errorf("number of robots should be positive : %d", config.NbrOfRobots))
	}
	if config.BufferSize <= 0 {
		panic(fmt.Errorf("buffer size should be positive : %d", config.BufferSize))
	}
	if config.PercentageOfLost < 0 {
		panic(fmt.Errorf("percentage of lost should be positive : %d", config.PercentageOfLost))
	}
	if config.PercentageOfDuplicated < 0 {
		panic(fmt.Errorf("percentage of lost should be positive : %d", config.PercentageOfDuplicated))
	}
	if config.DuplicatedNumber < 0 {
		panic(fmt.Errorf("duplicated number should be positive : %d", config.DuplicatedNumber))
	}
	if config.MaxAttempts <= 0 {
		panic(fmt.Errorf("max attempts should be positive : %d", config.MaxAttempts))
	}

	secretManager := robot.SecretManager{Config: config}
	secret := secretManager.SplitSecret(config.Secret)
	robots := secretManager.CreateRobots(secret, config.NbrOfRobots, config.BufferSize)
	done = make(chan int)

	for _, r := range robots {
		go r.Start(done, secret)
	}

	for {
		secretManager.ExchangeMessage(robots, config.PercentageOfLost, config.PercentageOfDuplicated, config.DuplicatedNumber, config.MaxAttempts)
		select {
		case id := <-done:
			file, err := os.Create(config.OutputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			if _, err := file.WriteString(strings.Join(robots[id].Words, " ")); err != nil {
				panic(err)
			}
			log.Printf("Robot %d saved the message -> %s", id, config.OutputFile)
			for _, robot := range robots {
				close(robot.Inbox)
			}
			return
		default:
		}
	}
}
