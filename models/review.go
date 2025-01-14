package models

type Review struct {
	ID         int     `gorm:"primaryKey" json:"id"`
	ReviewText string  `json:"review_text"`
	Rating     int     `json:"rating"`
	UserID     int     `json:"user_id" gorm:"foreignKey:UserID"`
	ProductID  int     `json:"product_id" gorm:"foreignKey:ProductID"`
	Product    Product `json:"product" gorm:"foreignKey:ProductID" swaggerignore:"true"`
	User       User    `json:"user" gorm:"foreignKey:UserID" swaggerignore:"true"`
}
