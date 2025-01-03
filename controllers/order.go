package controllers

import (
	"fmt"
	"net/http"
	"project/models"
	"project/services"
	"project/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateOrder godoc
// @Summary Создание нового заказа
// @Description Создает новый заказ и связывает с ним продукты. Если продукты не указаны, заказ будет создан без них.
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT токен пользователя"
// @Param request body models.CreateOrderRequest true "Данные для создания заказа"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse "Заказ успешно создан"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса или продукт не найден"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /orders [post]
func CreateOrder(c *gin.Context) {
	var request models.CreateOrderRequest

	// Чтение данных из запроса
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	// Получаем user_id из контекста (из JWT токена)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Создаем новый заказ
	order := models.Order{
		UserID: userID.(int),
	}

	tx := services.DB.Begin()

	if tx.Error != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error starting transaction")
		return
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Error creating order")
		return
	}

	if len(request.Products) > 0 {
		for _, p := range request.Products {
			var product models.Product
			if err := tx.First(&product, p.ProductID).Error; err != nil {
				tx.Rollback()
				utils.HandleError(c, http.StatusBadRequest, fmt.Sprintf("Product with ID %d not found", p.ProductID))
				return
			}

			orderProduct := models.OrderProduct{
				OrderID:   order.ID,
				ProductID: p.ProductID,
				Quantity:  p.Quantity,
			}

			if err := tx.Create(&orderProduct).Error; err != nil {
				utils.HandleError(c, http.StatusInternalServerError, "Error creating order product")
				return
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error committing transaction")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: fmt.Sprintf("Order created successfully. Order ID: %d", order.ID),
	})

}

// GetUserOrders godoc
// @Summary Получение списка заказов пользователя
// @Description Возвращает список заказов, связанных с пользователем, включая информацию о продуктах в заказах
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен доступа пользователя (JWT)"
// @Success 200 {array} models.Order "Список заказов с продуктами"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /orders [get]
func GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var orders []models.Order
	if err := services.DB.Preload("Products.Product").Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error fetching orders")
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetOrderByID godoc
// @Summary Получение информации о заказе по идентификатору
// @Description Возвращает данные заказа, включая связанные продукты, если заказ принадлежит авторизованному пользователю
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен доступа пользователя (JWT)"
// @Param id path int true "Идентификатор заказа"
// @Success 200 {object} models.Order "Информация о заказе с продуктами"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Заказ не найден"
// @Router /orders/{id} [get]
func GetOrderByID(c *gin.Context) {
	// Получение идентификатора заказа из параметров URL
	orderIDParam := c.Param("id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	// Получение идентификатора пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var order models.Order
	// Загрузка заказа с продуктами
	if err := services.DB.Preload("Products.Product").
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Order not found")
		return
	}

	// Возврат информации о заказе
	c.JSON(http.StatusOK, order)
}

// AddProductToOrder godoc
// @Summary Добавление продукта в заказ
// @Description Добавляет продукт в заказ текущего пользователя. Если продукт уже существует в заказе, его количество увеличивается.
// @Tags orders
// @Accept json
// @Produce json
// @Param        Authorization header string true "Токен пользователя"
// @Param        id path int true "ID заказа"
// @Param        product body models.ProductInOrder true "Продукт для добавления в заказ"
// @Success 200 {object} models.MessageResponse "Успешное добавление продукта"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Заказ не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка на сервере"
// @Router /orders/{id}/products [post]
func AddProductToOrder(c *gin.Context) {
    orderIDParam := c.Param("id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

    var request models.ProductInOrder
    if err := c.ShouldBindJSON(&request); err != nil {
        utils.HandleError(c, http.StatusBadRequest, "Invalid request data")
        return
    }

    // Получаем user_id из контекста
    userID, exists := c.Get("user_id")
    if !exists {
        utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
        return
    }

    // Проверяем, принадлежит ли заказ пользователю
    var order models.Order
    if err := services.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
        utils.HandleError(c, http.StatusNotFound, "Order not found")
        return
    }

	var orderProduct models.OrderProduct
	if err := services.DB.Where("order_id = ? AND product_id = ?", order.ID, request.ProductID).First(&orderProduct).Error; err == nil {
		// Если продукт найден, обновляем его количество
		orderProduct.Quantity += request.Quantity
		if err := services.DB.Save(&orderProduct).Error; err != nil {
			utils.HandleError(c, http.StatusInternalServerError, "Error updating product quantity")
			return
		}

		c.JSON(http.StatusOK, models.MessageResponse{
			Message: "Product quantity updated",
		})
		return
	}

    // Создаем новый OrderProduct
    orderProduct = models.OrderProduct{
        OrderID:   order.ID,
        ProductID: request.ProductID,
        Quantity:  request.Quantity,
    }

    if err := services.DB.Create(&orderProduct).Error; err != nil {
        utils.HandleError(c, http.StatusInternalServerError, "Error adding product to order")
        return
    }

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Product added to order",
	})
}

