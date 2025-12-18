package robot

import (
	"context"
	"math/rand"
	"robots/internal/conf"
	pb "robots/proto"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
)

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
	Config conf.Config
}

type ID int

func (id ID) ToInt() int {
	return int(id)
}

// Robot GossipSummary and GossipUpdate have to be inside the robot
// Because at any moment robot exchange with others
// They should have their own snapshot
type Robot struct {
	mu            sync.RWMutex
	ID            ID // Index of the robots
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

func ChooseRobot(current *Robot, robots []*Robot) *Robot {
	var receiver *Robot
	for {
		receiver = robots[rand.Intn(len(robots))]
		if receiver.ID != current.ID {
			break
		}
	}
	return receiver
}

func (r *Robot) Indexes() []int64 {
	return lo.Map(r.SecretParts, func(item SecretPart, _ int) int64 {
		return int64(item.Index)
	})
}

// GetWordsToSend retourne tous les mots que le destinataire n'a pas encore
func (r *Robot) GetWordsToSend(receiverIndexes []int) []SecretPart {
	var missing []SecretPart
	for _, sp := range r.SecretParts {
		if !lo.Contains(receiverIndexes, sp.Index) {
			missing = append(missing, sp)
		}
	}
	return missing
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
func (s SecretManager) CreateRobots(words []string) []*Robot {
	robots := make([]*Robot, s.Config.NbrOfRobots)
	for i := 0; i < s.Config.NbrOfRobots; i++ {
		robots[i] = &Robot{
			ID:            ID(i),
			SecretParts:   []SecretPart{},
			GossipSummary: make(chan []byte, s.Config.BufferSize),
			GossipUpdate:  make(chan []byte, s.Config.BufferSize),
			LastUpdatedAt: time.Now().UTC(),
		}
	}

	for index, word := range words {
		key := rand.Intn(s.Config.NbrOfRobots)
		secretPart := SecretPart{Index: index, Word: word}
		robots[key].SecretParts = append(robots[key].SecretParts, secretPart)
	}
	return robots
}

// MergeSecretPart merges a secret part into the robot's local state.
// Invariants enforced:
// - Monotonicity: secret parts are never removed.
// - Uniqueness: a given index can map to only one word.
// - Idempotence: receiving the same (index, word) multiple times has no effect.
// Behavior:
// - If the index already exists with a different word, this is a fatal invariant violation and triggers a panic.
// - If the index already exists with the same word, the update is ignored.
// - If the part is new, it is appended and LastUpdatedAt is refreshed.
//
// This method is the single entry point for mutating SecretParts
// and acts as the consistency boundary of the Robot.
func (r *Robot) MergeSecretPart(secretPart SecretPart) {
	r.mu.Lock()
	defer r.mu.Unlock()
	part, ok := findSecretPart(r.SecretParts, secretPart)
	if ok && part.Word != secretPart.Word {
		panic("invariant violation: same index, different word")
	}
	if ok {
		return
	}
	r.LastUpdatedAt = time.Now().UTC()
	r.SecretParts = append(r.SecretParts, secretPart)
}

func findSecretPart(secretParts []SecretPart, secretPart SecretPart) (SecretPart, bool) {
	return lo.Find(secretParts, func(item SecretPart) bool {
		return item.Index == secretPart.Index
	})
}

// IsSecretCompleted reports whether the robot has fully reconstructed the secret.
// The secret is considered complete if:
// - all indexes from 0 to the highest index are present (no gaps),
// - and the last word ends with the given end-of-secret marker.
// This prevents false positives caused by partial, unordered, or duplicated gossip messages.
func (r *Robot) IsSecretCompleted(endOfSecret string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	parts := r.SecretParts
	if len(parts) == 0 {
		return false
	}

	maxIndex := -1
	indexes := make(map[int]struct{}, len(parts))
	for _, p := range parts {
		indexes[p.Index] = struct{}{}
		if p.Index > maxIndex {
			maxIndex = p.Index
		}
	}

	// Check all indexes are present
	for i := 0; i <= maxIndex; i++ {
		if _, ok := indexes[i]; !ok {
			return false // Missing word !
		}
	}

	// Check end marker
	words := r.GetWords(true)
	return strings.HasSuffix(words[len(words)-1], endOfSecret)
}

func FromSecretPartsPb(secretPartsPb []*pb.SecretPart) []SecretPart {
	return lo.Map(secretPartsPb, func(item *pb.SecretPart, _ int) SecretPart {
		return SecretPart{Index: int(item.Index), Word: item.Word}
	})
}

func ToSecretPartsPb(secretParts []SecretPart) []*pb.SecretPart {
	return lo.Map(secretParts, func(item SecretPart, _ int) *pb.SecretPart {
		return &pb.SecretPart{
			Index: int64(item.Index),
			Word:  item.Word,
		}
	})
}
