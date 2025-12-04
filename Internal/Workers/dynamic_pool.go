package workers

import (
	event_bus "AgriTrace/Internal/EventBus"
	generic "AgriTrace/Internal/Generic"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

func ListenDynWork(b *event_bus.EventBus, topic string, minWorkers, maxWorkers int, db *sql.DB, job_result *generic.JobStore) chan event_bus.Event {
	sub := b.Subscribe(topic)
	jobs := make(chan event_bus.Event, 50)

	var mu sync.Mutex
	active := 0

	// spawn initial workers
	for active < minWorkers {
		go dynWorker(jobs, db, job_result)
		active++
	}

	// dispatcher
	go func() {
		for event := range sub {

			mu.Lock()
			canSpawn := active < maxWorkers
			mu.Unlock()

			select {
			case jobs <- event:
				// OK, queued
			default:
				// buffer penuh → spawn worker baru kalau bisa
				if canSpawn {
					go dynWorker(jobs, db, job_result)

					mu.Lock()
					active++
					mu.Unlock()

					jobs <- event
				} else {
					fmt.Println("[DYN] OVERLOAD: MAX WORKER REACHED")
				}
			}
		}

		// ketika topic closed → tutup jobs agar worker berhenti
		close(jobs)

	}()

	return jobs
}

// worker with auto-shutdown after 30s idle
func dynWorker(jobs <-chan event_bus.Event, db *sql.DB, job_store *generic.JobStore) {
	idleTimer := time.NewTimer(30 * time.Second)

	for {
		select {
		case event, ok := <-jobs:
			if !ok {
				// channel closed → worker mati
				return
			}
			idleTimer.Reset(30 * time.Second)

			effects, ok := event.Payload.([]generic.Effect)
			if !ok {
				job_store.Lock()
				job_store.Data[event.WorkId] = generic.JobResult{
					Status: "error",
					Error:  "invalid effect payload",
				}
				job_store.Unlock()
				continue
			}

			var finalResult any
			var finalErr error

			for _, ef := range effects {
				result := handleEffect(ef, db)()
				if result.Err != nil {
					finalErr = result.Err
					break
				}
				finalResult = result.Value
			}

			job_store.Lock()
			if finalErr != nil {
				job_store.Data[event.WorkId] = generic.JobResult{
					Status: "error",
					Error:  finalErr.Error(),
				}
			} else {
				job_store.Data[event.WorkId] = generic.JobResult{
					Status: "done",
					Result: map[string]any{
						"data": finalResult,
					},
				}
			}
			job_store.Unlock()

		case <-idleTimer.C:
			// auto shutdown
			return
		}
	}
}
