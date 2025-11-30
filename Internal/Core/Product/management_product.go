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
			EcexCommand: `
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
			EcexCommand: `
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

		// 1. Log pencarian
		// {
		// 	Type: generic.EffectDB,
		// 	EcexCommand: `
		// 		INSERT INTO search_logs (user_id, keywords, searched_at)
		// 		VALUES ($1, $2, $3)
		// 	`,
		// 	Args: []any{userID, keywords, now},
		// },, now time.Time, userID int, 

		// 2. Query pencarian produk
		{
			Type: generic.EffectDBQuery,
			EcexCommand: `
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
func ListenProduct(b *event_bus.EventBus, topic, worker_topic string, job_store *generic.JobStore){
	sub := b.Subscribe(topic)
	go func(job_store *generic.JobStore){
		for event := range sub{
			var data []generic.Effect
			// println("listener:", event.WorkId, event.SubTopic)
			switch event.SubTopic {
				case core.ProductListed:
					data = ProductListed()
				case core.ProductUnlisted:
					data = ProductUnlisted()
				case core.SerachProduct:
					payload, ok := event.Payload.(core.SerachKeyword)
					if !ok {
						return
					}
					data = SearchPerformed(payload.Keyword)
			}
			// userLogin, ok := event.Payload.(generic.UserLogin)
			var work_event event_bus.Event
			work_event.WorkId = event.WorkId
			work_event.Payload = data
			b.Publish(worker_topic, work_event)
		}
	}(job_store)
}