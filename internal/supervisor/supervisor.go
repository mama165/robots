package supervisor

import (
	"context"
	"fmt"
	"log/slog"
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
	Exec()
	Add(worker Worker) ISupervisor
	Start(worker Worker)
	Stop()
}

// Supervisor Own a context and a cancel function
// Run each worker in a goroutine
// Check panics and errors
// Restart workers automatically
// Shutdown properly if parent context is canceled
// Wait for the end of all goroutines via WaitGroup
type Supervisor struct {
	ctx     context.Context // Stop everything
	cancel  context.CancelFunc
	wg      *sync.WaitGroup // Wait for the end of goroutines
	log     *slog.Logger
	workers []Worker
}

func NewSupervisor(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, log *slog.Logger) Supervisor {
	return Supervisor{ctx: ctx, cancel: cancel, wg: wg, log: log}
}

func (s *Supervisor) Exec() {
	for _, worker := range s.workers {
		s.Start(worker)
	}
}

func (s *Supervisor) Add(worker Worker) ISupervisor {
	s.workers = append(s.workers, worker)
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
			select {
			case <-s.ctx.Done():
				s.log.Info(fmt.Sprintf("Stopping : %s", worker.GetName()))
				return
			default:
			}

			// Protection against panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						s.log.Error(fmt.Sprintf("Recovered panic in %s", worker.GetName()))
					}
				}()

				// Execute the children goroutine
				// Restarted after a crash
				// Not restarting the entire goroutine
				if err := worker.Run(s.ctx); err != nil {
					s.log.Error(fmt.Sprintf("Error %v in %s", err, worker.GetName()))
				}
			}()

			// Petite pause pour éviter les boucles infinies
			time.Sleep(200 * time.Millisecond)

			s.log.Info(fmt.Sprintf("Restarting : %s", worker.GetName()))
		}
	}()
}

func (s *Supervisor) Stop() {
	s.cancel()
	s.wg.Wait()
}
