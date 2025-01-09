package controllers

import (
	"net/http"
	"project/models"
	"project/services"
	"project/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllOrders godoc
// @Summary Получение списка всех заказов
// @Description Возвращает список заказов, включая информацию о продуктах в заказах
// @Tags orders
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен доступа пользователя (JWT)"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Param sort query string false "Поле для сортировки" default(id)
// @Param order query string false "Направление сортировки" default(asc)
// @Param user_id query string false "ID пользователя" 
// @Success 200 {array} models.OrderResponse "Список заказов с продуктами"
// @Failuer 400 {object} models.ErrorResponse "Некорректные данные"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /admin/orders [get]
func GetAllOrders(c *gin.Context) {
	var orders []models.Order
	var total int64

	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	sort := c.DefaultQuery("sort", "id")
	order := c.DefaultQuery("order", "asc")
	user_id := c.Query("user_id")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Incorrect page number")
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Incorrect limit")
	}
	offset := (pageInt - 1) * limitInt

	query := services.DB.Model(&models.Order{})

	if user_id != "" {
		query = query.Where("user_id = ?", user_id)
	}

	query.Count(&total)

	if order != "asc" && order != "desc" {
		order = "asc" 
	}
	query = query.Order(sort + " " + order).Limit(limitInt).Offset(offset)




	if err := query.Preload("Products.Product").Find(&orders).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error fetching orders")
		return
	}

	c.JSON(http.StatusOK, models.OrderResponse{
		Data: orders,
		Total: total,
		Page: pageInt,
		Limit: limitInt,
	})
}
