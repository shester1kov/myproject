package models

type User struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex" json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
