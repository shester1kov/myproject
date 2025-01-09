package models

type CreateOrderRequest struct {
	Products []ProductInOrder `json:"products,omitempty"` // Опциональный список продуктов
}

type UpdateProductQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

type CreateReviewRequest struct {
	ReviewText string `json:"review_text"`
	Rating     int    `json:"rating"`
}
