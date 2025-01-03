package middlewares

import (
	"net/http"
	"project/models"
	"project/services"
	"project/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		claims := &models.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return services.JwtKey, nil
		})

		if err != nil || !token.Valid {
			if err == jwt.ErrSignatureInvalid {
				utils.HandleError(c, http.StatusUnauthorized, "invalid token")
				c.Abort() // Прерываем обработку запроса
				return
			}

			// Обработка истёкшего токена
			if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors == jwt.ValidationErrorExpired {
				utils.HandleError(c, http.StatusUnauthorized, "token expired")
				c.Abort()
				return
			}

			utils.HandleError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
