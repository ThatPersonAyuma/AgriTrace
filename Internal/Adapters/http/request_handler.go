package http_adapters

import (
	core "AgriTrace/Internal/Core"
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal/Generic"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/schema"
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


func CreateFuncHandler[T any](b *event_bus.EventBus, jobStore *generic.JobStore, method string, topic string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		status := "submitted"

		result := ParseRequestParams[T](r)

		response := map[string]string{
			"job_id": id,
			"status": status,
		}

		if result.Err != nil {
			status = "error"
			response["status"] = status
			response["error"] = result.Err.Error()

			jobStore.Lock()
			jobStore.Data[id] = generic.JobResult{
				Status: "Error",
				Error:  result.Err.Error(),
			}
			jobStore.Unlock()
			fmt.Print(r.Body);
			fmt.Printf("Error on request: %s\n", result.Err)
		} else {
			jobStore.Lock()
			jobStore.Data[id] = generic.JobResult{
				Status: "Processing",
			}
			jobStore.Unlock()

			b.Publish(topic, event_bus.Event{
				WorkId:  id,
				Payload: result.Value,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
func HandleCheckpointPhotoUploadMultipart(b *event_bus.EventBus, jobStore *generic.JobStore, topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form (max 10MB)
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB max
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get checkpoint_id from form
		checkpointIDStr := r.FormValue("checkpoint_id")
		checkpointID, err := strconv.Atoi(checkpointIDStr)
		if err != nil {
			http.Error(w, "Invalid checkpoint_id", http.StatusBadRequest)
			return
		}

		// Get file from form
		file, header, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, "Failed to get file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read file data
		fileData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		// Create request
		req := core.CheckpointPhotoUploadReq{
			CheckpointID: checkpointID,
			Filename:     header.Filename,
			FileData:     fileData,
		}

		// Generate job ID
		jobID := uuid.New().String()

		// Store initial job status
		jobStore.Lock()
		jobStore.Data[jobID] = generic.JobResult{
			Status: "Processing",
		}
		jobStore.Unlock()

		// Publish to event bus
		b.Publish(topic, event_bus.Event{
			WorkId:  jobID,
			Payload: req,
		})

		// Return job ID
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"job_id": jobID,
			"status": "submitted",
		})
	}
}
func HandleServeUploadedFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get filename from URL path
		filename := strings.TrimPrefix(r.URL.Path, "/uploads/")
		if filename == "" {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Prevent directory traversal
		filename = filepath.Clean(filename)
		if strings.Contains(filename, "..") {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}

		filePath := filepath.Join("./uploads", filename)
		http.ServeFile(w, r, filePath)
	}
}