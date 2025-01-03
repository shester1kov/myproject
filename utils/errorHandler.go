package utils

import (
	"project/models"

	"github.com/gin-gonic/gin"
)

func HandleError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Code:		statusCode,
		Message:	message,
	})
}
