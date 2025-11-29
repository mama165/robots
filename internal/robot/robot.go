package robot

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/samber/lo"
	"log/slog"
	"math/rand"
	"os"
	robotpb "robots/proto/pb-go"
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
	GossipTime             time.Duration `env:"GOSSIP_TIME,required=true"`
	LogLevel               string        `env:"LOG_LEVEL,default=INFO"`
}

type ISecretManager interface {
	SplitSecret(word string) []string
	CreateRobots(words []string) []Robot
	FindSecret(robots []Robot)
	StartGossip(robot Robot, robots []Robot)
	CollectMessages(robot Robot, winner chan<- Robot)
	ExchangeMessage(sender, receiver Robot) int
	WriteSecret(secret string) error
}

type SecretManager struct {
	Config Config
	Log    *slog.Logger
}

type Robot struct {
	ID            int // Index of the robots
	SecretParts   []SecretPart
	Message       chan []byte
	LastUpdatedAt time.Time // Necessary to know if no words have been received since a long time
}

// SecretPart Represents a word and the position from the secret
type SecretPart struct {
	Index int // Index of the word
	Word  string
}

func ChooseRobot(current Robot, robots []Robot) Robot {
	var receiver Robot
	for {
		receiver = robots[rand.Intn(len(robots))]
		if receiver.ID != current.ID {
			break
		}
	}
	return receiver
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

// SplitSecret Initial sentences split into words
func (s SecretManager) SplitSecret(word string) []string {
	return strings.Fields(word)
}

// CreateRobots Randomly assign words to n robots
// Each of the contains word with indexes
func (s SecretManager) CreateRobots(words []string) []Robot {
	robots := make([]Robot, s.Config.NbrOfRobots)
	for i := 0; i < s.Config.NbrOfRobots; i++ {
		robots[i] = Robot{
			ID:            i,
			SecretParts:   []SecretPart{},
			Message:       make(chan []byte, s.Config.BufferSize),
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

// FindSecret Start the secret exchange
func (s SecretManager) FindSecret(robots []Robot) {
	winner := make(chan Robot)
	timeout := time.After(s.Config.Timeout)

	// Running two goroutines for each robot to start
	for _, robot := range robots {
		go s.CollectMessages(robot, winner)
		go s.StartGossip(robot, robots)
	}

	for {
		select {
		case r := <-winner:
			// A winner has been found
			if err := s.WriteSecret(r.BuildSecret()); err != nil {
				panic(err)
			}
			s.Log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", r.ID, s.Config.OutputFile))
			for _, robotChan := range robots {
				// Properly close the channel
				close(robotChan.Message)
			}
			return
		case <-timeout:
			s.Log.Info(fmt.Sprintf("Timeout after %s", s.Config.Timeout))
			return
		}
	}
}

func (s SecretManager) StartGossip(robot Robot, robots []Robot) {
	ticker := time.NewTicker(s.Config.GossipTime)
	defer ticker.Stop()
	for range ticker.C {
		receiver := ChooseRobot(robot, robots)
		s.ExchangeMessage(robot, receiver)
	}
}

// CollectMessages Collect all messages exchanged
// Opened as long as robot communicate
func (s SecretManager) CollectMessages(robot Robot, winner chan<- Robot) {
	for message := range robot.Message {
		var decoded robotpb.Message
		if err := proto.Unmarshal(message, &decoded); err != nil {
			s.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
			continue
		}
		secretParts := fromSecretPartsPb(decoded.SecretParts)
		for _, secretPart := range secretParts {
			// Updating LastUpdatedAt if the word doesn't exist
			if !containsIndex(robot.SecretParts, secretPart.Index) {
				robot.LastUpdatedAt = time.Now().UTC()
				robot.SecretParts = append(robot.SecretParts, secretPart)
				continue
			}
		}
		// For each message merged with word
		// Check the secret has been completed
		// Only if no update since a chosen duration
		elapsed := robot.LastUpdatedAt.Add(s.Config.QuietPeriod).Before(time.Now().UTC())
		if elapsed && IsSecretCompleted(robot.GetWords(false), s.Config.EndOfSecret) {
			winner <- robot
			return
		}
	}
}

func containsIndex(secretParts []SecretPart, index int) bool {
	return lo.ContainsBy(secretParts, func(item SecretPart) bool {
		return item.Index == index
	})
}

// ExchangeMessage r1 send a message to r2
// Simulate lost and duplicated messages
func (s SecretManager) ExchangeMessage(sender, receiver Robot) int {
	if sender.ID == receiver.ID {
		return 0
	}
	messageSent := 0
	for i := 0; i < s.Config.MaxAttempts; i++ {
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
			message := robotpb.Message{
				From:        int64(sender.ID),
				SecretParts: toSecretPartsPb(sender.SecretParts),
			}
			msg, err := proto.Marshal(&message)
			if err != nil {
				s.Log.Info(fmt.Sprintf("Unable to encode proto message : %s", err.Error()))
				continue
			}
			select {
			case receiver.Message <- msg: // For the range iterating over messages
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

func fromSecretPartsPb(secretPartsPb []*robotpb.SecretPart) []SecretPart {
	return lo.Map(secretPartsPb, func(item *robotpb.SecretPart, _ int) SecretPart {
		return SecretPart{Index: int(item.Index), Word: item.Word}
	})
}

func toSecretPartsPb(secretParts []SecretPart) []*robotpb.SecretPart {
	return lo.Map(secretParts, func(item SecretPart, _ int) *robotpb.SecretPart {
		return &robotpb.SecretPart{
			Index: int64(item.Index),
			Word:  item.Word,
		}
	})
}
