package models

type Order struct {
	OrderUID          string   `json:"order_uid" validate:"required,max=64"`
	TrackNumber       string   `json:"track_number" validate:"required,max=64"`
	Entry             string   `json:"entry" validate:"required,max=32"`
	Delivery          Delivery `json:"delivery" validate:"required"`
	Payment           Payment  `json:"payment" validate:"required"`
	Items             []Item   `json:"items" validate:"required,min=1"`
	Locale            string   `json:"locale" validate:"omitempty,max=8"`
	InternalSignature string   `json:"internal_signature" validate:"max=256"`
	CustomerID        string   `json:"customer_id" validate:"required,max=128"`
	DeliveryService   string   `json:"delivery_service" validate:"required,max=128"`
	ShardKey          string   `json:"shardkey" validate:"omitempty,max=16"`
	SmID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	OofShard          string   `json:"oof_shard" validate:"omitempty,max=8"`
}

type Delivery struct {
	Name    string `json:"name" validate:"required,max=128"`
	Phone   string `json:"phone" validate:"required,max=32"`
	Zip     string `json:"zip" validate:"max=32"`
	City    string `json:"city" validate:"required,max=128"`
	Address string `json:"address" validate:"required,max=256"`
	Region  string `json:"region" validate:"max=128"`
	Email   string `json:"email" validate:"omitempty,email,max=128"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required,max=128"`
	RequestID    string `json:"request_id" validate:"max=128"`
	Currency     string `json:"currency" validate:"required,oneof=USD EUR RUB UAH KZT BYN PLN GBP"`
	Provider     string `json:"provider" validate:"required,max=128"`
	Amount       int    `json:"amount" validate:"required,min=0"`
	PaymentDT    int64  `json:"payment_dt" validate:"required,min=0"`
	Bank         string `json:"bank" validate:"max=128"`
	DeliveryCost int    `json:"delivery_cost" validate:"min=0"`
	GoodsTotal   int    `json:"goods_total" validate:"min=0"`
	CustomFee    int    `json:"custom_fee" validate:"min=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required,min=0"`
	TrackNumber string `json:"track_number" validate:"required,max=64"`
	Price       int    `json:"price" validate:"required,min=0"`
	RID         string `json:"rid" validate:"required,max=128"`
	Name        string `json:"name" validate:"required,max=256"`
	Sale        int    `json:"sale" validate:"min=0"`
	Size        string `json:"size" validate:"max=16"`
	TotalPrice  int    `json:"total_price" validate:"required,min=0"`
	NmID        int    `json:"nm_id" validate:"min=0"`
	Brand       string `json:"brand" validate:"max=128"`
	Status      int    `json:"status" validate:"min=0"`
}
