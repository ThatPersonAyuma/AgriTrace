package logistic

import ( 
    core "AgriTrace/Internal/Core"
    event_bus "AgriTrace/Internal/EventBus"
    generic "AgriTrace/Internal/Generic"
    "fmt"
    "time"

)

// If the order table has this column
// order
// id
// buyer_id
// status
// delivery_staff_id
// total_price
// estimated_time
// start_delivery
// end_delivery
// start_checkpoint (foreign to checkpoint), nullable
// end_checkpoint (foreign to checkpoint), nullable
// created_at
// updated_at

// checkpoint table
// id int
// order_id int
// status carvar 50
// timestamp timestamp
// location_lat numeric
// location_long numeric
// notes text
// type Command func(tx *sql.Tx) error
// func ShipmentCreated(order_id, start_location_et, start_location_long, end_location_et, end_location_long  int)generic.Result[bool]{

// }


func GetShipmentCreatedEffect( // Fungsi untuk membuat data yang dibutuhkan effect atau work
    orderID int,
    startLat float64,
    startLong float64,
    endLat float64,
    endLong float64,
    now time.Time,
) []generic.Effect {
    return []generic.Effect{
        {
			Type: generic.EffectDB,
            ExecCommand: `
                INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, notes)
                VALUES ($1, $2, $3, $4, $5, $6)
            `,
            Args: []any{orderID, "START_CREATED", now, startLat, startLong, "Initial start checkpoint"},
        },
        {
            Type: generic.EffectDB,
            ExecCommand: `
                INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, notes)
                VALUES ($1, $2, $3, $4, $5, $6)
            `,
            Args: []any{orderID, "END_CREATED", now, endLat, endLong, "Initial end checkpoint"},
        },
        {
            Type: generic.EffectDB,
            ExecCommand: `
                UPDATE "order"
                SET status=$1, start_delivery=$2, updated_at=$3
                WHERE id=$4
            `,
            Args: []any{"CREATED", now, now, orderID},
        },
        {
            Type: generic.EffectLog,
            Msg:  fmt.Sprintf("Shipment %d created at %v", orderID, now),
        },
    }
}

func ShipmentCreated(
	orderID int,
	deliveryStaffID int,
	checkpoints []core.Checkpoint,
	estimatedTime time.Time,
	now time.Time,
) generic.Result[[]generic.Effect] {
	// Validate checkpoints
	if err := validateCheckpoints(checkpoints); err != nil {
		return generic.Result[[]generic.Effect]{Value: nil, Err:err}
	}

	effects := make([]generic.Effect, 0, len(checkpoints)+2)

	// Insert all checkpoints
	for _, cp := range checkpoints {
		status := fmt.Sprintf("%s_CREATED", cp.Type)
		
		effects = append(effects, generic.Effect{
			Type: generic.EffectDBQuery,
			ExecCommand: `
				INSERT INTO checkpoints (order_id, staff_id, status, timestamp, location_lat, location_long, "type", notes)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id
			`,
			Args: []any{orderID, deliveryStaffID, status, cp.Timestamp, cp.Lat, cp.Long, string(cp.Type), cp.Notes},
		})
	}

	// Update order with shipment info
	effects = append(effects, generic.Effect{
		Type: generic.EffectDB,
		ExecCommand: `
			UPDATE orders
			SET status = $1,
				estimated_time = $2,
				updated_at = $3
			WHERE id = $4
		`,
		Args: []any{"SHIPMENT_CREATED", estimatedTime, now, orderID},
	})

	// Log
	effects = append(effects, generic.Effect{
		Type: generic.EffectLog,
		Msg:  fmt.Sprintf("Shipment created for order %d by staff %d with %d checkpoints", orderID, deliveryStaffID, len(checkpoints)),
	})

	return generic.Result[[]generic.Effect]{Value: effects, Err:nil}
}

