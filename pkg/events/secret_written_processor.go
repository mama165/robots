package events

import (
	"fmt"
	"log/slog"
	"os"
	"robots/internal/conf"
	"robots/internal/robot"
	"robots/pkg/errors"
)

type SecretWrittenProcessor struct {
	config conf.Config
	log    *slog.Logger
	winner chan robot.Robot
}

func NewSecretWrittenProcessor(config conf.Config, log *slog.Logger, winner chan robot.Robot) *SecretWrittenProcessor {
	return &SecretWrittenProcessor{config: config, log: log, winner: winner}
}

func (p SecretWrittenProcessor) Handle(event Event) {
	switch event.EventType {
	case EventSecretWritten:
		payload, ok := event.Payload.(SecretWrittenEvent)
		if !ok {
			p.log.Error(errors.ErrInvalidPayload.Error())
		}
		file, err := os.Create(p.config.OutputFile)
		if err != nil {
			p.log.Error(err.Error())
		}
		defer file.Close()
		if _, err = file.WriteString(payload.Robot.BuildSecret()); err != nil {
			p.log.Error(err.Error())
		}
		p.log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", payload.Robot.ID, p.config.OutputFile))
	}
}
