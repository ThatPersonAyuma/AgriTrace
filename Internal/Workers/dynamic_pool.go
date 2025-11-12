package workers

import (
	"AgriTrace/Internal/EventBus"
	"fmt"
)

func ListenDynWork(b *event_bus.EventBus, topic string, min_workers, max_workers int) chan func() error{
	sub := b.Subscribe(topic)
	jobs := make(chan func() error, 50) // Jobs Query
	i:=0
	for i<min_workers{
		go worker(jobs)
		i++
	}
	go func(count, max_count int){
		for event := range sub{
			job := event.Payload.(func() error)
			select {
			case jobs <- job:
			default:
				if count < max_count{
					go worker(jobs)
					go func() { jobs <- job }()
				}else{
					fmt.Printf("Fatal Too Much Request")
				}
			}
		}
	}(i, max_workers)

	return jobs
}

func worker(jobs <- chan func() error){ // receiver only notatiom. Only read from channel
	for job := range jobs {
		if err := job(); err != nil {
			fmt.Printf("worker Dyn, error: %s", err)
		}
	}
}