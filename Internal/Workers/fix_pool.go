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
            _, err := db.Exec(e.ExecCommand, e.Args...)
            return generic.Result[any]{Value: nil, Err: err}
        }

    case generic.EffectDBQuery:
        return func() generic.Result[any] {
            rows, err := db.Query(e.ExecCommand, e.Args...)
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
func ListenFixWorks(
    b *event_bus.EventBus,
    topic string,
    workers int,
    db *sql.DB,
    job_store *generic.JobStore,
) chan event_bus.Event {

    sub := b.Subscribe(topic)
    jobs := make(chan event_bus.Event, 50)

    // Forward subscriber → jobs
    go func() {
        defer close(jobs)
        for event := range sub {
            jobs <- event
        }
    }()

    // Launch worker pool
    for i := 0; i < workers; i++ {
        go func(id int) {
            for event := range jobs {

                effects, ok := event.Payload.([]generic.Effect)
                if !ok {
                    job_store.Lock()
                    job_store.Data[event.WorkId] = generic.JobResult{
                        Status: "error",
                        Error:  "invalid effect payload",
                    }
                    job_store.Unlock()
                    continue
                }

                var finalValue any
                var finalErr error

                // Execute effects sequentially
                for _, ef := range effects {
                    result := handleEffect(ef, db)()

                    if result.Err != nil {
                        finalErr = result.Err
                        break
                    }
                    if result.Value!=nil {finalValue = result.Value}
                }

                // Save final result ONCE
                job_store.Lock()
                if finalErr != nil {
                    job_store.Data[event.WorkId] = generic.JobResult{
                        Status: "error",
                        Error:  finalErr.Error(),
                    }
                } else {
                    job_store.Data[event.WorkId] = generic.JobResult{
                        Status: "done",
                        Result: map[string]any{"data": finalValue},
                        Error:  "",
                    }
                }
                job_store.Unlock()
            }
        }(i)
    }

    return jobs
}
