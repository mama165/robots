package robot

import (
	"fmt"
	"github.com/samber/lo"
	"log/slog"
	"math/rand"
	"os"
	"slices"
	"sort"
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
	LogLevel               string        `env:"LOG_LEVEL,default=INFO"`
}

type Robot struct {
	ID            int // Index of the robots
	SecretParts   []SecretPart
	Inbox         chan Inbox
	LastUpdatedAt time.Time // Necessary to know if no words have been received since a long time
}

// SecretPart Represents a word and the position from the secret
type SecretPart struct {
	Index int // Index of the word
	Word  string
}

// Inbox represents the message sent by a robot
type Inbox struct {
	From        int // Index of the robot (only for logging)
	SecretParts []SecretPart
}

// GetWords Returns words contained in the robot
// Can be ordered or unordered by index of the initial secret
func (r *Robot) GetWords(ordered bool) []string {
	parts := r.SecretParts
	if ordered {
		tmp := make([]SecretPart, len(parts))
		copy(tmp, parts)
		sort.Slice(tmp, func(i, j int) bool {
			return tmp[i].Index < tmp[j].Index
		})
		parts = tmp
	}
	return lo.Map(parts, func(p SecretPart, _ int) string {
		return p.Word
	})
}

func (r *Robot) BuildSecret() string {
	return strings.Join(r.GetWords(true), " ")
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
	Log    *slog.Logger
}

// SplitSecret Initial sentences split into words
func (s SecretManager) SplitSecret(word string) []string {
	return strings.Fields(word)
}

// CreateRobots Randomly assign words to n robots
// Each of the contains word with indexes
func (s SecretManager) CreateRobots(words []string) []*Robot {
	robots := make([]*Robot, s.Config.NbrOfRobots)
	for i := 0; i < s.Config.NbrOfRobots; i++ {
		robots[i] = &Robot{
			ID:            i,
			SecretParts:   []SecretPart{},
			Inbox:         make(chan Inbox, s.Config.BufferSize),
			LastUpdatedAt: time.Unix(0, 0),
		}
	}

	for index, word := range words {
		key := rand.Intn(s.Config.NbrOfRobots)
		secretPart := SecretPart{Index: index, Word: word}
		robots[key].SecretParts = append(robots[key].SecretParts, secretPart)
	}
	return robots
}

// StartRobot Collect all messages exchanged
// Opened as long as robot communicate
func (s SecretManager) StartRobot(robot *Robot, winner chan<- Robot) {
	for inbox := range robot.Inbox {
		for _, secretPart := range inbox.SecretParts {
			// Updating LastUpdatedAt if the word doesn't exist
			if !slices.Contains(robot.SecretParts, secretPart) {
				robot.LastUpdatedAt = time.Now().UTC()
				robot.SecretParts = append(robot.SecretParts, secretPart)
				continue
			}
		}
		// For each inbox merged with word
		// Check the secret has been completed
		// Only if no update since a chosen duration
		elapsed := robot.LastUpdatedAt.Add(s.Config.QuietPeriod).Before(time.Now().UTC())
		if elapsed && IsSecretCompleted(robot.GetWords(false), s.Config.EndOfSecret) {
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
		msg := Inbox{From: sender.ID, SecretParts: sender.SecretParts}

		s.Log.Debug(fmt.Sprintf("Robot %d communicates with robot %d", sender.ID, receiver.ID))

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

// WriteSecret Write the secret in a file
func (s SecretManager) WriteSecret(secret string) error {
	file, err := os.Create(s.Config.OutputFile)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(secret)
	return err
}

// IsSecretCompleted Verify if a word contains a "."
func IsSecretCompleted(words []string, endOfSecret string) bool {
	for _, word := range words {
		if strings.HasSuffix(word, endOfSecret) {
			return true
		}
	}
	return false
}