// UpdateProductQuantity godoc
// @Summary Обновление количества продукта в заказе
// @Description Обновляет количество указанного продукта в заказе текущего пользователя.
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен пользователя"
// @Param id path int true "ID заказа"
// @Param product_id path int true "ID продукта"
// @Param quantity body models.UpdateProductQuantityRequest true "Новое количество продукта"
// @Success 200 {object} models.MessageResponse "Успешное обновление количества продукта"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Продукт или заказ не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка на сервере"
// @Router /orders/{id}/products/{product_id} [patch]
func UpdateProductQuantity(c *gin.Context) {
	orderIDParam := c.Param("id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	productIDParam := c.Param("product_id")
	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var request models.UpdateProductQuantityRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	if request.Quantity <= 0 {
		utils.HandleError(c, http.StatusBadRequest, "Quantity must be greater than zero")
		return
	}

	// Получаем user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Проверяем, принадлежит ли заказ пользователю
	var order models.Order
	if err := services.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Order not found")
		return
	}

	// Проверяем, существует ли продукт в заказе
	var orderProduct models.OrderProduct
	if err := services.DB.Where("order_id = ? AND product_id = ?", order.ID, productID).First(&orderProduct).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Product not found in the order")
		return
	}

	// Обновляем количество
	orderProduct.Quantity = request.Quantity
	if err := services.DB.Save(&orderProduct).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error updating product quantity")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Product quantity updated successfully",
	})
}

// DeleteProductFromOrder godoc
// @Summary Удаление продукта из заказа
// @Description Удаляет указанный продукт из заказа текущего пользователя.
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен пользователя"
// @Param id path int true "ID заказа"
// @Param product_id path int true "ID продукта для удаления"
// @Success 200 {object} models.MessageResponse "Успешное удаление продукта"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Продукт или заказ не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка на сервере"
// @Router /orders/{id}/products/{product_id} [delete]
func DeleteProductFromOrder(c *gin.Context) {
	orderIDParam := c.Param("id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	productIDParam := c.Param("product_id")
	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Получаем user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Проверяем, принадлежит ли заказ пользователю
	var order models.Order
	if err := services.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Order not found")
		return
	}

	// Удаляем продукт из заказа
	if err := services.DB.Where("order_id = ? AND product_id = ?", order.ID, productID).Delete(&models.OrderProduct{}).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error deleting product from order")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Product removed from order successfully",
	})
}

// DeleteOrder godoc
// @Summary Удаление заказа
// @Description Удаляет указанный заказ текущего пользователя вместе с привязанными продуктами.
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен пользователя"
// @Param id path int true "ID заказа"
// @Success 200 {object} models.MessageResponse "Успешное удаление заказа"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Заказ не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка на сервере"
// @Router /orders/{id} [delete]
func DeleteOrder(c *gin.Context) {
	orderIDParam := c.Param("id")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	// Получаем user_id из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Проверяем, принадлежит ли заказ пользователю
	var order models.Order
	if err := services.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Order not found")
		return
	}

	// Удаление всех связанных продуктов
	if err := services.DB.Where("order_id = ?", order.ID).Delete(&models.OrderProduct{}).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error deleting order products")
		return
	}

	// Удаление самого заказа
	if err := services.DB.Delete(&order).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error deleting order")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Order deleted successfully",
	})
}
