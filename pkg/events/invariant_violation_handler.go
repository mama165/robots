package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

// InvariantViolationHandler handles detected violations of system invariants.
// It is invoked when a safety or consistency invariant is broken, indicating
// a bug, data corruption, or unexpected behavior. The handler defines how the
// system reacts (e.g. logging, metrics, panic, or recovery).
// It must not attempt to repair state silently or hide invariant violations.
type InvariantViolationHandler struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewInvariantViolationHandler(log *slog.Logger, counter *Counter) *InvariantViolationHandler {
	return &InvariantViolationHandler{log: log, counter: counter}
}

func (p *InvariantViolationHandler) Handle(event Event) {
	switch event.EventType {
	case EventInvariantViolationSameIndexDiffWords:
		_, ok := event.Payload.(InvariantViolationEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.counter.Increment(EventInvariantViolationSameIndexDiffWords)
		p.log.Debug(fmt.Sprintf("Invariant violation %s occurred, total: %d\n", EventInvariantViolationSameIndexDiffWords, p.counter.Get(event.EventType)))
	}
}
