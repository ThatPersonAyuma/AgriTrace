package logistic

import (
	generic "AgriTrace/Internal/Generic"
	// "database/sql"
	"fmt"
	"time"

	// "github.com/lib/pq"
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
            Type: generic.EffectDBExec,
            EcexCommand: `
                INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, notes)
                VALUES ($1, $2, $3, $4, $5, $6)
            `,
            Args: []any{orderID, "START_CREATED", now, startLat, startLong, "Initial start checkpoint"},
        },
        {
            Type: generic.EffectDBExec,
            EcexCommand: `
                INSERT INTO checkpoint (order_id, status, timestamp, location_lat, location_long, notes)
                VALUES ($1, $2, $3, $4, $5, $6)
            `,
            Args: []any{orderID, "END_CREATED", now, endLat, endLong, "Initial end checkpoint"},
        },
        {
            Type: generic.EffectDBExec,
            EcexCommand: `
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
func ShipmentCreated(){
    
}

func CheckpointCreate(){

}
func CheckpointAdded(){

}
func CheckpointPhotoUploaded(){

}
func CheckpointVerified(){

}
func ShipmentCompleted(){

}
func ShipmentDelayed(){

}