func CheckpointAdded(
	orderID int,
	lat, long float64,
	status string,
	notes string,
	now time.Time,
) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
				INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, type, notes)
				VALUES ($1, $2, $3, $4, $5, 'PROGRESS', $6)
			`,
			Args: []any{orderID, status, now, lat, long, notes},
		},
		{
			Type: generic.EffectDB,
			ExecCommand: `
				UPDATE orders
				SET updated_at = $1
				WHERE id = $2
			`,
			Args: []any{now, orderID},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Checkpoint added for order %d: %s at (%.6f, %.6f)", orderID, status, lat, long),
		},
	}
}

func CheckpointPhotoUploaded(
	checkpointID int,
	photoURL string,
	now time.Time,
) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
				UPDATE checkpoint
				SET photo_url = $1,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{photoURL, now, checkpointID},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Photo uploaded for checkpoint %d: %s", checkpointID, photoURL),
		},
	}
}

func CheckpointVerified(
	checkpointID int,
	verifiedBy int,
	now time.Time,
) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
				UPDATE checkpoint
				SET verified = true,
					verified_by = $1,
					verified_at = $2
				WHERE id = $3
			`,
			Args: []any{verifiedBy, now, checkpointID},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Checkpoint %d verified by user %d", checkpointID, verifiedBy),
		},
	}
}

func ShipmentCompleted(
	orderID int,
	finalLat, finalLong float64,
	deliveryProof string,
	now time.Time,
) []generic.Effect {
	return []generic.Effect{
		// Insert final checkpoint
		{
			Type: generic.EffectDB,
			ExecCommand: `
				INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, type, notes)
				VALUES ($1, $2, $3, $4, $5, 'END', $6)
			`,
			Args: []any{orderID, "DELIVERED", now, finalLat, finalLong, deliveryProof},
		},
		// Update order status
		{
			Type: generic.EffectDB,
			ExecCommand: `
				UPDATE orders
				SET status = $1,
					end_delivery = $2,
					updated_at = $2
				WHERE id = $3
			`,
			Args: []any{"COMPLETED", now, orderID},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Shipment %d completed at %v", orderID, now),
		},
	}
}

func ShipmentDelayed(
	orderID int,
	reason string,
	newEstimatedTime time.Time,
	now time.Time,
) []generic.Effect {
	return []generic.Effect{
		{
			Type: generic.EffectDB,
			ExecCommand: `
				UPDATE orders
				SET status = $1,
					estimated_time = $2,
					updated_at = $3,
					delay_reason = $4
				WHERE id = $5
			`,
			Args: []any{"DELAYED", newEstimatedTime, now, reason, orderID},
		},
		{
			Type: generic.EffectLog,
			Msg:  fmt.Sprintf("Shipment %d delayed: %s. New ETA: %v", orderID, reason, newEstimatedTime),
		},
		// Optional: Send notification effect
		{
			Type: generic.EffectNotify,
			Msg:  fmt.Sprintf("Your order #%d is delayed. Reason: %s", orderID, reason),
		},
	}
}
func validateCheckpoints(checkpoints []core.Checkpoint) error {
	if len(checkpoints) < 2 {
		return fmt.Errorf("at least START and END checkpoints are required")
	}

	hasStart := false
	hasEnd := false

	for _, cp := range checkpoints {
		switch cp.Type {
		case core.CheckpointStart:
			if hasStart {
				return fmt.Errorf("only one START checkpoint is allowed")
			}
			hasStart = true
		case core.CheckpointEnd:
			if hasEnd {
				return fmt.Errorf("only one END checkpoint is allowed")
			}
			hasEnd = true
		case core.CheckpointOnRoad:
		default:
			return fmt.Errorf("invalid checkpoint type: %s", cp.Type)
		}
	}

	if !hasStart {
		return fmt.Errorf("START checkpoint is required")
	}
	if !hasEnd {
		return fmt.Errorf("END checkpoint is required")
	}

	return nil
}
// Validate validates the shipment creation request
func Validate(req core.ShipmentCreatedReq) error {
	if req.OrderID <= 0 {
		return fmt.Errorf("invalid order_id")
	}
	if req.DeliveryStaffID <= 0 {
		return fmt.Errorf("invalid delivery_staff_id")
	}
	if req.EstimatedTime.IsZero() {
		return fmt.Errorf("estimated_time is required")
	}
	return validateCheckpoints(req.Checkpoints)
}
// ============================================================================
// EVENT LISTENER - Convert Events to Effects
// ============================================================================

