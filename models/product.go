package models

type Product struct {
	ID           int     `gorm:"primaryKey" json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	CategoryID   int     `json:"category_id"`
	Price        float64 `json:"price"`
	Manufacturer string  `json:"manufacturer"`
	Rating       float64 `json:"rating" grom:"default:0.0"`
}

type ProductInOrder struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
