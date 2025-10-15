package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"modular/broker"
	"modular/modules/order"
)

type user_order struct{
	OrderId int `json:"OrderId"`
}

func createOrderHandler(b *event_bus.EventBus)func(http.ResponseWriter, *http.Request){
	return func (w http.ResponseWriter, r *http.Request){
		// Pastikan method POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Decode JSON body
		var userOrder user_order
		err := json.NewDecoder(r.Body).Decode(&userOrder)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
		order.CreateOrder(b, userOrder.OrderId)
	}
}

func main() {
	eventBus := event_bus.NewEventBus()
	order.ListenOrder(eventBus)
	// Subscribe to the "post" topic event
	subscribe := eventBus.Subscribe("post")

	go func(){
		for event := range subscribe {
		fmt.Println(event.Payload)
		}
	}()

	eventBus.Publish("post", event_bus.Event{Payload: map[string]any{
		"postId": 1,
		"title":  "Welcome to Leapcell",
		"author": "Leapcell",
	}})
	// Topic with no subscribers
	eventBus.Publish("pay", event_bus.Event{Payload: "pay"})
	subscribe2 := eventBus.Subscribe("pay")
	go func() {
		for event := range subscribe2 {
		fmt.Println(event.Payload)
		}
	}()
	time.Sleep(time.Second * 2)
	// Unsubscribe from the "post" topic event
	eventBus.Unsubscribe("post", subscribe)
	// Register handlers for different URL paths
	// http.HandleFunc("/about", aboutHandler)

	// Start the HTTP server and listen on port 8080
	mux := http.NewServeMux()
	mux.HandleFunc("/order", createOrderHandler(eventBus))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Jalankan server di goroutine agar tidak blocking
	go func() {
		fmt.Println("Server started at http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// Tangkap sinyal interrupt (Ctrl+C) atau SIGTERM (docker, systemd, dsb)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit // blok sampai ada sinyal masuk
	fmt.Println("\nShutting down gracefully...")

	// Buat context dengan timeout (misal 5 detik)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Jalankan shutdown
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	} else {
		fmt.Println("Server stopped cleanly.")
	}
}