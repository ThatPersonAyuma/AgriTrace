package workers

import (
	"AgriTrace/Internal/EventBus"
	generic "AgriTrace/Internal/Generic"
	"database/sql"
	"fmt"
)
type Job = func() error
// import (
// 	"AgriTrace/Internal/EventBus"
// 	"fmt"
// )

func handleEffect(e generic.Effect, db *sql.DB) func() error {
    switch e.Type {
    case generic.EffectDB:
        // contoh DB handler (sesuaikan dengan DB Anda)
        return func() error {_, err := db.Exec(e.EcexCommand, e.Args...)
        return err}

	// case generic.EffectDBQuerry:
    //     // contoh DB handler (sesuaikan dengan DB Anda)
    //     return func() error {_, err := db.Query(e.EcexCommand, e.Args...)
    //     return err}

    case generic.EffectLog:
        return func() error {fmt.Println("[LOG]", e.Msg)
        return nil}

    case generic.EffectNotify:
        return func() error{fmt.Println("[NOTIFY]", e.Msg)
        return nil}

    case generic.EffectEmail:
        return func() error {fmt.Println("[EMAIL]", e.Msg)
        return nil}

	case generic.EffectComplex:
		return e.Fn

    default:
        return func() error {return fmt.Errorf("unknown effect type: %v", e.Type)}
    }
}

func ListenFixWorks(b *event_bus.EventBus, topic string, workers int, db *sql.DB) chan []generic.Effect {
	sub := b.Subscribe(topic)
	jobs := make(chan []generic.Effect, 50)

	for i := 0; i < workers; i++ {
		go func(id int, db *sql.DB) {
			for effects := range jobs {
				for _, ef := range effects {
					if err := handleEffect(ef, db)(); err != nil {
						fmt.Printf("[FIXED WORKER %d] error: %v\n", id, err)
					}
				}
			}
		}(i, db)
	}

	go func() {
		for event := range sub {
			jobs <- event.Payload.([]generic.Effect)
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