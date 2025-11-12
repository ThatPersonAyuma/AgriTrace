package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	Latency time.Duration
	OK      bool
}

func main() {
	var (
		targetURL   = flag.String("url", "http://localhost:8080/login", "login URL (POST)")
		statusURL   = flag.String("status-url", "http://localhost:8080/check-work", "status URL (GET)")
		bodyFile    = flag.String("body", "", "path to JSON body file (optional)")
		concurrency = flag.Int("c", 10, "concurrency (goroutines)")
		total       = flag.Int("n", 100, "total requests")
		timeout     = flag.Duration("timeout", 5*time.Second, "http client timeout")
		poll        = flag.Bool("poll", true, "poll job status after submit")
	)
	flag.Parse()

	var bodyBytes []byte
	var err error
	if *bodyFile != "" {
		bodyBytes, err = os.ReadFile(*bodyFile)
		if err != nil {
			fmt.Printf("failed to read body file: %v\n", err)
			return
		}
	} else {
		body := map[string]any{"username": "admin", "password": "admin123"}
		bodyBytes, _ = json.Marshal(body)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 1000,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: *timeout,
	}

	jobs := make(chan int, *total)
	results := make(chan Result, *total)

	var sent int32
	var errs int32

	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				start := time.Now()

				req, _ := http.NewRequest(http.MethodPost, *targetURL, bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt32(&errs, 1)
					results <- Result{Latency: 0, OK: false}
					continue
				}

				var respData map[string]string
				json.NewDecoder(resp.Body).Decode(&respData)
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				ok := resp.StatusCode >= 200 && resp.StatusCode < 300
				success := false

				if *poll && ok {
					jobID, exists := respData["job_id"]
					if exists && jobID != "" {
						for {
							statusReq, _ := http.NewRequest(http.MethodGet, *statusURL+"?id="+jobID, nil)
							statusResp, err := client.Do(statusReq)
							if err != nil {
								break
							}
							var statusData struct {
								JobID  string                 `json:"job_id"`
								Result map[string]any         `json:"result"`
								Status string                 `json:"status"`
								Error  string                 `json:"error,omitempty"`
							}
							json.NewDecoder(statusResp.Body).Decode(&statusData)
							io.Copy(io.Discard, statusResp.Body)
							statusResp.Body.Close()

							if statusData.Status == "Done" {
								success = true
								break
							}
							time.Sleep(10 * time.Millisecond)
						}
					}
				} else {
					success = ok
				}

				results <- Result{Latency: time.Since(start), OK: success}
				atomic.AddInt32(&sent, 1)
			}
		}()
	}

	startAll := time.Now()
	go func() {
		for i := 0; i < *total; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	var latencies []time.Duration
	var successCount int
	for i := 0; i < *total; i++ {
		r := <-results
		if r.OK {
			latencies = append(latencies, r.Latency)
			successCount++
		}
	}
	wg.Wait()
	totalTime := time.Since(startAll)

	var sum time.Duration
	for _, l := range latencies {
		sum += l
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	p := func(q float64) time.Duration {
		if len(latencies) == 0 {
			return 0
		}
		idx := int(float64(len(latencies)-1) * q)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(latencies) {
			idx = len(latencies) - 1
		}
		return latencies[idx]
	}

	fmt.Println("======== LOAD TEST RESULT ========")
	fmt.Printf("Target            : %s\n", *targetURL)
	fmt.Printf("Total requests    : %d\n", *total)
	fmt.Printf("Concurrency       : %d\n", *concurrency)
	fmt.Printf("Sent              : %d\n", atomic.LoadInt32(&sent))
	fmt.Printf("Success responses : %d\n", successCount)
	fmt.Printf("Errors            : %d\n", atomic.LoadInt32(&errs))
	fmt.Printf("Total time        : %v\n", totalTime)
	if successCount > 0 {
		fmt.Printf("Avg latency       : %v\n", sum/time.Duration(successCount))
		fmt.Printf("p50               : %v\n", p(0.50))
		fmt.Printf("p90               : %v\n", p(0.90))
		fmt.Printf("p99               : %v\n", p(0.99))
	}
}
