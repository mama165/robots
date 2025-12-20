package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"time"

	"github.com/hako/durafmt"
)

// QuiescenceDetectorHandler handles events related to robot quiescence.
// It is triggered periodically to report the last activity timestamp of a robot.
// Can be used to detect when all robots have reached a stable state.
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
