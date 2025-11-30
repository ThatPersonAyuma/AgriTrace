package http_adapters

import (
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal/Generic"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"github.com/gorilla/schema"
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
		// fmt.Fprintf(w, "ID Anda adalah: %s", id)

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

var decoder = schema.NewDecoder()

func ParseRequestParams[T any](r *http.Request) generic.Result[T] {
	var payload T

	// ------------ GET Query Params (decode to struct) ------------
	if r.Method == http.MethodGet {
		if err := decoder.Decode(&payload, r.URL.Query()); err != nil {
			return generic.Result[T]{Err: err}
		}
		return generic.Result[T]{Value: payload}
	}

	// ------------- POST / PUT / PATCH ----------------
	contentType := r.Header.Get("Content-Type")

	switch {

	// ------------ JSON Body ------------
	case strings.Contains(contentType, "application/json"):
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			return generic.Result[T]{Err: errors.New("invalid JSON: " + err.Error())}
		}
		return generic.Result[T]{Value: payload}

	// ------------ x-www-form-urlencoded ------------
	case strings.Contains(contentType, "application/x-www-form-urlencoded"):
		if err := r.ParseForm(); err != nil {
			return generic.Result[T]{Err: err}
		}
        // Form always treated like query
		if err := decoder.Decode(&payload, r.Form); err != nil {
			return generic.Result[T]{Err: err}
		}
		return generic.Result[T]{Value: payload}

	// ------------ multipart/form-data ------------
	case strings.Contains(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			return generic.Result[T]{Err: err}
		}

		// Decode only textual values
		if err := decoder.Decode(&payload, r.MultipartForm.Value); err != nil {
			return generic.Result[T]{Err: err}
		}

		// Optional: handle files (user can add)
		return generic.Result[T]{Value: payload}

	default:
		// Unsupported Content-Type → decode nothing
		return generic.Result[T]{
			Err: errors.New("unsupported content type: " + contentType),
		}
	}
}


func CreateFuncHandler[T any](b *event_bus.EventBus, job_store *generic.JobStore, method string, topic string)func(http.ResponseWriter, *http.Request){
	return func (w http.ResponseWriter, r *http.Request){

		// if r.Method != method {
		// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// 	return
		// }
		// var payload T
		// var payload T
		// // Decode JSON body
		// err := json.NewDecoder(r.Body).Decode(&payload)

		id := uuid.New().String()
		status := "submitted"
		result := ParseRequestParams[T](r)
		if result.Err!=nil{
			fmt.Printf("Error on request: %s", result.Err)
			status = "error"
			job_store.Data[id] = generic.JobResult{
										Status: "Error",
										Error: result.Err.Error(),
									}
		}else{
			job_store.Data[id] = generic.JobResult{
										Status: "Processing",
									}
			b.Publish(topic, event_bus.Event{WorkId: id, Payload: result.Value})
		}

		w.Header().Set("Content-Type", "application/json")
		respone := map[string]string{
            "job_id": id,
            "status": status,
        }
		if status == "error"{
			respone["error"] = result.Err.Error()
		}
        json.NewEncoder(w).Encode(respone)
	}
}