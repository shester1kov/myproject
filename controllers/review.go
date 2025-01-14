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

// CreateReview godoc
// @Summary Создание нового отзыва
// @Description Создает новый отзыв.
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "JWT токен пользователя"
// @Param id path string true "ID продукта"
// @Param request body models.CreateReviewRequest true "Данные для создания отзыва"
// @Security BearerAuth
// @Success 200 {object} models.MessageResponse "Отзыв успешно создан"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products/{id}/reviews [post]
func CreateReview(c *gin.Context) {
	productIDParam := c.Param("id")
	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var product models.Product

	if err := services.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		utils.HandleError(c, http.StatusBadRequest, fmt.Sprintf("Product with ID %d not found", productID))
		return
	}

	var request models.CreateReviewRequest

	if err := c.BindJSON(&request); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	if request.Rating > 5 || request.Rating < 1 {
		utils.HandleError(c, http.StatusBadRequest, "Invalid rating")
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unathorized")
		return
	}

	var existingReview models.Review

	if err := services.DB.Where("product_id = ? AND user_id = ?", productID, userID).First(&existingReview).Error; err == nil {
		utils.HandleError(c, http.StatusBadRequest, "You already have review")
		return
	}

	review := models.Review{
		ReviewText: request.ReviewText,
		Rating:     request.Rating,
		UserID:     userID.(int),
		ProductID:  productID,
	}

	tx := services.DB.Begin()

	if tx.Error != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error starting transaction")
		return
	}

	if err := tx.Create(&review).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Error creating review")
		return
	}

	var newRating float64

	if err := tx.Model(&models.Review{}).Select("AVG(rating) as rating").Group("product_id").Where("product_id = ?", productID).Scan(&newRating).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Error getting new rating")
		return
	}

	product.Rating = newRating

	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Error updating rating")
	}

	if err := tx.Commit().Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error committing transaction") //?
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: fmt.Sprintf("Review created successfully. Review ID: %d", review.ID),
	})
}

// @Summary Get product reviews
// @Description Get all reviews for a specific product
// @Tags Reviews
// @Param id path int true "Product ID"
// @Success 200 {object} map[string][]models.Review
// @Failure 400 {object} models.MessageResponse
// @Failure 500 {object} models.MessageResponse
// @Router /products/{id}/reviews [get]
func GetProductReviews(c *gin.Context) {
	// Получаем идентификатор товара из параметров запроса
	productIDParam := c.Param("id")
	productID, err := strconv.Atoi(productIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Массив для хранения отзывов
	var reviews []models.Review

	// Запрашиваем отзывы из базы данных
	if err := services.DB.Where("product_id = ?", productID).Find(&reviews).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error fetching reviews")
		return
	}

	c.JSON(http.StatusOK, reviews)
}
