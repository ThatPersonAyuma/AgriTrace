package internal

import (
	"AgriTrace/Internal/Adapters/Http"
	core "AgriTrace/Internal/Core"
	order "AgriTrace/Internal/Core/Order"
	"AgriTrace/Internal/EventBus"
	"AgriTrace/Internal/Generic"
	workers "AgriTrace/Internal/Workers"
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
)

func Setup() func(*http.ServeMux, *event_bus.EventBus, *sql.DB){
	return func(mux *http.ServeMux, eventBus *event_bus.EventBus, db *sql.DB){
		job_store := generic.JobStore{Data: map[string]generic.JobResult{}}
		core.ListenLogin(eventBus, "Login", "DynWorks", &job_store)
		order.ListenOrder(eventBus, "Order", "DynWorks", &job_store)
		workers.ListenDynWork(eventBus, "DynWorks", 6, 10, db)
		workers.ListenFixWorks(eventBus, "FixWorks", 5, db)
		// mux.HandleFunc("/order", createOrderHandler(eventBus))
		mux.HandleFunc("/login", http_adapters.CreateFuncHandler[generic.UserLogin](eventBus, &job_store, http.MethodPost, "Login"))
		// Register Handle For Order Feature
		mux.HandleFunc("/order-create", http_adapters.CreateFuncHandler[core.OrderCreatedReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderCreated)))
		mux.HandleFunc("/order-paid", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderPaid)))
		mux.HandleFunc("/order-canceled", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderCancelled)))
		mux.HandleFunc("/order-confirmed", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderConfirmedByFarmer)))
		mux.HandleFunc("/order-prepared", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderPrepared)))
		mux.HandleFunc("/order-shipped", http_adapters.CreateFuncHandler[core.OrderCoordinateReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderShipped)))
		mux.HandleFunc("/order-delivered", http_adapters.CreateFuncHandler[core.OrderCoordinateReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderDelivered)))
		mux.HandleFunc("/order-completed", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderCompleted)))
		
		mux.HandleFunc("/check-work", http_adapters.CreateGetStatusHandler(&job_store))
	}
}