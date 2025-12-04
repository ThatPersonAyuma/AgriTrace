package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// ===============================================
// CONFIG
// ===============================================

type TestConfig struct {
	TargetURL     string        // endpoint utama yang mengembalikan job_id
	CheckURL      string        // endpoint polling: /work/check
	Payload       any           // payload request
	Workers       int           // jumlah worker paralel
	Interval      time.Duration // interval antar request
	PollInterval  time.Duration // interval pengecekan status job
	PollTimeout   time.Duration // maksimal waktu polling
	Duration      time.Duration // durasi load test
}

// Response dari endpoint utama
type JobSubmitResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// Response dari /work/check
type JobCheckResponse struct {
	JobID  string `json:"job_id"`
	Result struct {
		Status string `json:"status"`
		Result any    `json:"result"`
		Error  string `json:"error"`
	} `json:"result"`
}

// ===============================================
// POLLING SYSTEM
// ===============================================

func pollJob(jobID string, cfg TestConfig) {
	client := &http.Client{Timeout: 5 * time.Second}

	timeout := time.After(cfg.PollTimeout)

	for {
		select {
		case <-timeout:
			fmt.Printf("[POLL] Job %s -> TIMEOUT after %v\n", jobID, cfg.PollTimeout)
			return

		default:
			url := fmt.Sprintf("%s?id=%s", cfg.CheckURL, jobID)

			resp, err := client.Get(url)
			if err != nil {
				fmt.Printf("[POLL] Request error for job %s: %v\n", jobID, err)
				time.Sleep(cfg.PollInterval)
				continue
			}

			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			var check JobCheckResponse
			json.Unmarshal(body, &check)

			// response valid?
			if check.Result.Status == "done" {
				fmt.Printf("[POLL] Job %s -> DONE ✓\nResult: %v\n", jobID, check.Result.Result)
				return
			}

			if check.Result.Status == "error" {
				fmt.Printf("[POLL] Job %s -> ERROR: %s\n", jobID, check.Result.Error)
				return
			}

			// if still submitted | processing
			fmt.Printf("[POLL] Job %s -> %s ...\n", jobID, check.Result.Status)
			time.Sleep(cfg.PollInterval)
		}
	}
}

// ===============================================
// WORKER
// ===============================================

func startWorker(id int, cfg TestConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()

	for time.Since(start) < cfg.Duration {

		body, _ := json.Marshal(cfg.Payload)

		resp, err := client.Post(cfg.TargetURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("[Worker %d] ERROR sending request: %v\n", id, err)
			time.Sleep(cfg.Interval)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var submit JobSubmitResponse
		json.Unmarshal(respBody, &submit)

		if submit.JobID == "" {
			fmt.Printf("[Worker %d] INVALID RESPONSE: %s\n", id, string(respBody))
			time.Sleep(cfg.Interval)
			continue
		}

		fmt.Printf("[Worker %d] Job submitted: %s\n", id, submit.JobID)

		// Start polling in a separate goroutine
		go pollJob(submit.JobID, cfg)

		time.Sleep(cfg.Interval)
	}
}

// ===============================================
// MAIN
// ===============================================

func main() {

	cfg := TestConfig{
		TargetURL:    "http://localhost:8080/product/listed", // endpoint utama
		CheckURL:     "http://localhost:8080/work/check",     // endpoint polling
		Payload:      map[string]any{},                       // payload bebas
		Workers:      2,
		Interval:     1 * time.Second,
		PollInterval: 1000 * time.Millisecond,
		PollTimeout:  15 * time.Second,
		Duration:     20 * time.Second,
	}

	fmt.Println("\n===== LOAD TEST WITH JOB POLLING STARTED =====")
	fmt.Printf("TargetURL: %s\n", cfg.TargetURL)
	fmt.Printf("CheckURL:  %s\n", cfg.CheckURL)
	fmt.Printf("Workers:   %d\n", cfg.Workers)
	fmt.Printf("Duration:  %v\n", cfg.Duration)
	fmt.Println("==============================================\n")

	var wg sync.WaitGroup

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go startWorker(i, cfg, &wg)
	}

	wg.Wait()

	fmt.Println("\n===== LOAD TEST FINISHED =====")
}
