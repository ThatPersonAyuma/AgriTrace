package event_bus

import (
	"strings"
	"sync"
)

// Define a payload to store any value
type Event struct {
	WorkId string
	Payload any
	SubTopic string
}
// Define type of EventChan its a chan that hold event datas
type EventChan chan Event

// Define EventBus that hold subcriber
type EventBus struct {
	mu          sync.RWMutex           // reader/writer mutual exclusion lock
	subscribers map[string][]EventChan // Map that holds string as key val and slice of eventChan
}

// func matchTopic(pattern, topic string) bool {
// 	// wildcard global
// 	if pattern == "*" {
// 		return true
// 	}

// 	// wildcard prefix: order.*
// 	if strings.HasSuffix(pattern, ".*") {
// 		prefix := strings.TrimSuffix(pattern, ".*")
// 		return strings.HasPrefix(topic, prefix)
// 	}

// 	// direct match
// 	return pattern == topic
// }

// Define create new EventBus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventChan),
	}
}

// Define Publish method for EventBus, so an event can be published
func (bk *EventBus) Publish(topic string, event Event){
	bk.mu.RLock()
	defer bk.mu.RUnlock()
	parts := strings.SplitN(topic, ".", 2)
	println("Panjang:", len(parts))
	// Copy a new subscriber list to avoid modifying the list while publishing
	subscribers := append([]EventChan{}, bk.subscribers[parts[0]]...)
	var new_event Event
	new_event.WorkId = event.WorkId
	new_event.Payload = event.Payload
	
	if len(parts)>1{
		new_event.SubTopic = parts[1]
	}
	go func() {
		for _, subscriber := range subscribers {
		subscriber <- new_event
		}
	}()
}

// Define a subcribe method that return a eventchan that can be used to listen the publisher/topic
func (bk *EventBus) Subscribe(topic string) EventChan {
	bk.mu.Lock()
	defer bk.mu.Unlock()
	ch := make(EventChan)
	bk.subscribers[topic] = append(bk.subscribers[topic], ch)
	return ch
}

// Define a method to unsubcribe a topic/channel
func (bk *EventBus) Unsubscribe(topic string, ch EventChan) {
	bk.mu.Lock()
	defer bk.mu.Unlock()
	if subscribers, ok := bk.subscribers[topic]; ok {
		for i, subscriber := range subscribers {
		if ch == subscriber {
			bk.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
			close(ch)
			// Drain the channel, ensure the channel closed and the buffered data are drained
			for range ch {
			}
			return
		}
		}
	}
}