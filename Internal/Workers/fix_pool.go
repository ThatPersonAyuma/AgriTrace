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

func handleEffect(e generic.Effect, db *sql.DB) func() generic.Result[any] {
    switch e.Type {

    case generic.EffectDB:
        return func() generic.Result[any] {
            _, err := db.Exec(e.EcexCommand, e.Args...)
            return generic.Result[any]{Value: nil, Err: err}
        }

    case generic.EffectDBQuery:
        return func() generic.Result[any] {
            rows, err := db.Query(e.EcexCommand, e.Args...)
            if err != nil {
                return generic.Result[any]{Err: err}
            }

            // Convert rows → []map[string]any
            result, err := RowsToMaps(rows)
            return generic.Result[any]{Value: result, Err: err}
        }

    case generic.EffectLog:
        return func() generic.Result[any] {
            fmt.Println("[LOG]", e.Msg)
            return generic.Result[any]{Value: nil, Err: nil}
        }

    case generic.EffectComplex:
        return func() generic.Result[any] {
            r := e.Fn()
            return generic.Result[any]{Value: r.Value, Err: r.Err}
        }

    default:
        return func() generic.Result[any] {
            return generic.Result[any]{Err: fmt.Errorf("unknown effect")}
        }
    }
}

func RowsToMaps(rows *sql.Rows) ([]map[string]any, error) {
    defer rows.Close()

    cols, _ := rows.Columns()
    var results []map[string]any

    for rows.Next() {
        vals := make([]any, len(cols))
        ptrs := make([]any, len(cols))
        for i := range vals {
            ptrs[i] = &vals[i]
        }

        if err := rows.Scan(ptrs...); err != nil {
            return nil, err
        }

        rowMap := map[string]any{}
        for i, col := range cols {
            rowMap[col] = vals[i]
        }

        results = append(results, rowMap)
    }

    return results, nil
}

func ListenFixWorks(b *event_bus.EventBus, topic string, workers int, db *sql.DB, job_store *generic.JobStore) chan event_bus.Event {
	sub := b.Subscribe(topic)
	jobs := make(chan event_bus.Event, 50)
	
	go func() {
		for event := range sub {
			jobs <- event
		}
	}()

	for i := 0; i < workers; i++ {
		go func(id int, db *sql.DB) {
			for event := range jobs {
				// fmt.Printf("Fix Workers, id: %s", event.WorkId)

				effects, ok := event.Payload.([]generic.Effect)
				if !ok {
					// Jika payload bukan efek, tandai error saja
					job_store.Lock()
					job_store.Data[event.WorkId] = generic.JobResult{
						Status: "error",
						Error:  "invalid effect payload",
					}
					job_store.Unlock()
					continue
				}

				var finalResult any
				var finalErr error

				// Jalankan semua efek
				for _, ef := range effects {
					result := handleEffect(ef, db)()
					if result.Err != nil {
						finalErr = result.Err
						break // langsung stop jika error
					}
					finalResult = result.Value
				}

				// Simpan hasil (1x saja!)
				job_store.Lock()
				if finalErr != nil {
					job_store.Data[event.WorkId] = generic.JobResult{
						Status: "error",
						Error:  finalErr.Error(),
					}
				} else {
					job_store.Data[event.WorkId] = generic.JobResult{
						Status: "done",
						Result: map[string]any{
							"data": finalResult,
						},
						Error: "",
					}
				}
				job_store.Unlock()
			}
		}(i, db)
	}
	return jobs
}



// func ListenFixWorks(b *event_bus.EventBus, topic string, workers int, db *sql.DB, job_store *generic.JobStore) chan event_bus.Event {
// 	sub := b.Subscribe(topic)
// 	jobs := make(chan event_bus.Event, 50)

// 	for i := 0; i < workers; i++ {
// 		go func(id int, db *sql.DB) {
// 			for event := range jobs {
// 				fmt.Printf("Fix Workers, id: %s", event.WorkId)
// 				effects, is_effect :=  event.Payload.([]generic.Effect)
// 				if (!is_effect){

// 				}else{
// 					for _, ef := range effects {
// 						result := handleEffect(ef, db)()

// 						job_store.Lock()
// 						job_store.Data[event.WorkId] = generic.JobResult{
// 							Status: "done",
// 							Result: map[string]any{
// 								"data": result.Value,
// 							},
// 							Error: "",
// 						}
// 						job_store.Unlock()

// 						if result.Err != nil {
// 							job_store.Data[event.WorkId] = generic.JobResult{
// 								Status: "error",
// 								Error:  result.Err.Error(),
// 							}
// 						}else{

// 						}
// 					}
// 				}
// 			}
// 		}(i, db)
// 	}

// 	go func() {
// 		for event := range sub {
// 			jobs <- event
// 		}
// 		close(jobs)
// 	}()

// 	return jobs
// }

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