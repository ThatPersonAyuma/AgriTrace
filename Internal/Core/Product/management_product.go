package product

import (
	core "AgriTrace/Internal/Core"
	event_bus "AgriTrace/Internal/EventBus"
	generic "AgriTrace/Internal/Generic"
	"fmt"
	// "time"
)

func ProductListed() []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDBQuery,
			ExecCommand: `
				SELECT id, farmer_id, name, description, price, stock, min_order, updated_at
				FROM products
				WHERE stock > products.min_order
				ORDER BY name ASC
			`,
			Args: []any{},
		},
	}
}
func FarmerProductListed(farmer_id int) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDBQuery,
			ExecCommand: `
				SELECT id, farmer_id, name, description, price, stock, min_order, updated_at
				FROM products
				WHERE farmer_id = $1
				ORDER BY name ASC
			`,
			Args: []any{farmer_id},
		},
	}
}

func ProductUnlisted() []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDBQuery,
			ExecCommand: `
				SELECT id, farmer_id, name, description, price, stock, updated_at
				FROM products
				ORDER BY name ASC
			`,
			Args: []any{},
		},
	}
}

func SearchPerformed(keywords string) []generic.Effect {
	return []generic.Effect{

		// 2. Query pencarian produk
		{
			Type: generic.EffectDBQuery,
			ExecCommand: `
				SELECT id, farmer_id, name, description, price, stock
				FROM products
				WHERE LOWER(name) LIKE LOWER($1)
				ORDER BY name ASC
			`,
			Args: []any{
				fmt.Sprintf("%%%s%%", keywords), // agar prefix/infix/suffix
			},
		},
	}
}
func CreateProduct(farmerID int, name, description string, price float64, stock, min_order int) []generic.Effect {
	return []generic.Effect{{
		Type: generic.EffectDBQuery,
		ExecCommand: `
			INSERT INTO products (farmer_id, name, description, price, stock, min_order)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`,
		Args: []any{
			farmerID,
			name,
			description,
			price,
			stock,
			min_order,
		},
	}}
}

func ListenProduct(b *event_bus.EventBus, topic, worker_topic string, job_store *generic.JobStore) {
	sub := b.Subscribe(topic)

	go func() {
		for event := range sub {
			var effects []generic.Effect
			var err error

			switch event.SubTopic {

			case core.ProductListed:
				effects = ProductListed()

			case core.ProductUnlisted:
				effects = ProductUnlisted()

			case core.SerachProduct:
				payload, ok := event.Payload.(core.SerachKeyword)
				if !ok {
					err = fmt.Errorf("invalid payload for SearchProduct")
				} else {
					effects = SearchPerformed(payload.Keyword)
				}

			case core.ProductCreated:
				payload, ok := event.Payload.(core.ProductCreatedReq)
				if !ok {
					err = fmt.Errorf("invalid payload for SearchProduct")
				}else{
					effects = CreateProduct(
						payload.FarmerID,
						payload.Name,
						payload.Description,
						payload.Price,
						payload.Stock,
						payload.MinOrder,
					)
				}

			case core.FarmerProducts:
				payload, ok := event.Payload.(core.FarmeIdReq)
				if !ok {
					err = fmt.Errorf("invalid payload for FarmerProducts")
				}else{
					effects = FarmerProductListed(
						payload.FarmerID,
					)
				}

			default:
				err = fmt.Errorf("unknown product subtopic: %s", event.SubTopic)
			}

			if err != nil {
				job_store.Lock()
				job_store.Data[event.WorkId] = generic.JobResult{
					Status: "error",
					Error:  err.Error(),
				}
				job_store.Unlock()
				continue
			}

			b.Publish(worker_topic, event_bus.Event{
				WorkId:  event.WorkId,
				Payload: effects,
			})
		}
	}()
}
