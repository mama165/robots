package robot

import (
	"context"
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
	FindSecret(ctx context.Context, robots []Robot)
	StartGossip(ctx context.Context, robot *Robot, robots []Robot)
	ProcessSummary(ctx context.Context, robot *Robot)
	SuperviseRobot(ctx context.Context, robot *Robot, winner chan Robot)
	ExchangeMessage(ctx context.Context, sender, receiver *Robot)
	WriteSecret(winner chan Robot)
}

type SecretManager struct {
	Config Config
	Log    *slog.Logger
}

// Robot GossipSummary and GossipUpdate have to be inside the robot
// Because at any moment robot exchange with others
// They should have their own snapshot
type Robot struct {
	ID            int // Index of the robots
	SecretParts   []SecretPart
	GossipSummary chan []byte // Represents a channel of current indexes of robots
	GossipUpdate  chan []byte // Represents a channel of missing secretParts
	LastUpdatedAt time.Time   // Necessary to know if no words have been received since a long time
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

func (r *Robot) updateSecretParts(secretPart SecretPart) {
	r.SecretParts = append(r.SecretParts, secretPart)
}

func (r *Robot) Indexes() []int64 {
	return lo.Map(r.SecretParts, func(item SecretPart, index int) int64 {
		return int64(item.Index)
	})
}

func (r *Robot) getMissingParts(indexes []int) []SecretPart {
	return nil
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
			GossipSummary: make(chan []byte, s.Config.BufferSize),
			GossipUpdate:  make(chan []byte, s.Config.BufferSize),
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
func (s SecretManager) FindSecret(ctx context.Context, robots []Robot) {
	winner := make(chan Robot)

	// Running two goroutines for each robot to start
	for _, robot := range robots {
		go s.ProcessSummary(ctx, &robot)
		go s.Update(ctx, &robot)
		go s.SuperviseRobot(ctx, &robot, winner)
		go s.StartGossip(&robot, robots)
	}
	// Only the winner goroutine handle the writing
	go s.WriteSecret(ctx, winner)
}

func (s SecretManager) StartGossip(robot *Robot, robots []Robot) {
	ticker := time.NewTicker(s.Config.GossipTime)
	defer ticker.Stop()
	for range ticker.C {
		receiver := ChooseRobot(*robot, robots)
		s.ExchangeMessage(robot, &receiver)
	}
}

func (s SecretManager) Update(ctx context.Context, robot *Robot) {
	for {
		select {
		case updateMsg := <-robot.GossipUpdate:
			// On récupère les parties manquantes venant de n'importe qui
			// On sait qu'on a récupéré uniquement les parties manquantes car c'est du gossip push-pull
			var gossipUpdate robotpb.GossipUpdate
			err := proto.Unmarshal(updateMsg, &gossipUpdate)
			if err != nil {
				s.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
				continue
			}
			// TODO il faudra transformer robot.SecretParts en chan []byte pour écrire dedans
			//TODO Doit-on contiuer à vérifier si on contient déjà le mot ?
			secretParts := fromSecretPartsPb(gossipUpdate.SecretParts)
			for _, secretPart := range secretParts {
				// Updating LastUpdatedAt if the word doesn't exist
				if !ContainsIndex(robot.SecretParts, secretPart.Index) {
					robot.LastUpdatedAt = time.Now().UTC()
					robot.SecretParts = append(robot.SecretParts, secretPart)
					continue
				}
			}
		case <-ctx.Done():
			s.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
		}
	}
}

func (s SecretManager) ProcessSummary(ctx context.Context, robot *Robot) {
	for {
		select {
		case summaryMsg := <-robot.GossipSummary:
			// On doit donc retourner les secretParts manquant
			var gossipSummary robotpb.GossipSummary
			if err := proto.Unmarshal(summaryMsg, &gossipSummary); err != nil {
				s.Log.Info(fmt.Sprintf("Unable to decode proto message : %s", err.Error()))
				continue
			}
			// TODO Retourner les mots manquants ici
			indexes := lo.Map(gossipSummary.Indexes, func(item int64, _ int) int {
				return int(item)
			})
			// TODO construire getMissingParts
			secretParts := robot.getMissingParts(indexes)
			msg, err := proto.Marshal(&robotpb.GossipUpdate{SecretParts: toSecretPartsPb(secretParts)})
			if err != nil {
				s.Log.Info(fmt.Sprintf("Unable to encode proto message : %s", err.Error()))
				continue
			}
			// ⚠️ Don't forget to add a select case and default (not just writing)
			// ⚠️ If the channel robot.GossipUpdate is slowly dequeued
			// ⚠️ Can block the process
			select {
			case robot.GossipUpdate <- msg:
				// Successfully sent the message
			default:
				s.Log.Debug(fmt.Sprintf("Robot %d : buffer is full, message is ignored", robot.ID))
			}
		case <-ctx.Done():
			s.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
		}
	}
}

// SuperviseRobot
// For each message merged with word
// Check the secret has been completed
// Only if no update since a chosen duration
func (s SecretManager) SuperviseRobot(ctx context.Context, robot *Robot, winner chan Robot) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := robot.LastUpdatedAt.Add(s.Config.QuietPeriod).Before(time.Now().UTC())
			if elapsed && robot.IsSecretCompleted(s.Config.EndOfSecret) {
				// Send the winner in the channel without blocking any other possible winner
				select {
				case winner <- *robot:
					s.Log.Info(fmt.Sprintf("Robot %d won", robot.ID))
				case <-ctx.Done():
					s.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
				default:
					s.Log.Debug(fmt.Sprintf("Robot %d wanted to win but another one won", robot.ID))
				}
				return
			}
		case <-ctx.Done():
			s.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
		}
	}
}

func ContainsIndex(secretParts []SecretPart, index int) bool {
	return lo.ContainsBy(secretParts, func(item SecretPart) bool {
		return item.Index == index
	})
}

// ExchangeMessage r1 send a message to r2
// Simulate lost and duplicated messages
func (s SecretManager) ExchangeMessage(sender, receiver *Robot) {
	if sender.ID == receiver.ID {
		return
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
			// Sender sends his own indexes to receiver
			gossipSender := robotpb.GossipSummary{Indexes: sender.Indexes()}
			msgSender, err := proto.Marshal(&gossipSender)
			if err != nil {
				s.Log.Info(fmt.Sprintf("Unable to encode proto message : %s", err.Error()))
				continue
			}
			select {
			case receiver.GossipSummary <- msgSender:
				messageSent++
			default:
			}
		}
	}
}

// WriteSecret Write the secret in a file
func (s SecretManager) WriteSecret(ctx context.Context, winner chan Robot) {
	for {
		select {
		case robot := <-winner:
			file, err := os.Create(s.Config.OutputFile)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			if _, err = file.WriteString(robot.BuildSecret()); err != nil {
				panic(err)
			}
			s.Log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", robot.ID, s.Config.OutputFile))
			return
		case <-ctx.Done():
			s.Log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
		}
	}
}

// IsSecretCompleted Verify if a word contains a "."
func (r *Robot) IsSecretCompleted(endOfSecret string) bool {
	for _, word := range r.GetWords(false) {
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
