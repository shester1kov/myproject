package controllers

import (
	"context"
	"log"
	"net/http"
	"project/models"
	"project/services"
	"project/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetProductsByPriceRange godoc
// @Summary Получение продуктов по диапазону цен
// @Description Возвращает список продуктов, цены которых находятся в заданном диапазоне
// @Tags products
// @Accept  json
// @Produce  json
// @Param        Authorization header string true "токен"
// @Param        minPrice query number true "Минимальная цена"
// @Param        maxPrice query number true "Максимальная цена"
// @Success 200 {array} models.Product "Список продуктов в заданном диапазоне цен"
// @Failure 400 {object} models.ErrorResponse "Некорректные значения цен"
// @Failure 404 {object} models.ErrorResponse "Продукты не найдены в указанном диапазоне"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products/price-range [get]
func GetProductsByPriceRange(c *gin.Context) {
	minPrice, err1 := strconv.ParseFloat(c.Query("minPrice"), 64)
	maxPrice, err2 := strconv.ParseFloat(c.Query("maxPrice"), 64)

	if err1 != nil || err2 != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid price range values")
		return
	}

	var products []models.Product
	if err := services.DB.Where("price BETWEEN ? AND ?", minPrice, maxPrice).Find(&products).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error fetching products")
		return
	}

	if len(products) == 0 {
		utils.HandleError(c, http.StatusNotFound, "No products found in the given price range")
		return
	}

	c.JSON(http.StatusOK, products)

}

// UpdateProductsManufacturer godoc
// @Summary Массовое обновление производителя продуктов
// @Description Обновляет поле "manufacturer" у всех продуктов в базе данных на указанное значение.
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "токен"
// @Param manufacturer query string true "Новое значение для производителя"
// @Success 200 {object} models.MessageResponse "Успешное обновление"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера или транзакции"
// @Router /products/manufacturer [put]
func UpdateProductsManufacturer(c *gin.Context) {
	manufacturer := c.Query("manufacturer")

	if manufacturer == "" {
		utils.HandleError(c, http.StatusBadRequest, "manufacturer query parameter is required")
		return
	}

	// начало транзакции
	tx := services.DB.Begin()

	// проверяем, что транзакция инициализирована корректно
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		utils.HandleError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	log.Println("Transaction started successfully.")

	// попытка массового обновления
	if err := tx.Model(&models.Product{}).Where("1 = 1").Update("manufacturer", manufacturer).Error; err != nil {
		tx.Rollback() // откатываем изменения при ошибке
		log.Println("Error during update operation:", err)
		utils.HandleError(c, http.StatusInternalServerError, "Error updating manufacturer: "+err.Error())
		return
	}
	log.Println("Manufacturer update operation successful.")

	// коммит транзакции
	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		utils.HandleError(c, http.StatusInternalServerError, "Transaction commit failed: "+err.Error())
		return
	}
	log.Println("Transaction committed successfully.")

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Manufacturer updated successfully",
	})
}

// CountProductsByManufacturer godoc
// @Summary Подсчет количества продуктов по производителям
// @Description Выполняет агрегацию, подсчитывая количество продуктов, сгруппированных по производителям.
// @Tags products
// @Accept json
// @Produce json
// @Param Authorization header string true "токен"
// @Success 200 {array} models.CountProdutsResponse "Результаты агрегации с производителями и их количеством"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products/count-by-manufacturer [get]
func CountProductsByManufacturer(c *gin.Context) {
	var result []models.CountProdutsResponse

	// Выполняем агрегацию по производителю и подсчитываем количество товаров
	if err := services.DB.Model(&models.Product{}).
		Select("manufacturer, COUNT(*) as count").
		Group("manufacturer").
		Scan(&result).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error counting products by manufacturer: "+err.Error())
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, result)
}



