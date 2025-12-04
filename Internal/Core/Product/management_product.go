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
