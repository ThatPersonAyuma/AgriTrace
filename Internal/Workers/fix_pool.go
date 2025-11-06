package workers

import (
	"AgriTrace/Internal/EventBus"
	"fmt"
)

func ListenFixWorks(b *event_bus.EventBus, topic string, workers int) chan func() error{
	sub := b.Subscribe(topic)
	jobs := make(chan func() error, 50) // Jobs Query

	for i:=0;i<workers;i++{
		go func(id int) {
            for job := range jobs {
                if err := job(); err != nil {
                    fmt.Printf("worker %d, error: %s", id, err)
                }
            }
        }(i)
	}

	go func(){
		for event := range sub {
			jobs <- event.Payload.(func() error)
		}
		close(jobs)
	}()
	return jobs
}