package controllers

import (
	"context"
	"net/http"
	"project/models"
	"project/services"
	"project/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCategoriesWithTimeout godoc
// @Summary Получение списка категорий с тайм-аутом
// @Description Возвращает список категорий с предзагрузкой связанных продуктов, ограничивая время выполнения запроса до 2 секунд.
// @Tags categories
// @Accept json
// @Produce json
// @Param Authorization header string true "токен"
// @Success 200 {array} models.Category "Список категорий с предзагруженными продуктами"
// @Failure 408 {object} models.ErrorResponse "Тайм-аут запроса"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /categories [get]
func GetCategoriesWithTimeout(c *gin.Context) {
	// Создаем контекст с тайм-аутом 2 секунды
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var categories []models.Category
	if err := services.DB.WithContext(ctx).Preload("Products").Find(&categories).Error; err != nil {
		if err == context.DeadlineExceeded {
			utils.HandleError(c, http.StatusRequestTimeout, "Request timed out")
		} else {
			utils.HandleError(c, http.StatusInternalServerError, "Failed to fetch categories")
		}
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryByID godoc
// @Summary Получение категории по ID
// @Description Возвращает информацию о категории на основе переданного идентификатора.
// @Tags categories
// @Accept json
// @Produce json
// @Param Authorization header string true "токен"
// @Param id path int true "Идентификатор категории"
// @Success 200 {object} models.Category "Информация о категории"
// @Failure 404 {object} models.ErrorResponse "Категория не найдена"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /categories/{id} [get]
func GetCategoryByID(c *gin.Context) {
	id := c.Param("id")
	var category models.Category
	if err := services.DB.Preload("Products").First(&category, id).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Category not found")
		return
	}
	c.JSON(http.StatusOK, category)
}

// CreateCategory godoc
// @Summary Создание новой категории
// @Description Создает новую категорию в базе данных на основе переданных данных.
// @Tags categories
// @Accept json
// @Produce json
// @Param Authorization header string true "токен"
// @Param category body models.Category true "Данные категории"
// @Success 201 {object} models.Category "Созданная категория"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /categories [post]
func CreateCategory(c *gin.Context) {
	var newCategory models.Category
	if err := c.BindJSON(&newCategory); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := services.DB.Create(&newCategory).Error; err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request")
		return
	}
	c.JSON(http.StatusCreated, newCategory)
}

// UpdateCategory godoc
// @Summary Обновление категории
// @Description Обновляет категорию с переданными данными на основе ID
// @Tags categories
// @Accept json
// @Produce json
// @Param Authorization header string true "токен"
// @Param id path int true "Идентификатор категории"
// @Param category body models.Category true "Обновленные данные категории"
// @Success 200 {object} models.Category "Категория успешно обновлена"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 404 {object} models.ErrorResponse "Категория не найдена"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var updatedCategory models.Category
	if err := c.BindJSON(&updatedCategory); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// Проверяем, существует ли категория с этим ID
	var category models.Category
	if err := services.DB.First(&category, id).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Category not found")
		return
	}

	// Обновляем категорию
	if err := services.DB.Model(&category).Updates(updatedCategory).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Failed to update category")
		return
	}

	c.JSON(http.StatusOK, updatedCategory)
}

// DeleteCategory godoc
// @Summary Удаление категории
// @Description Удаляет категорию по переданному ID
// @Tags categories
// @Accept  json
// @Produce  json
// @Param Authorization header string true "токен"
// @Param id path int true "Идентификатор категории"
// @Success 200 {object} models.MessageResponse "Категория успешно удалена"
// @Failure 404 {object} models.ErrorResponse "Категория не найдена"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := services.DB.Delete(&models.Category{}, id).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Category not found")
		return
	}
	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "category deleted",
	})
}
