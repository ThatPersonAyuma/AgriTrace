package core

type SubTopic string
const (
	OrderCreated = "Created"
	OrderPaid = "Paid"
	OrderCancelled = "Canceled"
	OrderConfirmedByFarmer = "Confirmed"
	OrderPrepared = "Prepared"
	OrderShipped = "Shipped"
	OrderDelivered = "Delivered"
	OrderCompleted = "Completed"
)
type OrderIDReq struct{
	OrderID int `json:"order_id"`
}
type OrderCreatedReq struct{
	BuyerID int 	`json:"buyer_id"`
	// StaffID int 	`json:"_staff_id"`
	
}

type OrderCoordinateReq struct{
	OrderID int `json:"order_id"`
	Lat float64 `json:"lat"`
	Long float64 `json:"long"`
}