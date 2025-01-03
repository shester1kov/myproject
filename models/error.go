package models

type ErrorResponse struct {
	Code    int    `json:"code"`    // Код ошибки, например, 400 или 500
	Message string `json:"message"` // Сообщение об ошибке
}
