package main

import (
	"project/controllers"
	_ "project/docs"
	"project/middlewares"
	"project/services"

	"github.com/gin-gonic/gin"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Sports Nutrition Store API
// @version         1.0
// @description     API для интернет-магазина спортивного питания. Содержит функционал для управления пользователями, продуктами, заказами и отзывами.
// @termsOfService  http://example.com/terms/

// @contact.name   Support Team
// @contact.url    http://example.com/support
// @contact.email  support@example.com

// @license.name  MIT License
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /
//
// @tag.name auth
// @tag.description Регистрация и авторизация
//
// @tag.name users
// @tag.description Операции, связанные с пользователями
//
// @tag.name products
// @tag.description Управление продуктами и отзывами
//
// @tag.name orders
// @tag.description Работа с заказами пользователей
//
// @tag.name categories
// @tag.description Управление категориями

func main() {
	services.InitDB()
	router := gin.Default()

	router.GET("/swagger/*any", gin.WrapF(httpSwagger.WrapHandler))

	router.POST("/login", controllers.Login)
	router.POST("/register", controllers.Register)
	router.POST("/refresh", controllers.Refresh)

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("/products/count-by-manufacturer", controllers.CountProductsByManufacturer)
		protected.GET("/products/price-range", controllers.GetProductsByPriceRange)
		protected.PUT("/products/manufacturer", middlewares.RoleMiddleware("admin"), controllers.UpdateProductsManufacturer)

		protected.GET("/products", controllers.GetProductsWithTimeout)
		protected.GET("/products/:id", controllers.GetProductByID)
		protected.POST("/products", middlewares.RoleMiddleware("admin"), controllers.CreateProduct)
		protected.PUT("/products/:id", middlewares.RoleMiddleware("admin"), controllers.UpdateProduct)
		protected.DELETE("/products/:id", middlewares.RoleMiddleware("admin"), controllers.DeleteProduct)
		protected.POST("/products/:id/reviews", controllers.CreateReview)
		router.GET("/products/:id/reviews", controllers.GetProductReviews)

		protected.GET("/categories", controllers.GetCategoriesWithTimeout)
		protected.GET("/categories/:id", controllers.GetCategoryByID)
		protected.POST("/categories", middlewares.RoleMiddleware("admin"), controllers.CreateCategory)
		protected.PUT("/categories/:id", middlewares.RoleMiddleware("admin"), controllers.UpdateCategory)
		protected.DELETE("/categories/:id", middlewares.RoleMiddleware("admin"), controllers.DeleteCategory)

		protected.GET("/orders", controllers.GetUserOrders)
		protected.GET("/orders/:id", controllers.GetOrderByID)
		protected.POST("orders/:id/products", controllers.AddProductToOrder)
		protected.POST("/orders", controllers.CreateOrder)
		protected.PATCH("orders/:id/products/:product_id", controllers.UpdateProductQuantity)
		protected.DELETE("/orders/:id/products/:product_id", controllers.DeleteProductFromOrder)
		protected.DELETE("/orders/:id", controllers.DeleteOrder)
		protected.GET("/admin/orders", middlewares.RoleMiddleware("admin"), controllers.GetAllOrders)
		protected.DELETE("/admin/orders/:id", middlewares.RoleMiddleware("admin"), controllers.DeleteOrderAdmin)

		protected.GET("users/me", controllers.GetUserInfo)
		protected.DELETE("users/me", controllers.DeleteSelf)
		protected.PATCH("users/me/username", controllers.UpdateUserName)
		protected.PATCH("users/me/password", controllers.UpdateUserPassword)
		protected.PATCH("/users/:id/role", middlewares.RoleMiddleware("admin"), controllers.UpdateUserRole)
		protected.DELETE("/users/:id", middlewares.RoleMiddleware("admin"), controllers.DeleteUser)
		protected.GET("/users", middlewares.RoleMiddleware("admin"), controllers.GetAllUsers)
		protected.GET("/users/:id", middlewares.RoleMiddleware("admin"), controllers.GetUserByID)
	}

	router.Run(":8080")
}
