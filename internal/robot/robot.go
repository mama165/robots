package robot

import (
	"log"
	"math/rand"
	"os"
	"slices"
	"strings"
	"time"
)

type Config struct {
	NbrOfRobots            int           `env:"NBR_OF_ROBOTS,required=true"`
	Secret                 string        `env:"SECRET,required=true"`
	OutputFile             string        `env:"OUTPUT_FILE,required=true"`
	BufferSize             int           `env:"BUFFER_SIZE,required=true"`
	EndOfSecret            string        `env:"END_OF_SECRET,required=true"`
	PercentageOfLost       int           `env:"PERCENTAGE_OF_LOST,required=true"`
	PercentageOfDuplicated int           `env:"PERCENTAGE_OF_DUPLICATED,required=true"`
	DuplicatedNumber       int           `env:"DUPLICATED_NUMBER,required=true"`
	MaxAttempts            int           `env:"MAX_ATTEMPTS,required=true"`
	Timeout                time.Duration `env:"TIMEOUT,required=true"`
	QuietPeriod            time.Duration `env:"QUIET_PERIOD,required=true"`
}

type Robot struct {
	ID            int // Index of the robots
	Words         []string
	Inbox         chan Inbox
	LastUpdatedAt time.Time // Necessary to know if no words have been received since a long time
}

func (r *Robot) BuildSecret() string {
	return strings.Join(r.Words, " ")
}

// Inbox represents the message sent by a robot
type Inbox struct {
	From  int // Index of the robot (only for logging)
	Words []string
}
type ISecretManager interface {
	SplitSecret(word string) []string
	CreateRobots(words []string) []*Robot
	StartRobot(robot *Robot, winner chan<- Robot)
	ExchangeMessage(sender, receiver Robot) int
	WriteSecret(secret string) error
}

type SecretManager struct {
	Config Config
}

// SplitSecret Initial sentences split into words
func (s SecretManager) SplitSecret(word string) []string {
	return strings.Fields(word)
}

// CreateRobots Randomly assign words to n robots
func (s SecretManager) CreateRobots(words []string) []*Robot {
	robots := make([]*Robot, s.Config.NbrOfRobots)
	for i := 0; i < s.Config.NbrOfRobots; i++ {
		robots[i] = &Robot{
			ID:            i,
			Words:         []string{},
			Inbox:         make(chan Inbox, s.Config.BufferSize),
			LastUpdatedAt: time.Unix(0, 0),
		}
	}

	for _, word := range words {
		key := rand.Intn(s.Config.NbrOfRobots)
		robots[key].Words = append(robots[key].Words, word)
	}
	return robots
}

func (s SecretManager) StartRobot(robot *Robot, winner chan<- Robot) {
	// Collect all messages exchanged
	// Opened as long as robot communicate
	for inbox := range robot.Inbox {
		for _, word := range inbox.Words {
			// Updating LastUpdatedAt if the word doesn't exist
			// Supposing no duplicated words
			if !slices.Contains(robot.Words, word) {
				robot.LastUpdatedAt = time.Now().UTC()
				robot.Words = append(robot.Words, word)
				continue
			}
		}
		// For each inbox merged with word
		// Check the secret has been completed
		// Only if no update since 5 seconds
		elapsed := robot.LastUpdatedAt.Add(s.Config.QuietPeriod).Before(time.Now().UTC())

		if elapsed && IsSecretCompleted(robot.Words, s.Config.EndOfSecret) {
			winner <- *robot
			return
		}
	}
}

// ExchangeMessage r1 send a message to r2
// Simulate lost and duplicated messages
func (s SecretManager) ExchangeMessage(sender, receiver Robot) int {
	if sender.ID == receiver.ID {
		return 0
	}
	messageSent := 0
	for i := 0; i < s.Config.MaxAttempts; i++ {
		msg := Inbox{From: sender.ID, Words: sender.Words}

		log.Printf("Robot %d communicates with robot %d", sender.ID, receiver.ID)

		// Calculate and simulate a random percentage
		isSimulated := func(percentage int) bool {
			return rand.Float32() < float32(percentage)/100.0
		}

		// Percentage of lost messages
		if isSimulated(s.Config.PercentageOfLost) {
			continue
		}

		// Percentage of duplicated messages
		var times int
		if isSimulated(s.Config.PercentageOfDuplicated) {
			times = s.Config.DuplicatedNumber
		}

		for j := 0; j <= times; j++ {
			select {
			case receiver.Inbox <- msg: // For the range iterating over Inboxes
				messageSent++
			default:
			}
		}
	}
	return messageSent
}

func (s SecretManager) WriteSecret(secret string) error {
	file, err := os.Create(s.Config.OutputFile)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(secret)
	return err
}

// IsSecretCompleted Verify if the last word contains a "."
func IsSecretCompleted(words []string, endOfSecret string) bool {
	for _, w := range words {
		if strings.HasSuffix(w, endOfSecret) {
			return true
		}
	}
	return false
}
