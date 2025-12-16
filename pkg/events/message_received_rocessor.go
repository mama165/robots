package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"sync"
)

type MessageReceivedProcessor struct {
	log     *slog.Logger
	mu      sync.Mutex
	counter *Counter
}

func NewMessageReceivedProcessor(log *slog.Logger, counter *Counter) *MessageReceivedProcessor {
	return &MessageReceivedProcessor{log: log, counter: counter}
}

func (p *MessageReceivedProcessor) Handle(event Event) {
	switch event.EventType {
	case EventMessageReceived:
		payload, ok := event.Payload.(MessageReceivedEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		p.counter.Increment(EventMessageReceived)
		p.log.Debug(fmt.Sprintf("Robot %d received a message", payload.ReceiverID))
	}
}
