package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"robots/pkg/errors"
	"robots/pkg/events"
	"sync"
	"time"
)

// Worker doesn't protect itself
// Can be silly, focused
type Worker interface {
	WithName(name string) Worker
	GetName() string
	Run(ctx context.Context) error
}

type ISupervisor interface {
	Run()
	Add(worker ...Worker) ISupervisor
	Start(worker Worker)
	Stop()
}

// Supervisor Own a context and a Cancel function
// Run each worker in a goroutine
// Check panics and errors
// Restart workers automatically
// Shutdown properly if parent context is canceled
// Wait for the end of all goroutines via WaitGroup
type Supervisor struct {
	Ctx     context.Context    // To communicate a stop to all workers
	Cancel  context.CancelFunc // To stop the context
	wg      *sync.WaitGroup    // Wait for the end of goroutines
	log     *slog.Logger
	workers []Worker
	Event   chan events.Event
}

func NewSupervisor(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, log *slog.Logger) Supervisor {
	return Supervisor{Ctx: ctx, Cancel: cancel, wg: wg, log: log}
}

func (s *Supervisor) Run() {
	for _, worker := range s.workers {
		s.Start(worker)
	}
}

func (s *Supervisor) Add(worker ...Worker) ISupervisor {
	s.workers = append(s.workers, worker...)
	return s
}

// Start Imagine une machine à laver (la goroutine superviseur) qui fait tourner des vêtements (les workers).
// Si un vêtement explose (panic),
// Tu l’attrapes sans casser la machine (recover)
// Tu le remets dans la machine (restart Run)
// Et la machine continue son cycle (la goroutine ne s’arrête pas).
func (s *Supervisor) Start(worker Worker) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for {
			if s.Ctx.Err() != nil {
				s.log.Info(fmt.Sprintf("Stopping : %s", worker.GetName()))
				return
			}

			// protection panic directement
			err := func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						s.sendRestartEvent(worker)
						err = errors.ErrWorkerPanic
					}
				}()
				// Execute the children goroutine
				// Restarted after a crash
				// Not restarting the entire goroutine
				return worker.Run(s.Ctx)
			}()

			if err == nil {
				// Terminated properly, never restart !
				s.log.Info(fmt.Sprintf("Worker finished : %s", worker.GetName()))
				return
			}

			s.log.Info(fmt.Sprintf("Restarting : %s", worker.GetName()))
			time.Sleep(200 * time.Millisecond)
		}
	}()
}

// Stop Cancel all goroutines listening channel for Ctx.Done
// Supervisor will wait for all goroutines to finish
func (s *Supervisor) Stop() {
	s.Cancel()
	s.wg.Wait()
}

func (s *Supervisor) sendRestartEvent(worker Worker) {
	select {
	case s.Event <- events.Event{
		EventType: events.EventWorkerRestartedAfterPanic,
		CreatedAt: time.Now().UTC(),
		Payload:   events.WorkerRestartedAfterPanicEvent{WorkerName: worker.GetName()},
	}:
	case <-s.Ctx.Done():
		s.log.Info("Timeout ou Ctrl+C : arrêt de toutes les goroutines")
	default:
		s.log.Warn("Event channel full, WorkerRestartedAfterPanic event dropped")
	}
}
