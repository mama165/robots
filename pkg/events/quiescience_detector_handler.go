package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"time"

	"github.com/hako/durafmt"
)

type QuiescenceDetectorHandler struct {
	log *slog.Logger
}

func NewQuiescenceDetectorHandler(log *slog.Logger) *QuiescenceDetectorHandler {
	return &QuiescenceDetectorHandler{log: log}
}

func (p QuiescenceDetectorHandler) Handle(event Event) {
	switch event.EventType {
	case EventQuiescenceDetector:
		payload, ok := event.Payload.(QuiescenceDetectorEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		elapsed := time.Now().Sub(payload.LastActivity.Date())
		p.log.Debug(fmt.Sprintf("Robot %d last activity was %s ago", payload.ID, durafmt.Parse(elapsed)))
	}
}
