package supervisor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ISupervisor interface {
	Add(worker Worker) ISupervisor
	Start(name string, worker Worker)
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
	wg      sync.WaitGroup // Wait for the end of goroutines
	workers []Worker
}

func NewSupervisor(ctx context.Context) Supervisor {
	return Supervisor{ctx: ctx}
}

func (s *Supervisor) Add(worker Worker) ISupervisor {
	s.workers = append(s.workers, worker)
	return s
}

func (s *Supervisor) Start(name string, worker Worker) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.ctx.Done():
				fmt.Println("Stopping:", name)
				return
			default:
			}

			// Protection contre les panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Recovered panic in", name, ":", r)
					}
				}()

				// Exécute la goroutine enfant
				err := worker.Run(s.ctx)
				if err != nil {
					fmt.Println("Error in", name, ":", err)
				}
			}()

			// Petite pause pour éviter les boucles infinies
			time.Sleep(200 * time.Millisecond)

			fmt.Println("Restarting:", name)
		}
	}()
}

func (s *Supervisor) Stop() {
	s.cancel()
	s.wg.Wait()
}

type Worker interface {
	Run(ctx context.Context) error
}
