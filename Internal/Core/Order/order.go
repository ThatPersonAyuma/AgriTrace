package order

import (
	core "AgriTrace/Internal/Core"
	event_bus "AgriTrace/Internal/EventBus"
	generic "AgriTrace/Internal/Generic"
	"fmt"
	"time"
)

// {
// 	Type: generic.EffectDB,
// 	EcexCommand: `
// 		INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, notes)
// 		VALUES ($1, $2, $3, $4, $5, $6)
// 	`,
// 	Args: []any{orderID, "START_CREATED", now, startLat, startLong, "Initial start checkpoint"},
// },
func OrderCreated(buyerID int, totalPrice float32, now time.Time) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			EcexCommand: `
				INSERT INTO orders (buyer_id, status, total_price, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $4)
			`,
			Args: []any{buyerID, "KERANJANG", totalPrice, now},
		},
	}
}

func OrderPaid(orderID int, now time.Time) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			EcexCommand: `
				UPDATE orders
				SET status = $1,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"PAID", now, orderID},
		},
	}
}

func Pay() generic.Result[bool]{
	return generic.Result[bool]{Value: true}
}
func OrderCanceled(orderID int, now time.Time) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			EcexCommand: `
				UPDATE orders
				SET status = $1,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"CANCELED", now, orderID},
		},
	}
}
func OrderConfirmedByFarmer(orderID int, now time.Time) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			EcexCommand: `
				UPDATE orders
				SET status = $1,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"CONFIRMED", now, orderID},
		},
	}
}
func OrderPrepared(orderID int, now time.Time) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			EcexCommand: `
				UPDATE orders
				SET status = $1,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"PREPARED", now, orderID},
		},
	}
}
func OrderShipped(orderID int, now time.Time, lat, long float64) []generic.Effect {
	return []generic.Effect{
		// Checkpoint START
		{
			Type: generic.EffectDB,
			EcexCommand: `
				INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, type, notes)
				VALUES ($1, $2, $3, $4, $5, 'START', $6)
			`,
			Args: []any{orderID, "START_SHIPPING", now, lat, long, "Delivery started"},
		},

		// Update orders
		{
			Type: generic.EffectDB,
			EcexCommand: `
				UPDATE orders
				SET status = $1,
					start_delivery = $2,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"SHIPPED", now, orderID},
		},
	}
}
func OrderDelivered(orderID int, now time.Time, lat, long float64) []generic.Effect {
	return []generic.Effect{
		// Checkpoint END
		{
			Type: generic.EffectDB,
			EcexCommand: `
				INSERT INTO checkpoints (order_id, status, timestamp, location_lat, location_long, type, notes)
				VALUES ($1, $2, $3, $4, $5, 'END', $6)
			`,
			Args: []any{orderID, "DELIVERED", now, lat, long, "Delivered to buyer"},
		},

		// Update order
		{
			Type: generic.EffectDB,
			EcexCommand: `
				UPDATE orders
				SET status = $1,
					end_delivery = $2,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"DELIVERED", now, orderID},
		},
	}
}
func OrderCompleted(orderID int, now time.Time) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			EcexCommand: `
				UPDATE orders
				SET status = $1,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"COMPLETED", now, orderID},
		},
	}
}

func ListenOrder(b *event_bus.EventBus, topic, worker_topic string, job_store *generic.JobStore){
	sub := b.Subscribe(topic)
	go func(job_store *generic.JobStore){
		for event := range sub{
			var data []generic.Effect
			println("listener:", event.WorkId, event.SubTopic)
			fmt.Printf("%q\n", core.OrderCreated)
			fmt.Printf("%q\n", event.SubTopic)
			fmt.Printf("%q\n", "Created")
			switch event.SubTopic {
				case core.OrderCreated:
					println("running case")
					payload, ok := event.Payload.(core.OrderCreatedReq)
					if !ok {
						fmt.Println("Wrong data:", event.Payload)
						return
					}
					data = OrderCreated(payload.BuyerID, 1000.0, time.Now().UTC())
					println("correct")
				case core.OrderPaid:
					payload, ok := event.Payload.(core.OrderIDReq)
					if !ok {
						return
					}
					data = OrderPaid(payload.OrderID, time.Now().UTC())
				case core.OrderCancelled:
					payload, ok := event.Payload.(core.OrderIDReq)
					if !ok {
						return
					}
					data = OrderCanceled(payload.OrderID, time.Now().UTC())
				case core.OrderConfirmedByFarmer:
					payload, ok := event.Payload.(core.OrderIDReq)
					if !ok {
						return
					}
					data = OrderConfirmedByFarmer(payload.OrderID, time.Now().UTC())
				case core.OrderPrepared:
					payload, ok := event.Payload.(core.OrderIDReq)
					if !ok {
						return
					}
					data = OrderPrepared(payload.OrderID, time.Now().UTC())
				case core.OrderShipped:
					payload, ok := event.Payload.(core.OrderCoordinateReq)
					if !ok {
						return
					}
					data = OrderShipped(payload.OrderID, time.Now().UTC(), payload.Lat, payload.Long)
				case core.OrderDelivered:
					payload, ok := event.Payload.(core.OrderCoordinateReq)
					if !ok {
						return
					}
					data = OrderDelivered(payload.OrderID, time.Now().UTC(), payload.Lat, payload.Long)
				case core.OrderCompleted:
					payload, ok := event.Payload.(core.OrderIDReq)
					if !ok {
						return
					}
					data = OrderCompleted(payload.OrderID, time.Now().UTC())
			}
			// userLogin, ok := event.Payload.(generic.UserLogin)
			var work_event event_bus.Event
			work_event.WorkId = event.WorkId
			work_event.Payload = data
			b.Publish(worker_topic, work_event)
		}
	}(job_store)
}