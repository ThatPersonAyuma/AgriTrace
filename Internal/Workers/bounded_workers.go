package workers

import (
	"AgriTrace/Internal/EventBus"
)

func ListenWork(b *event_bus.EventBus){
	sub := b.Subscribe("Works")
	go func(){
		for event := range sub{
			work := event.Payload.(func() error)
			work()
		}
	}()
}