package event_bus

import (
	"strings"
	"sync"
)

type Event struct {
	WorkId   string
	Payload  any
	SubTopic string
}

type EventChan chan Event

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventChan
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]EventChan),
	}
}


func (bk *EventBus) Publish(topic string, event Event) {
	bk.mu.RLock()
	subs, ok := bk.subscribers[strings.SplitN(topic, ".", 2)[0]]
	bk.mu.RUnlock()

	if !ok {
		return 
	}
	parts := strings.SplitN(topic, ".", 2)
	newEvent := Event{
		WorkId:  event.WorkId,
		Payload: event.Payload,
	}

	if len(parts) > 1 {
		newEvent.SubTopic = parts[1]
	}

	subscribers := append([]EventChan{}, subs...)

	for _, ch := range subscribers {
		select {
		case ch <- newEvent:
		default:
		}
	}
}
func (bk *EventBus) Subscribe(topic string) EventChan {
	bk.mu.Lock()
	defer bk.mu.Unlock()

	ch := make(EventChan, 64) 
	bk.subscribers[topic] = append(bk.subscribers[topic], ch)

	return ch
}
func (bk *EventBus) Unsubscribe(topic string, ch EventChan) {
	bk.mu.Lock()
	defer bk.mu.Unlock()

	subs, ok := bk.subscribers[topic]
	if !ok {
		return
	}

	for i, sub := range subs {
		if sub == ch {
			bk.subscribers[topic] = append(subs[:i], subs[i+1:]...)

			close(ch)

			return
		}
	}
}
