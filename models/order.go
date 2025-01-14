package models

type Order struct {
	ID       int            `gorm:"primaryKey" json:"order_id"`
	UserID   int            `json:"user_id"`
	Products []OrderProduct `gorm:"foreignKey:OrderID" json:"products"`
	User     User           `json:"user" gorm:"foreignKey:UserID" swaggerignore:"true"`
}
