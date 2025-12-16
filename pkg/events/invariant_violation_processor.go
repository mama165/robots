package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

type InvariantViolationProcessor struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewInvariantViolationProcessor(log *slog.Logger, counter *Counter) *InvariantViolationProcessor {
	return &InvariantViolationProcessor{log: log, counter: counter}
}

func (p *InvariantViolationProcessor) Handle(event Event) {
	switch event.EventType {
	case EventInvariantViolationSecretPartDecreased, EventInvariantViolationSameIndexDiffWords:
		_, ok := event.Payload.(InvariantViolationEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.counter.Increment(event.EventType)
		p.log.Debug(fmt.Sprintf("Invariant violation %s occurred, total: %d\n", event.EventType, p.counter.Get(event.EventType)))
	}
}
