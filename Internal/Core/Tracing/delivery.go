package tracing

import (
	event_bus "AgriTrace/Internal/EventBus"
	generic "AgriTrace/Internal/Generic"
)

func DeliveryCreated(){
	
}
func CheckpointAdded(){

}
func CheckpointPhotoUploaded(){

}
func CheckpointVerified(){

}
func DeliveryCompleted(){

}
func DeliveryDelayed(){

}
func ListenTracing(b *event_bus.EventBus, topic, worker_topic string, job_store *generic.JobStore) {
	
}