package event_bus

import (
	// "log"
	// "strings"
	"strings"
	"sync"
)

// Define a payload to store any value
type Event struct {
	WorkId string
	Payload any
	SubTopic string
}
// type EventChan chan Event

// type topicRuntime struct {
//     queue     chan Event
//     semaphore chan struct{}
//     workers   int
// }

// type EventBus struct {
//     mu          sync.RWMutex
//     subscribers map[string][]EventChan
//     topics      map[string]*topicRuntime
//     queueSize   int
//     workerCount int
//     maxPush     int
// }

// func NewEventBus(queueSize, workerCount, maxPush int) *EventBus {
//     return &EventBus{
//         subscribers: make(map[string][]EventChan),
//         topics:      make(map[string]*topicRuntime),
//         queueSize:   queueSize,
//         workerCount: workerCount,
//         maxPush:     maxPush,
//     }
// }

// // =====================================================
// // Subscribe
// // =====================================================
// func (bus *EventBus) Subscribe(topic string) EventChan {
//     bus.mu.Lock()
//     defer bus.mu.Unlock()

//     ch := make(EventChan, 20)
//     bus.subscribers[topic] = append(bus.subscribers[topic], ch)

//     // Init topic workers (only once)
//     if _, ok := bus.topics[topic]; !ok {
//         rt := &topicRuntime{
//             queue:     make(chan Event, bus.queueSize),
//             semaphore: make(chan struct{}, bus.maxPush),
//             workers:   bus.workerCount,
//         }
//         bus.topics[topic] = rt

//         // Start workers
//         for i := 0; i < bus.workerCount; i++ {
//             go bus.workerLoop(topic, rt)
//         }
//     }
//     return ch
// }

// // =====================================================
// // Worker Loop
// // =====================================================
// func (bus *EventBus) workerLoop(topic string, rt *topicRuntime) {
//     for event := range rt.queue {

//         bus.mu.RLock()
//         subs := append([]EventChan{}, bus.subscribers[topic]...)
//         sem := rt.semaphore
//         bus.mu.RUnlock()

//         for _, sub := range subs {
//             sem <- struct{}{} // acquire

//             go func(s EventChan, ev Event) {
//                 s <- ev
//                 <-sem // release
//             }(sub, event)
//         }
//     }
// }

// // =====================================================
// // Publish
// // =====================================================
// func (bus *EventBus) Publish(topic string, event Event) {
// 	parts := strings.SplitN(topic, ".", 2)
// 	if len(parts)>1{
// 		event.SubTopic = parts[1]
// 	}
//     bus.mu.RLock()
//     rt := bus.topics[parts[0]]
//     bus.mu.RUnlock()

//     if rt == nil {
//         return // no subscribers → no work
//     }

//     rt.queue <- event // FAST push
// }

// // =====================================================
// // Unsubscribe
// // =====================================================
// func (bus *EventBus) Unsubscribe(topic string, ch EventChan) {
//     bus.mu.Lock()
//     defer bus.mu.Unlock()

//     subs := bus.subscribers[topic]
//     for i, sub := range subs {
//         if sub == ch {
//             close(sub)
//             bus.subscribers[topic] = append(subs[:i], subs[i+1:]...)
//             return
//         }
//     }
// }
// Define type of EventChan its a chan that hold event datas
type EventChan chan Event

// Define EventBus that hold subcriber
type EventBus struct {
	mu          sync.RWMutex           // reader/writer mutual exclusion lock
	subscribers map[string][]EventChan // Map that holds string as key val and slice of eventChan
}
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
	// Copy a new subscriber list to avoid modifying the list while publishing
	subscribers := append([]EventChan{}, bk.subscribers[parts[0]]...)
	var new_event Event
	new_event.WorkId = event.WorkId
	new_event.Payload = event.Payload
	
	if len(parts)>1{
		new_event.SubTopic = parts[1]
	}
	for _, subscriber := range subscribers {
		// select {
		go func() {
			subscriber <- new_event
		}()
	}
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