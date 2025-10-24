package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"AgriTrace/broker"
	"AgriTrace/modules/order"
)

func BenchmarkCreateOrderHandler(b *testing.B) {
	eventBus := event_bus.NewEventBus()
	order.ListenOrder(eventBus)

	handler := createOrderHandler(eventBus)
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := &http.Client{}

	orderData := map[string]int{"OrderId": 123}
	body, _ := json.Marshal(orderData)

	b.ResetTimer() // mulai hitung waktu benchmark
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", server.URL, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			b.Fatalf("Request error: %v", err)
		}
		resp.Body.Close()
	}
}