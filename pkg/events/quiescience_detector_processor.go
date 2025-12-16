package events

import (
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"time"

	"github.com/hako/durafmt"
)

type QuiescenceDetectorProcessor struct {
	log *slog.Logger
}

func NewQuiescenceDetectorProcessor(log *slog.Logger) *QuiescenceDetectorProcessor {
	return &QuiescenceDetectorProcessor{log: log}
}

func (p QuiescenceDetectorProcessor) Handle(event Event) {
	switch event.EventType {
	case EventQuiescenceDetector:
		payload, ok := event.Payload.(QuiescenceDetectorEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		elapsed := time.Now().Sub(payload.LastActivity.Date())

		p.log.Debug(fmt.Sprintf("Robot %d last activity was %s ago", payload.RobotID, durafmt.Parse(elapsed)))
	}
}
