package models

type OrderProduct struct {
	OrderID   int `gorm:"primaryKey" json:"order_id"`
	ProductID int `gorm:"primaryKey" json:"product_id"`
	Quantity  int `json:"quantity"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product"`
}
