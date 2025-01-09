package models

type ProductResponse struct {
	Data  []Product `json:"data"`
	Total int64       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

type OrderResponse struct {
	Data  []Order `json:"data"`
	Total int64       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type CountProdutsResponse struct {
	Manufacturer string `json:"manufacturer"`
	Count        int    `json:"count"`
}

type UserInfoResponse struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
}