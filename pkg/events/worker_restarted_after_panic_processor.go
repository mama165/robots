package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

type WorkerRestartedAfterPanicProcessor struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewWorkerRestartedAfterPanicProcessor(log *slog.Logger, counter *Counter) *WorkerRestartedAfterPanicProcessor {
	return &WorkerRestartedAfterPanicProcessor{
		log:     log,
		counter: counter,
	}
}

func (p *WorkerRestartedAfterPanicProcessor) Handle(event Event) {
	switch event.EventType {
	case EventWorkerRestartedAfterPanic:
		payload, ok := event.Payload.(WorkerRestartedAfterPanicEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
			return
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.counter.Increment(EventWorkerRestartedAfterPanic)
		p.log.Debug(fmt.Sprintf("Worker %s restarted after panic, total: %d", payload.WorkerName, p.counter.Get(EventWorkerRestartedAfterPanic)))
	}
}
