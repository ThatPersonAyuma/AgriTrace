package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"AgriTrace/Internal/Adapters/Http"
	"AgriTrace/Internal/Core"
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal/Generic"
)

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BenchmarkConfig struct {
	NumRequests      int
	PollInterval     time.Duration
	MaxConcurrency   int
}

type JobMetrics struct {
	Latency time.Duration
	Success bool
}

func BenchmarkAsyncLoginSafe(b *testing.B) {
	config := BenchmarkConfig{
		NumRequests:    2000,
		PollInterval:   10 * time.Millisecond,
		MaxConcurrency: 100,
	}

	eventBus := event_bus.NewEventBus()
	jobStore := &generic.JobStore{Data: make(map[string]generic.JobResult)}
	core.ListenLogin(eventBus, "login", "login_worker", jobStore)

	handler := http_adapters.CreateFuncHandler[UserLogin](eventBus, jobStore, "POST", "login")
	statusHandler := http_adapters.CreateGetStatusHandler(jobStore)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/submit":
			handler(w, r)
		case "/status":
			statusHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 50,
			IdleConnTimeout:     10 * time.Second,
		},
	}

	var wg sync.WaitGroup
	var successCount int32
	metrics := make([]JobMetrics, config.NumRequests)
	jobIDs := make([]string, config.NumRequests)
	sem := make(chan struct{}, config.MaxConcurrency)

	// Submit semua job async
	for i := 0; i < config.NumRequests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			start := time.Now()
			login := UserLogin{Username: "admin", Password: "admin123"}
			body, _ := json.Marshal(login)

			req, _ := http.NewRequest("POST", server.URL+"/submit", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				metrics[idx] = JobMetrics{Latency: 0, Success: false}
				return
			}

			var respData map[string]string
			json.NewDecoder(resp.Body).Decode(&respData)
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			jobIDs[idx] = respData["job_id"]
			metrics[idx].Latency = time.Since(start)
		}(i)
	}

	wg.Wait()

	// Batched polling untuk semua job (1 goroutine)
	done := make(chan struct{})
	go func() {
		defer close(done)
		pending := make(map[int]string)
		for i, id := range jobIDs {
			pending[i] = id
		}

		for len(pending) > 0 {
			for idx, jobID := range pending {
				req, _ := http.NewRequest("GET", server.URL+"/status?id="+jobID, nil)
				resp, err := client.Do(req)
				if err != nil {
					continue
				}

				var statusData map[string]any
				json.NewDecoder(resp.Body).Decode(&statusData)
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				if result, ok := statusData["result"].(map[string]any); ok {
					if _, loggedIn := result["logged_in"]; loggedIn {
						atomic.AddInt32(&successCount, 1)
						metrics[idx].Success = true
						delete(pending, idx)
					}
				}
			}
			time.Sleep(config.PollInterval)
		}
	}()

	<-done

	// Hitung rata-rata latency
	var totalLatency time.Duration
	for _, m := range metrics {
		if m.Success {
			totalLatency += m.Latency
		}
	}
	avgLatency := time.Duration(0)
	if successCount > 0 {
		avgLatency = totalLatency / time.Duration(successCount)
	}

	b.Logf("Benchmark selesai: total job sukses %d / %d", successCount, config.NumRequests)
	b.Logf("Rata-rata latency per job: %v", avgLatency)
}


// func BenchmarkAsyncLoginRefactored(b *testing.B) {
// 	config := BenchmarkConfig{
// 		NumRequests:    200,                  // total jobs
// 		PollInterval:   5 * time.Millisecond, // polling interval
// 		MaxConcurrency: 50,                   // max goroutine concurrent
// 	}

// 	// Setup EventBus dan JobStore
// 	eventBus := event_bus.NewEventBus()
// 	jobStore := &generic.JobStore{Data: make(map[string]generic.JobResult)}

// 	// Start login worker
// 	core.ListenLogin(eventBus, "login", "login_worker", jobStore)

// 	// HTTP handlers
// 	handler := http_adapters.CreateFuncHandler[UserLogin](eventBus, jobStore, "POST", "login")
// 	statusHandler := http_adapters.CreateGetStatusHandler(jobStore)

// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		switch r.URL.Path {
// 		case "/submit":
// 			handler(w, r)
// 		case "/status":
// 			statusHandler(w, r)
// 		default:
// 			http.NotFound(w, r)
// 		}
// 	}))
// 	defer server.Close()

