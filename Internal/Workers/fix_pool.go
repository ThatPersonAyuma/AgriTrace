package workers

import (
	"AgriTrace/Internal/EventBus"
	logistic "AgriTrace/Internal/Core/Logistic"
	"fmt"
)
type Job = func() error
// import (
// 	"AgriTrace/Internal/EventBus"
// 	"fmt"
// )

func handleEffect(e logistic.Effect) error {
    switch e.Type {
    case logistic.EffectDB:
        // contoh DB handler (sesuaikan dengan DB Anda)
        _, err := db.Exec(e.Query, e.Args...)
        return err

    case logistic.EffectLog:
        fmt.Println("[LOG]", e.Msg)
        return nil

    case logistic.EffectNotify:
        fmt.Println("[NOTIFY]", e.Msg)
        return nil

    case logistic.EffectEmail:
        fmt.Println("[EMAIL]", e.Msg)
        return nil

    default:
        return fmt.Errorf("unknown effect type: %v", e.Type)
    }
}

func ListenFixWorks(b *event_bus.EventBus, topic string, workers int) chan []logistic.Effect {
	sub := b.Subscribe(topic)
	jobs := make(chan []logistic.Effect, 50)

	for i := 0; i < workers; i++ {
		go func(id int) {
			for effects := range jobs {
				for _, ef := range effects {
					if err := handleEffect(ef); err != nil {
						fmt.Printf("[FIXED WORKER %d] error: %v\n", id, err)
					}
				}
			}
		}(i)
	}

	go func() {
		for event := range sub {
			jobs <- event.Payload.([]logistic.Effect)
		}
		close(jobs)
	}()

	return jobs
}

// func ListenFixWorks(b *event_bus.EventBus, topic string, workers int) chan func() error{
// 	sub := b.Subscribe(topic)
// 	jobs := make(chan func() error, 50) // Jobs Query

// 	for i:=0;i<workers;i++{
// 		go func(id int) {
//             for job := range jobs {
//                 if err := job(); err != nil {
//                     fmt.Printf("worker %d, error: %s", id, err)
//                 }
//             }
//         }(i)
// 	}

// 	go func(){
// 		for event := range sub {
// 			jobs <- event.Payload.(func() error)
// 		}
// 		close(jobs)
// 	}()
// 	return jobs
// }