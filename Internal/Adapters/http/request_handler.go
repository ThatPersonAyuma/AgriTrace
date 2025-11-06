package http_adapters

import (
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal/Generic"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/google/uuid"
)

func CreateGetStatusHandler(job_store *generic.JobStore) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.URL.Query().Get("id")
		if id == "" {
            http.Error(w, "Parameter 'id' dibutuhkan", http.StatusBadRequest)
            return
        }
		fmt.Fprintf(w, "ID Anda adalah: %s", id)

		job_store.RLock()
		res, ok := job_store.Data[id]
		job_store.RUnlock()
		
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "job_id": id,
            "result": res,
        })
		// delete(job_store.Data, id)
	}
}

func CreateFuncHandler[T any](b *event_bus.EventBus, job_store *generic.JobStore, method string, topic string)func(http.ResponseWriter, *http.Request){
	return func (w http.ResponseWriter, r *http.Request){
		// Pastikan method POST
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload T
		// Decode JSON body
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		id := uuid.New().String()
		job_store.Data[id] = generic.JobResult{
									Status: "Processing",
								}
		b.Publish(topic, event_bus.Event{WorkId: id, Payload: payload})
		// Create Return Job ID
		w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "job_id": id,
            "status": "submitted",
        })
	}
}