func ListenLogistic(b *event_bus.EventBus, topic, worker_topic string, job_store *generic.JobStore) {
	sub := b.Subscribe(topic)
	
	go func(job_store *generic.JobStore) {
		for event := range sub {
			now := time.Now().UTC()
			var data []generic.Effect

			switch event.SubTopic {
			case core.ShipmentCreated:
				payload, ok := event.Payload.(core.ShipmentCreatedReq)
				if !ok {
					fmt.Printf("ERROR: Invalid payload for ShipmentCreated: %+v\n", event.Payload)
					continue
				}
				result := ShipmentCreated(
					payload.OrderID,
					payload.DeliveryStaffID,
					payload.Checkpoints,
					payload.EstimatedTime,
					time.Now(),
				)
				if (result.Err != nil){
					fmt.Print(result.Err.Error());
					continue
				}else{
					data = result.Value
				}

			case core.CheckpointAdded:
				payload, ok := event.Payload.(core.CheckpointAddedReq)
				if !ok {
					fmt.Printf("ERROR: Invalid payload for CheckpointAdded: %+v\n", event.Payload)
					continue
				}
				data = CheckpointAdded(
					payload.OrderID,
					payload.Lat,
					payload.Long,
					payload.Status,
					payload.Notes,
					now,
				)

			case core.CheckpointPhotoUploaded:
				payload, ok := event.Payload.(core.CheckpointPhotoReq)
				if !ok {
					fmt.Printf("ERROR: Invalid payload for CheckpointPhotoUploaded: %+v\n", event.Payload)
					continue
				}
				data = CheckpointPhotoUploaded(
					payload.CheckpointID,
					payload.PhotoURL,
					now,
				)

			case core.CheckpointVerified:
				payload, ok := event.Payload.(core.CheckpointVerifyReq)
				if !ok {
					fmt.Printf("ERROR: Invalid payload for CheckpointVerified: %+v\n", event.Payload)
					continue
				}
				data = CheckpointVerified(
					payload.CheckpointID,
					payload.VerifiedBy,
					now,
				)

			case core.ShipmentCompleted:
				payload, ok := event.Payload.(core.ShipmentCompletedReq)
				if !ok {
					fmt.Printf("ERROR: Invalid payload for ShipmentCompleted: %+v\n", event.Payload)
					continue
				}
				data = ShipmentCompleted(
					payload.OrderID,
					payload.FinalLat,
					payload.FinalLong,
					payload.DeliveryProof,
					now,
				)

			case core.ShipmentDelayed:
				payload, ok := event.Payload.(core.ShipmentDelayedReq)
				if !ok {
					fmt.Printf("ERROR: Invalid payload for ShipmentDelayed: %+v\n", event.Payload)
					continue
				}
				data = ShipmentDelayed(
					payload.OrderID,
					payload.Reason,
					payload.NewEstimatedTime,
					now,
				)

			default:
				fmt.Printf("WARN: Unknown SubTopic: %s\n", event.SubTopic)
				continue
			}

			// Validation
			if len(data) == 0 {
				fmt.Printf("WARN: No effects generated for %s\n", event.SubTopic)
				continue
			}

			// Publish to worker
			work_event := event_bus.Event{
				WorkId:  event.WorkId,
				Payload: data,
			}
			b.Publish(worker_topic, work_event)
		}
	}(job_store)
}