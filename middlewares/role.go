package middlewares

import (
	"net/http"
	"project/models"
	"project/services"
	"project/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		claims := &models.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return services.JwtKey, nil
		})

		if err != nil || !token.Valid {
			utils.HandleError(c, http.StatusUnauthorized, "unauthorized")

			c.Abort()
			return
		}

		if claims.Role != requiredRole {
			utils.HandleError(c, http.StatusForbidden, "forbidden")

			c.Abort()
			return
		}

		c.Next()
	}
}
