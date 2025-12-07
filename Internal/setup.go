package internal

import (
	"AgriTrace/Internal/Adapters/Http"
	core "AgriTrace/Internal/Core"
	order "AgriTrace/Internal/Core/Order"
	product "AgriTrace/Internal/Core/Product"
	account "AgriTrace/Internal/Core/Account"
	logistic "AgriTrace/Internal/Core/Logistic"
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
		order.ListenOrder(eventBus, "Order", "DynWorks", &job_store)
		product.ListenProduct(eventBus, "Product", "FixWorks", &job_store)
		account.ListenAccount(eventBus, "Account", "FixWorks", &job_store)
		logistic.ListenLogistic(eventBus, "Logistic", "FixWorks", &job_store)
		workers.ListenDynWork(eventBus, "DynWorks", 2, 4, db, &job_store)
		workers.ListenFixWorks(eventBus, "FixWorks", 3, db, &job_store)
		mux.HandleFunc("/order-create", http_adapters.CreateFuncHandler[core.OrderCreatedReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderCreated)))
		mux.HandleFunc("/order-paid", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderPaid)))
		mux.HandleFunc("/order-canceled", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderCancelled)))
		mux.HandleFunc("/order-confirmed", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderConfirmedByFarmer)))
		mux.HandleFunc("/order-prepared", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderPrepared)))
		mux.HandleFunc("/order-shipped", http_adapters.CreateFuncHandler[core.OrderCoordinateReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderShipped)))
		mux.HandleFunc("/order-delivered", http_adapters.CreateFuncHandler[core.OrderCoordinateReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderDelivered)))
		mux.HandleFunc("/order-completed", http_adapters.CreateFuncHandler[core.OrderIDReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.OrderCompleted)))
		mux.HandleFunc("/product/listed", http_adapters.CreateFuncHandler[core.Nothing](eventBus, &job_store, http.MethodGet, fmt.Sprintf("Product.%s", core.ProductListed)))
		mux.HandleFunc("/product/unlisted", http_adapters.CreateFuncHandler[core.Nothing](eventBus, &job_store, http.MethodGet, fmt.Sprintf("Product.%s", core.ProductUnlisted)))
		mux.HandleFunc("/product/search", http_adapters.CreateFuncHandler[core.SerachKeyword](eventBus, &job_store, http.MethodGet, fmt.Sprintf("Product.%s", core.SerachProduct)))
		mux.HandleFunc("/product/create", http_adapters.CreateFuncHandler[core.ProductCreatedReq](eventBus, &job_store, http.MethodGet, fmt.Sprintf("Product.%s", core.ProductCreated)))
		mux.HandleFunc("/product/farmer", http_adapters.CreateFuncHandler[core.FarmeIdReq](eventBus, &job_store, http.MethodGet, fmt.Sprintf("Product.%s", core.FarmerProducts)))
		mux.HandleFunc("/order/add-product", http_adapters.CreateFuncHandler[core.AddToCartPayload](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Order.%s", core.AddToCart)))


		// Accounts
		mux.HandleFunc("/account/create", http_adapters.CreateFuncHandler[core.AccountCreatedReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Account.%s", core.AccountCreated)))
		mux.HandleFunc("/account/update", http_adapters.CreateFuncHandler[core.AccountUpdatedReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Account.%s", core.AccountUpdated)))

		// Logistic
		mux.HandleFunc("/logistic/checkpoint/upload", http_adapters.HandleCheckpointPhotoUploadMultipart(eventBus, &job_store, fmt.Sprintf("Logistic.%s", core.CheckpointPhotoUploaded)))
		mux.HandleFunc("/uploads/", http_adapters.HandleServeUploadedFile())
		mux.HandleFunc("/logistic/create", http_adapters.CreateFuncHandler[core.ShipmentCreatedReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Logistic.%s", core.ShipmentCreated)))
		mux.HandleFunc("/logistic/get", http_adapters.CreateFuncHandler[core.GetOrderCheckpointsReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Logistic.%s", core.GetShipment)))
		mux.HandleFunc("/logistic/getWithPhotos", http_adapters.CreateFuncHandler[core.GetOrderCheckpointsReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Logistic.%s", core.GetShipmentWithImage)))
		mux.HandleFunc("/logistic/checkpoint/added", http_adapters.CreateFuncHandler[core.CheckpointAddedReq](eventBus, &job_store, http.MethodPost, fmt.Sprintf("Logistic.%s", core.CheckpointAdded)))
		mux.HandleFunc("/logistic/checkpoint/verify", http_adapters.CreateFuncHandler[core.CheckpointVerifyReq](eventBus, &job_store,  http.MethodPost, fmt.Sprintf("Logistic.%s", core.CheckpointVerified)))
		mux.HandleFunc("/logistic/completed", http_adapters.CreateFuncHandler[core.ShipmentCompletedReq](eventBus, &job_store,  http.MethodPost, fmt.Sprintf("Logistic.%s", core.ShipmentCompleted)))
		mux.HandleFunc("/logistic/delayed", http_adapters.CreateFuncHandler[core.ShipmentDelayedReq](eventBus, &job_store,  http.MethodPost, fmt.Sprintf("Logistic.%s", core.ShipmentDelayed)))

		mux.HandleFunc("/work/check", http_adapters.CreateGetStatusHandler(&job_store))
		
	}
}