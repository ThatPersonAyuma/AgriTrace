package core

import (
	"encoding/json"
	"fmt"
	"time"
)

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
type CheckpointType string

const (
	CheckpointStart  CheckpointType = "START"
	CheckpointOnRoad CheckpointType = "ONROAD"
	CheckpointEnd    CheckpointType = "END"
)

type Checkpoint struct {
	Type      CheckpointType `json:"type"`
	Lat       float64        `json:"lat"`
	Long      float64        `json:"long"`
	Notes     string         `json:"notes,omitempty"`
	Timestamp time.Time      `json:"timestamp,omitempty"`
}

type ShipmentCreatedReq struct {
	OrderID         int          `json:"order_id"`
	DeliveryStaffID int          `json:"delivery_staff_id"`
	EstimatedTime   time.Time    `json:"estimated_time"`
	Checkpoints     []Checkpoint `json:"checkpoints"`
}

func (ct *CheckpointType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch CheckpointType(s) {
	case CheckpointStart, CheckpointOnRoad, CheckpointEnd:
		*ct = CheckpointType(s)
		return nil
	default:
		return fmt.Errorf("invalid checkpoint type: %s", s)
	}
}

func (ct CheckpointType) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ct))
}

