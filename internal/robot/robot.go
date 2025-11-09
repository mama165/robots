package robot

import (
	"math/rand"
	"slices"
	"strings"
)

type Config struct {
	NbrOfRobots            int    `env:"NBR_OF_ROBOTS"`
	Secret                 string `env:"SECRET"`
	OutputFile             string `env:"OUTPUT_FILE"`
	BufferSize             int    `env:"BUFFER_SIZE"`
	PercentageOfLost       int    `env:"PERCENTAGE_OF_LOST"`
	PercentageOfDuplicated int    `env:"PERCENTAGE_OF_DUPLICATED"`
	DuplicatedNumber       int    `env:"DUPLICATED_NUMBER"`
	MaxAttempts            int    `env:"MAX_ATTEMPTS"`
}

type Robot struct {
	ID    int // Index of the robots
	Words []string
	Inbox chan Inbox
}

type Inbox struct {
	From  int // Index of the robot
	Words []string
}

// SplitSecret Initial sentences split into words
func SplitSecret(word string) []string {
	return strings.Fields(word)
}

// CreateRobots Randomly assign words to n robots
func CreateRobots(words []string, nbrOfRobots, bufferSize int) []*Robot {
	robots := make([]*Robot, nbrOfRobots)
	for i := 0; i < nbrOfRobots; i++ {
		robots[i] = &Robot{
			ID:    i,
			Words: []string{},
			Inbox: make(chan Inbox, bufferSize),
		}
	}

	for _, word := range words {
		key := rand.Intn(nbrOfRobots)
		robots[key].Words = append(robots[key].Words, word)
	}
	return robots
}

func (r *Robot) Start(done chan<- int, secret []string) {
	for inbox := range r.Inbox {
		for _, w := range inbox.Words {
			if !slices.Contains(r.Words, w) {
				r.Words = append(r.Words, w)
			}
		}
		if HasReconstructedSecret(secret, r.Words) {
			done <- r.ID
		}
	}
}

// ExchangeMessage r1 send a message to r2
// Simulate lost and duplicated messages
func ExchangeMessage(robots []*Robot, percentageOfLost, percentageOfDuplicated, duplicatedNumber, maxAttempts int) int {
	messageSent := 0
	size := len(robots)
	for i := 0; i < maxAttempts; i++ {
		r1 := rand.Intn(size)
		r2 := rand.Intn(size)
		if r1 == r2 {
			continue
		}
		robot1 := robots[r1]
		robot2 := robots[r2]
		msg := Inbox{From: r1, Words: robot1.Words}

		// Percentage of lost messages
		if rand.Float32() < float32(percentageOfLost)/100.0 {
			continue
		}

		// Percentage of duplicated messages
		var times int
		if rand.Float32() < float32(percentageOfDuplicated)/100.0 {
			times = duplicatedNumber
		}

		for j := 0; j <= times; j++ {
			select {
			case robot2.Inbox <- msg:
				messageSent++
			default:
				continue
			}
		}
	}
	return messageSent
}

// HasReconstructedSecret
// Checking if a robot has finished rebuilding the message
func HasReconstructedSecret(expected, words []string) bool {
	occurrences := make(map[string]int)
	for _, word := range expected {
		occurrences[word]++
	}
	for _, word := range words {
		if _, ok := occurrences[word]; ok {
			occurrences[word]--
		}
	}
	for _, count := range occurrences {
		if count > 0 {
			return false
		}
	}
	return true
}
