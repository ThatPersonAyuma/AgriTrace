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
// 	ExecCommand: `
// 		INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, notes)
// 		VALUES ($1, $2, $3, $4, $5, $6)
// 	`,
// 	Args: []any{orderID, "START_CREATED", now, startLat, startLong, "Initial start checkpoint"},
// },
func OrderCreated(buyerID int, totalPrice float32, now time.Time) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
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
			ExecCommand: `
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
			ExecCommand: `
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
			ExecCommand: `
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
			ExecCommand: `
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
			ExecCommand: `
				INSERT INTO checkpoints (order_id, status, timestamp, location_lat, location_long, type, notes)
				VALUES ($1, $2, $3, $4, $5, 'START', $6)
			`,
			Args: []any{orderID, "START_SHIPPING", now, lat, long, "Delivery started"},
		},

		// Update orders
		{
			Type: generic.EffectDB,
			ExecCommand: `
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
			ExecCommand: `
				INSERT INTO checkpoints (order_id, status, timestamp, location_lat, location_long, type, notes)
				VALUES ($1, $2, $3, $4, $5, 'END', $6)
			`,
			Args: []any{orderID, "DELIVERED", now, lat, long, "Delivered to buyer"},
		},

		// Update order
		{
			Type: generic.EffectDB,
			ExecCommand: `
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
			ExecCommand: `
				UPDATE orders
				SET status = $1,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"COMPLETED", now, orderID},
		},
	}
}
func AddItemToCart(orderID, productID, quantity int, unitPrice float64) []generic.Effect {

	subtotal := float64(quantity) * unitPrice

	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
				INSERT INTO order_items (order_id, product_id, quantity, unit_price, subtotal)
				VALUES ($1, $2, $3, $4, $5)
			`,
			Args: []any{
				orderID,
				productID,
				quantity,
				unitPrice,
				subtotal,
			},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Added product %d x %d to order %d", productID, quantity, orderID),
		},
	}
}

func ListenOrder(b *event_bus.EventBus, topic, workerTopic string, jobStore *generic.JobStore) {
	sub := b.Subscribe(topic)

	go func() {
		for event := range sub {

			var effects []generic.Effect
			var err error

			switch event.SubTopic {

			case core.OrderCreated:
				payload, ok := event.Payload.(core.OrderCreatedReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderCreated")
				} else {
					effects = OrderCreated(payload.BuyerID, payload.TotalPrice, time.Now().UTC())
				}

			case core.OrderPaid:
				payload, ok := event.Payload.(core.OrderIDReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderPaid")
				} else {
					effects = OrderPaid(payload.OrderID, time.Now().UTC())
				}

			case core.OrderCancelled:
				payload, ok := event.Payload.(core.OrderIDReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderCancelled")
				} else {
					effects = OrderCanceled(payload.OrderID, time.Now().UTC())
				}

			case core.OrderConfirmedByFarmer:
				payload, ok := event.Payload.(core.OrderIDReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderConfirmedByFarmer")
				} else {
					effects = OrderConfirmedByFarmer(payload.OrderID, time.Now().UTC())
				}

			case core.OrderPrepared:
				payload, ok := event.Payload.(core.OrderIDReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderPrepared")
				} else {
					effects = OrderPrepared(payload.OrderID, time.Now().UTC())
				}

			case core.OrderShipped:
				payload, ok := event.Payload.(core.OrderCoordinateReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderShipped")
				} else {
					effects = OrderShipped(payload.OrderID, time.Now().UTC(), payload.Lat, payload.Long)
				}

			case core.OrderDelivered:
				payload, ok := event.Payload.(core.OrderCoordinateReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderDelivered")
				} else {
					effects = OrderDelivered(payload.OrderID, time.Now().UTC(), payload.Lat, payload.Long)
				}

			case core.OrderCompleted:
				payload, ok := event.Payload.(core.OrderIDReq)
				if !ok {
					err = fmt.Errorf("invalid payload for OrderCompleted")
				} else {
					effects = OrderCompleted(payload.OrderID, time.Now().UTC())
				}
			case core.AddToCart:
				payload, ok := event.Payload.(core.AddToCartPayload)
				if !ok {
					err = fmt.Errorf("invalid payload for AddToCart")
				} else {
					effects = AddItemToCart(payload.OrderID, payload.ProductID, payload.Quantity, payload.UnitPrice)
				}
			default:
				err = fmt.Errorf("unknown order subtopic: %s", event.SubTopic)
			}

			if err != nil {
				jobStore.Lock()
				jobStore.Data[event.WorkId] = generic.JobResult{
					Status: "error",
					Error:  err.Error(),
				}
				jobStore.Unlock()
				continue
			}

			b.Publish(workerTopic, event_bus.Event{
				WorkId:  event.WorkId,
				Payload: effects,
			})
		}
	}()
}
