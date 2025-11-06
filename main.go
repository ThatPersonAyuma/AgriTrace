package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal"
)

// type user_order struct{
// 	OrderId int `json:"OrderId"`
// }

// func createOrderHandler(b *event_bus.EventBus)func(http.ResponseWriter, *http.Request){
// 	return func (w http.ResponseWriter, r *http.Request){
// 		// Pastikan method POST
// 		if r.Method != http.MethodPost {
// 			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		// Decode JSON body
// 		var userOrder user_order
// 		err := json.NewDecoder(r.Body).Decode(&userOrder)
// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
// 			return
// 		}
// 		order.CreateOrder(b, userOrder.OrderId)
// 	}
// }



func main() {
	eventBus := event_bus.NewEventBus()
	mux := http.NewServeMux()

	internal.Setup()(mux, eventBus)

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