// GetProductsWithTimeout godoc
// @Summary Получение списка продуктов с тайм-аутом
// @Description Получает список продуктов с применением фильтров, сортировки и пагинации с тайм-аутом в 2 секунды
// @Tags products
// @Accept  json
// @Produce  json
// @Param        Authorization header string true "токен"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Param sort query string false "Поле для сортировки" default(id)
// @Param order query string false "Направление сортировки" default(asc)
// @Param name query string false "Название продукта" 
// @Param category_id query string false "ID категории" 
// @Success 200 {object} models.ProductResponse "Успешный запрос"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 404 {object} models.ErrorResponse "Продукты не найдены"
// @Failure 408 {object} models.ErrorResponse "Тайм-аут запроса"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products [get]
func GetProductsWithTimeout(c *gin.Context) {
	// Создаем контекст с тайм-аутом 2 секунды
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var products []models.Product
	var total int64

	// Получаем параметры фильтров, сортировки и пагинации
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")
	sort := c.DefaultQuery("sort", "id")
	order := c.DefaultQuery("order", "asc")
	name := c.Query("name")
	categoryID := c.Query("category_id")

	// Преобразуем строковые параметры в int
	pageInt, _ := strconv.Atoi(page)
	limitInt, _ := strconv.Atoi(limit)
	offset := (pageInt - 1) * limitInt

	query := services.DB.Model(&models.Product{})

	// Применяем фильтры
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	query.Count(&total)

	// Применяем сортировку
	if order != "asc" && order != "desc" {
		order = "asc" // По умолчанию ascending
	}
	query = query.Order(sort + " " + order).Limit(limitInt).Offset(offset)

	// Загружаем продукты с использованием контекста
	if err := query.WithContext(ctx).Find(&products).Error; err != nil {
		if err == context.DeadlineExceeded {
			utils.HandleError(c, http.StatusRequestTimeout, "Request timed out")
		} else {
			utils.HandleError(c, http.StatusInternalServerError, "Failed to fetch products")
		}
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, models.ProductResponse{
		Data:  products,
		Total: int(total),
		Page:  pageInt,
		Limit: limitInt,
	})
}

// GetProductByID godoc
// @Summary Получение продукта по ID
// @Description Получает информацию о продукте по уникальному идентификатору
// @Tags products
// @Produce  json
// @Param        Authorization header string true "токен"
// @Param        id path string true "ID продукта"
// @Success 200 {object} models.Product "Успешный запрос"
// @Failure 404 {object} models.ErrorResponse "Продукт не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products/{id} [get]
func GetProductByID(c *gin.Context) {
	id := c.Param("id")
	var product models.Product
	if err := services.DB.First(&product, id).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Product not found")
		return
	}
	c.JSON(http.StatusOK, product)

}

// CreateProduct godoc
// @Summary Создание нового продукта
// @Description Создает новый продукт с указанными параметрами
// @Tags products
// @Accept  json
// @Produce  json
// @Param        Authorization header string true "токен"
// @Param        product body models.Product true "Данные продукта"
// @Success 201 {object} models.Product "Успешное создание"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products [post]
func CreateProduct(c *gin.Context) {
	var newProduct models.Product

	if err := c.BindJSON(&newProduct); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	var category models.Category
	if err := services.DB.First(&category, newProduct.CategoryID).Error; err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if newProduct.Price <= 0 {
		utils.HandleError(c, http.StatusBadRequest, "Price must be greater than 0")
		return
	}

	services.DB.Create(&newProduct)
	c.JSON(http.StatusCreated, newProduct)

}

// UpdateProduct godoc
// @Summary Обновление продукта
// @Description Обновляет данные продукта по указанному ID
// @Tags products
// @Accept  json
// @Produce  json
// @Param        Authorization header string true "токен"
// @Param        id path int true "ID продукта"
// @Param        product body models.Product true "Обновленные данные продукта"
// @Success 200 {object} models.Product "Успешное обновление"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 404 {object} models.ErrorResponse "Продукт не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products/{id} [put]
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	var updatedProduct models.Product

	if err := c.BindJSON(&updatedProduct); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	if updatedProduct.Price <= 0 {
		utils.HandleError(c, http.StatusBadRequest, "Price must be greater than 0")
		return
	}

	if err := services.DB.Model(&models.Product{}).Where("id = ?", id).Updates(updatedProduct).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Product not found")
		return
	}

	c.JSON(http.StatusOK, updatedProduct)
}

// DeleteProduct godoc
// @Summary Удаление продукта
// @Description Удаляет продукт по указанному ID
// @Tags products
// @Produce  json
// @Param        Authorization header string true "токен"
// @Param        id path string true "ID продукта"
// @Success 200 {object} models.MessageResponse "Успешное удаление продукта"
// @Failure 404 {object} models.ErrorResponse "Продукт не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Router /products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	if err := services.DB.Delete(&models.Product{}, id).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "Product not found")
		return
	}
	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "product deleted",
	})
}

