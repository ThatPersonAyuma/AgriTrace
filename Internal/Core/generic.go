package core

import (
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
	AddToCart			   = "AddItem"

	ProductListed   = "Listed"
	ProductUnlisted = "Unlisted"
	SerachProduct   = "Search"
	ProductCreated = "Created"
	FarmerProducts = "FarmerProducts"

	ShipmentCreated         = "ShipmentCreated"
	GetShipment             = "GetShipment"
	CheckpointAdded         = "CheckpointAdded"
	CheckpointPhotoUploaded = "CheckpointPhotoUploaded"
	CheckpointVerified      = "CheckpointVerified"
	ShipmentCompleted       = "ShipmentCompleted"
	ShipmentDelayed         = "ShipmentDelayed"
	GetShipmentWithImage	= "GetShipmentWImg"

	AccountCreated = "Created"
	AccountUpdated = "Updated"
)

type OrderIDReq struct {
	OrderID int `json:"order_id"`
}
type OrderCreatedReq struct {
	BuyerID int `json:"buyer_id"`
	TotalPrice float32 `json:"total_price"`
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

type GetOrderCheckpointsReq struct {
	OrderID int `json:"order_id"`
}
type CheckpointPhotoUploadReq struct {
	CheckpointID int    `json:"checkpoint_id"`
	Filename     string `json:"filename"`
	FileData     []byte `json:"file_data"` // Base64 encoded file data
}
type ProductCreatedReq struct{
	FarmerID	int		`json:"farmer_id"`
	Name		string	`json:"name"`
	Description	string	`json:"description"`
	Price		float64	`json:"price"`
	Stock		int		`json:"stock"`
	MinOrder	int		`json:"min_order"`
}
type FarmeIdReq struct{
	FarmerID	int		`json:"farmer_id"`
}
type AddToCartPayload struct {
	OrderID   int     `json:"order_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}