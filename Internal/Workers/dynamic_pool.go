package workers

import (
	"AgriTrace/Internal/EventBus"
	"fmt"
	logistic "AgriTrace/Internal/Core/Logistic"
)

func ListenDynWork(b *event_bus.EventBus, topic string, minWorkers, maxWorkers int) chan []logistic.Effect {
	sub := b.Subscribe(topic)
	jobs := make(chan []logistic.Effect, 50)

	// Initial workers
	active := 0
	for active < minWorkers {
		go dynWorker(jobs)
		active++
	}

	// dispatcher
	go func(active, max int) {
		for event := range sub {
			effects := event.Payload.([]logistic.Effect)

			select {
			case jobs <- effects:
				// OK
			default:
				if active < max {
					go dynWorker(jobs)
					active++
					jobs <- effects
				} else {
					fmt.Println("[DYN] OVERLOAD: MAX WORKER REACHED")
				}
			}
		}
	}(active, maxWorkers)

	return jobs
}

func dynWorker(jobs <-chan []logistic.Effect) {
	for effects := range jobs {
		for _, e := range effects {
			if err := handleEffect(e); err != nil {
				fmt.Println("[DYNAMIC WORKER] error:", err)
			}
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