// 	// HTTP client dengan transport tuning
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			MaxIdleConns:        500,
// 			MaxIdleConnsPerHost: 500,
// 			IdleConnTimeout:     30 * time.Second,
// 		},
// 	}

// 	var wg sync.WaitGroup
// 	metrics := make([]JobMetrics, config.NumRequests)
// 	var successCount int32

// 	// Semaphore untuk limit concurrency
// 	sem := make(chan struct{}, config.MaxConcurrency)

// 	b.ResetTimer()

// 	// Submit all jobs async dengan concurrency limit
// 	// Submit semua job async dengan concurrency limit
// jobIDs := make([]string, config.NumRequests)
// for i := 0; i < config.NumRequests; i++ {
// 	wg.Add(1)
// 	go func(idx int) {
// 		defer wg.Done()
// 		sem <- struct{}{}
// 		defer func() { <-sem }()

// 		start := time.Now() // simpan waktu start

// 		login := UserLogin{Username: "admin", Password: "admin123"}
// 		body, _ := json.Marshal(login)
// 		req, _ := http.NewRequest("POST", server.URL+"/submit", bytes.NewBuffer(body))
// 		req.Header.Set("Content-Type", "application/json")

// 		resp, err := client.Do(req)
// 		if err != nil {
// 			b.Errorf("[Job %d] Submit error: %v", idx, err)
// 			metrics[idx] = JobMetrics{Latency: 0, Success: false}
// 			return
// 		}
// 		// baca seluruh body untuk menghindari goroutine startBackgroundRead
// 		var respData map[string]string
// 		json.NewDecoder(resp.Body).Decode(&respData)
// 		resp.Body.Close()

// 		jobID := respData["job_id"]
// 		jobIDs[idx] = jobID

// 		// Per-job polling sampai Done
// 		for {
// 			reqStatus, _ := http.NewRequest("GET", server.URL+"/status?id="+jobID, nil)
// 			respStatus, err := client.Do(reqStatus)
// 			if err != nil {
// 				b.Errorf("[Job %d] Status error: %v", idx, err)
// 				return
// 			}

// 			var statusData map[string]any
// 			json.NewDecoder(respStatus.Body).Decode(&statusData)
// 			io.Copy(io.Discard, respStatus.Body) // pastikan semua dibaca
// 			respStatus.Body.Close()

// 			result, ok := statusData["result"].(map[string]any)
// 			if ok {
// 				if _, loggedIn := result["logged_in"]; loggedIn {
// 					atomic.AddInt32(&successCount, 1)
// 					metrics[idx] = JobMetrics{
// 						Latency: time.Since(start),
// 						Success: true,
// 					}
// 					break
// 				}
// 			}
// 			time.Sleep(10 * time.Millisecond) // polling interval aman
// 		}
// 	}(i)
// }

// wg.Wait() // tunggu semua job selesai


// 	// Batched polling: cek semua job sampai Done
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		pending := make(map[int]string)
// 		for i, id := range jobIDs {
// 			pending[i] = id
// 		}

// 		for len(pending) > 0 {
// 			for idx, jobID := range pending {
// 				reqStatus, _ := http.NewRequest("GET", server.URL+"/status?id="+jobID, nil)
// 				respStatus, err := client.Do(reqStatus)
// 				if err != nil {
// 					b.Errorf("[Job %d] Status error: %v", idx, err)
// 					continue
// 				}
// 				var statusData map[string]any
// 				json.NewDecoder(respStatus.Body).Decode(&statusData)
// 				respStatus.Body.Close()

// 				result, ok := statusData["result"].(map[string]any)
// 				if ok {
// 					if _, loggedIn := result["logged_in"]; loggedIn {
// 						atomic.AddInt32(&successCount, 1)
// 						metrics[idx].Success = true
// 						metrics[idx].Latency = metrics[idx].Latency
// 						delete(pending, idx)
// 					}
// 				}
// 			}
// 			time.Sleep(config.PollInterval)
// 		}
// 	}()

// 	wg.Wait()

// 	// Hitung average latency
// 	var totalLatency time.Duration
// 	for _, m := range metrics {
// 		if m.Success {
// 			totalLatency += m.Latency
// 		}
// 	}
// 	avgLatency := time.Duration(0)
// 	if successCount > 0 {
// 		avgLatency = totalLatency / time.Duration(successCount)
// 	}

// 	b.Logf("Benchmark selesai: total job sukses %d / %d", successCount, config.NumRequests)
// 	b.Logf("Rata-rata latency per job: %v", avgLatency)
// }
