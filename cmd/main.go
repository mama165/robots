package main

import (
	"fmt"
	"github.com/Netflix/go-env"
	"log"
	"math/rand"
	"robots/internal/robot"
	"time"
)

var (
	config robot.Config
	winner chan robot.Robot
)

func main() {
	rand.Seed(time.Now().UnixNano())
	_, err := env.UnmarshalFromEnviron(&config)
	if err != nil {
		panic(err)
	}
	if config.NbrOfRobots <= 0 {
		log.Panicf("number of robots should be positive : %d", config.NbrOfRobots)
	}
	if config.BufferSize <= 0 {
		log.Panicf("buffer size should be positive : %d", config.BufferSize)
	}
	if config.PercentageOfLost < 0 {
		panic(fmt.Errorf("percentage of lost should be positive : %d", config.PercentageOfLost))
	}
	if config.PercentageOfDuplicated < 0 {
		log.Panicf("percentage of lost should be positive : %d", config.PercentageOfDuplicated)
	}
	if config.DuplicatedNumber < 0 {
		log.Panicf("duplicated number should be positive : %d", config.DuplicatedNumber)
	}
	if config.MaxAttempts <= 0 {
		log.Panicf("max attempts should be positive : %d", config.MaxAttempts)
	}

	// Add a timeout for the overall execution
	timeout := time.After(config.Timeout)

	secretManager := robot.SecretManager{Config: config}
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
			if err = secretManager.WriteSecret(r.BuildSecret()); err != nil {
				panic(err)
			}
			log.Printf("Robot %d won and saved the message in file -> %s", r.ID, config.OutputFile)
			for _, robotChan := range robots {
				// Properly close the channel
				close(robotChan.Inbox)
			}
			return
		case <-timeout:
			log.Printf("Timeout after %s", config.Timeout)
			return
		default:
			time.Sleep(10 * time.Millisecond) // To avoid the 100% of CPU
		}
	}
}
