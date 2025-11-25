// package workers

// import (
// 	"AgriTrace/Internal/EventBus"
// 	generic "AgriTrace/Internal/Generic"
// 	"database/sql"
// 	"fmt"
// )

// func ListenDynWork(b *event_bus.EventBus, topic string, minWorkers, maxWorkers int, db *sql.DB) chan []generic.Effect {
// 	sub := b.Subscribe(topic)
// 	jobs := make(chan []generic.Effect, 50)

// 	// Initial workers
// 	active := 0
// 	for active < minWorkers {
// 		go dynWorker(jobs, db)
// 		active++
// 	}

// 	// dispatcher
// 	go func(active, max int) {
// 		for event := range sub {
// 			effects := event.Payload.([]generic.Effect)

// 			select {
// 			case jobs <- effects:
// 				// OK
// 			default:
// 				if active < max {
// 					go dynWorker(jobs, db)
// 					active++
// 					jobs <- effects
// 				} else {
// 					fmt.Println("[DYN] OVERLOAD: MAX WORKER REACHED")
// 				}
// 			}
// 		}
// 	}(active, maxWorkers)

// 	return jobs
// }

// func dynWorker(jobs <-chan []generic.Effect, db *sql.DB) {
// 	for effects := range jobs {
// 		for _, e := range effects {
// 			if err := handleEffect(e, db); err != nil {
// 				fmt.Println("[DYNAMIC WORKER] error:", err)
// 			}
// 			fmt.Println("Running work, work id:")
// 		}
// 	}
// }




package workers

import (
	"AgriTrace/Internal/EventBus"
	generic "AgriTrace/Internal/Generic"
	"database/sql"
	"fmt"
)

func ListenDynWork(b *event_bus.EventBus, topic string, minWorkers, maxWorkers int, db *sql.DB) chan event_bus.Event {
	sub := b.Subscribe(topic)
	jobs := make(chan event_bus.Event, 50)

	// Initial workers
	active := 0
	for active < minWorkers {
		go dynWorker(jobs, db)
		active++
	}

	// dispatcher
	go func(active, max int) {
		for event := range sub {
			select {
			case jobs <- event:
				// OK
			default:
				if active < max {
					go dynWorker(jobs, db)
					active++
					jobs <- event
				} else {
					fmt.Println("[DYN] OVERLOAD: MAX WORKER REACHED")
				}
			}
		}
	}(active, maxWorkers)

	return jobs
}

func dynWorker(jobs <-chan event_bus.Event, db *sql.DB) {
	for event := range jobs {
		payload, ok := event.Payload.([]generic.Effect)
		if ok {
			for _, e := range payload {
				if err := handleEffect(e, db)(); err != nil {
					fmt.Println("[DYNAMIC WORKER] error:", err)
				}
				fmt.Println("Running work, work id:", event.WorkId)
			}
		}else{
			fmt.Println("Fatal Error, In DynWorkers")
		}
	}
}

// func ListenDynWork(b *event_bus.EventBus, topic string, min_workers, max_workers int) chan func() error{
// 	sub := b.Subscribe(topic)
// 	jobs := make(chan func() error, 50) // Jobs Query
// 	i:=0
// 	for i<min_workers{
// 		go worker(jobs)
// 		i++
// 	}
// 	go func(count, max_count int){
// 		for event := range sub{
// 			job := event.Payload.(func() error)
// 			select {
// 			case jobs <- job:
// 			default:
// 				if count < max_count{
// 					go worker(jobs)
// 					go func() { jobs <- job }()
// 				}else{
// 					fmt.Printf("Fatal Too Much Request")
// 				}
// 			}
// 		}
// 	}(i, max_workers)

// 	return jobs
// }

// func worker(jobs <- chan func() error){ // receiver only notatiom. Only read from channel
// 	for job := range jobs {
// 		if err := job(); err != nil {
// 			fmt.Printf("worker Dyn, error: %s", err)
// 		}
// 	}
// }