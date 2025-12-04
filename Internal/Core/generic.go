package core

import "time"

type SubTopic string

const (
	OrderCreated           = "Created"
	OrderPaid              = "Paid"
	OrderCancelled         = "Canceled"
	OrderConfirmedByFarmer = "Confirmed"
	OrderPrepared          = "Prepared"
	OrderShipped           = "Shipped"
	OrderDelivered         = "Delivered"
	OrderCompleted         = "Completed"

	ProductListed   = "Listed"
	ProductUnlisted = "Unlisted"
	SerachProduct   = "Search"

	ShipmentCreated         = "ShipmentCreated"
	CheckpointAdded         = "CheckpointAdded"
	CheckpointPhotoUploaded = "CheckpointPhotoUploaded"
	CheckpointVerified      = "CheckpointVerified"
	ShipmentCompleted       = "ShipmentCompleted"
	ShipmentDelayed         = "ShipmentDelayed"

	AccountCreated = "Created"
	AccountUpdated = "Updated"
)

type OrderIDReq struct {
	OrderID int `json:"order_id"`
}
type OrderCreatedReq struct {
	BuyerID int `json:"buyer_id"`
	// StaffID int 	`json:"_staff_id"`

}

type OrderCoordinateReq struct {
	OrderID int     `json:"order_id"`
	Lat     float64 `json:"lat"`
	Long    float64 `json:"long"`
}

type SerachKeyword struct {
	Keyword string `json:"keyword"`
}

type Nothing struct {
}
type ShipmentCreatedReq struct {
	OrderID         int       `json:"order_id"`
	DeliveryStaffID int       `json:"delivery_staff_id"`
	StartLat        float64   `json:"start_lat"`
	StartLong       float64   `json:"start_long"`
	EndLat          float64   `json:"end_lat"`
	EndLong         float64   `json:"end_long"`
	EstimatedTime   time.Time `json:"estimated_time"`
}

type CheckpointAddedReq struct {
	OrderID int     `json:"order_id"`
	Lat     float64 `json:"lat"`
	Long    float64 `json:"long"`
	Status  string  `json:"status"`
	Notes   string  `json:"notes"`
}

type CheckpointPhotoReq struct {
	CheckpointID int    `json:"checkpoint_id"`
	PhotoURL     string `json:"photo_url"`
}

type CheckpointVerifyReq struct {
	CheckpointID int `json:"checkpoint_id"`
	VerifiedBy   int `json:"verified_by"`
}

type ShipmentCompletedReq struct {
	OrderID       int     `json:"order_id"`
	FinalLat      float64 `json:"final_lat"`
	FinalLong     float64 `json:"final_long"`
	DeliveryProof string  `json:"delivery_proof"`
}

type ShipmentDelayedReq struct {
	OrderID          int       `json:"order_id"`
	Reason           string    `json:"reason"`
	NewEstimatedTime time.Time `json:"new_estimated_time"`
}
type AccountCreatedReq struct {
    AccountID    string    `json:"account_id"`
    UsersID      string    `json:"users_id"`
    Name         string    `json:"name"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"password_hash"`
    Phone        string    `json:"phone"`
}
type AccountUpdatedReq struct {
    AccountID string    `json:"account_id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Phone     string    `json:"phone"`
}