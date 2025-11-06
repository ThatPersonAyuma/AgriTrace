package internal

import (
	"AgriTrace/Internal/Adapters/Http"
	core "AgriTrace/Internal/Core"
	"AgriTrace/Internal/EventBus"
	workers "AgriTrace/Internal/Workers"
	"AgriTrace/Internal/Generic"
	"net/http"
)

func Setup() func(*http.ServeMux, *event_bus.EventBus){
	return func(mux *http.ServeMux, eventBus *event_bus.EventBus){
		// order.ListenOrder(eventBus)
		job_store := generic.JobStore{Data: map[string]generic.JobResult{}}
		core.ListenLogin(eventBus, "Login", "DynWorks", &job_store)
		workers.ListenDynWork(eventBus, "DynWorks", 10, 20)
		workers.ListenFixWorks(eventBus, "FixWorks", 8)
		// mux.HandleFunc("/order", createOrderHandler(eventBus))
		mux.HandleFunc("/login", http_adapters.CreateFuncHandler[generic.UserLogin](eventBus, &job_store, http.MethodPost, "Login"))
		mux.HandleFunc("/check-work", http_adapters.CreateGetStatusHandler(&job_store))
	}
}