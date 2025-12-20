package events

import (
	"fmt"
	"io"
	"log/slog"
	"robots/internal/conf"
	"robots/pkg/errors"
	"robots/pkg/robot"
	"sync"
)

type WinnerElectedHandler struct {
	Config conf.Config
	log    *slog.Logger
	Robots []*robot.Robot
	once   *sync.Once
	writer io.Writer
}

func NewWinnerElectedHandler(Config conf.Config, log *slog.Logger,
	robots []*robot.Robot, once *sync.Once,
	writer io.Writer) *WinnerElectedHandler {
	if writer == nil {
		panic("WinnerElectedHandler: writer must be injected")
	}
	return &WinnerElectedHandler{Config: Config, log: log, Robots: robots,
		once: once, writer: writer,
	}
}

func (w *WinnerElectedHandler) Handle(event Event) {
	switch event.EventType {
	case EventWinnerElected:
		payload, ok := event.Payload.(WinnerElectedEvent)
		if !ok {
			w.log.Error(errors.ErrInvalidPayload.Error())
		}
		if payload.ID <= len(w.Robots) {
			w.log.Error(fmt.Sprintf("Robot %d doesn't exist", payload.ID))
			return
		}
		w.writeSecret(w.Robots[payload.ID])
	}
}

// Send the winner in the channel without blocking any other possible winner
func (w *WinnerElectedHandler) writeSecret(r *robot.Robot) {
	w.once.Do(func() {
		if _, err := w.writer.Write([]byte(r.BuildSecret())); err != nil {
			w.log.Error("failed to write secret", "err", err)
		}
		w.log.Info(fmt.Sprintf("Robot %d won and saved the message in file -> %s", r.ID, w.Config.OutputFile))
	})
}
