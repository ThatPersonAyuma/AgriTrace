package order

import (
	"fmt"
	"AgriTrace/Internal/EventBus"
)

func CreateOrder(b *event_bus.EventBus, orderID int) {
	fmt.Println("[Order] Order created:", orderID)
	// Publish event bahwa order baru dibuat
	b.Publish("OrderCreated", event_bus.Event{orderID})
}

func ListenOrder(b *event_bus.EventBus){
	sub := b.Subscribe("OrderCreated")
	go func(){
		for event := range sub{
			orderID := event.Payload.(int)
			fmt.Println("[Notification] order: ", orderID)
		}
	}()
}
