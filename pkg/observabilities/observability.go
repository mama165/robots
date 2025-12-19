package observabilities

import (
	"sync"
	"time"
)

// ObservabilityWorker Store all metrics of workers
type ObservabilityWorker struct {
	mu                 sync.Mutex
	timestamp          time.Time
	messagesSent       map[int]int
	messagesReceived   map[int]int
	messagesLost       int
	messagesDuplicated int
	messagesReordered  int
	invariantViolation map[int]int
	workerRestarted    map[string]int
	lastActive         time.Time
	channelCapacity    map[string]ChannelCapacity
}

func NewObservabilityWorker() *ObservabilityWorker {
	return &ObservabilityWorker{
		timestamp:          time.Now(),
		messagesSent:       make(map[int]int),
		messagesLost:       0,
		messagesReceived:   make(map[int]int),
		messagesDuplicated: 0,
		messagesReordered:  0,
		invariantViolation: make(map[int]int),
		lastActive:         time.Now(),
		channelCapacity:    make(map[string]ChannelCapacity),
	}
}

type ChannelCapacity struct {
	Capacity int
	Length   int
}

func (s *ObservabilityWorker) IncSent(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesSent[id]++
}

func (s *ObservabilityWorker) IncLost() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesLost++
}

func (s *ObservabilityWorker) IncReceived(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesReceived[id]++
}

func (s *ObservabilityWorker) IncDuplicated() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesDuplicated++
}

func (s *ObservabilityWorker) IncReordered() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messagesReordered++
}

func (s *ObservabilityWorker) IncInvariantViolation(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invariantViolation[id]++
}

func (s *ObservabilityWorker) IncWorkerRestart(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workerRestarted[name]++
}

func (s *ObservabilityWorker) UpdateLastActivity(activity time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActive = activity
}

func (s *ObservabilityWorker) UpdateCapacity(name string, cap, len int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channelCapacity[name] = ChannelCapacity{Capacity: cap, Length: len}
}
