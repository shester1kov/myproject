package controllers

import (
	"log"
	"net/http"
	"project/models"
	"project/services"
	"project/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserInfo godoc
// @Summary Получение информации о пользователе
// @Description Получает информацию о текущем пользователе, включая его имя и роль. Пароль в ответе не передается.
// @Tags users
// @Accept  json
// @Produce  json
// @Param        Authorization  header  string  true  "Токен пользователя"
// @Success 200 {object} models.UserInfoResponse "Успешный запрос. Информация о пользователе"
// @Failure 401 {object} models.ErrorResponse "Неавторизованный доступ"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Security BearerAuth
// @Router /users/me [get]
func GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user models.User
	if err := services.DB.First(&user, userID).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "User not found")
		return
	}

	userInfoResponse := models.UserInfoResponse{
		Name: user.Username,
		Role: user.Role,
	}

	c.JSON(http.StatusOK, userInfoResponse)
}

// UpdateUserName godoc
// @Summary Обновление имени пользователя
// @Description Позволяет авторизованному пользователю обновить свое имя
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string false "Токен авторизации"
// @Param request body models.UpdateUsernameRequest true "Данные для обновления имени пользователя"
// @Success 200 {object} models.MessageResponse "Имя пользователя успешно обновлено"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 409 {object} models.ErrorResponse "Имя пользователя уже занято"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /users/me/username [patch]
func UpdateUserName(c *gin.Context) {
	var request models.UpdateUsernameRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var existingUser models.User
	if err := services.DB.Where("username = ?", request.Username).First(&existingUser).Error; err == nil {
		utils.HandleError(c, http.StatusConflict, "Username already taken")
		return
	}

	var user models.User
	if err := services.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "User not found")
		return
	}

	if len(request.Username) < 2 {
		utils.HandleError(c, http.StatusBadRequest, "Username length is less then 2")
		return
	}

	user.Username = request.Username
	if err := services.DB.Save(&user).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error updating user name")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "User name updated successfully",
	})
}

// UpdateUserPassword godoc
// @Summary Обновление пароля пользователя
// @Description Позволяет авторизованному пользователю изменить свой пароль, требуется указать старый и новый пароли
// @Tags users
// @Accept json
// @Produce json
// @Param Authorization header string false "Токен авторизации"
// @Param request body models.UpdatePasswordRequest true "Данные для обновления пароля"
// @Success 200 {object} models.MessageResponse "Пароль успешно обновлен"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} models.ErrorResponse "Пользователь не авторизован или старый пароль указан неверно"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /users/me/password [patch]
func UpdateUserPassword(c *gin.Context) {
	var request models.UpdatePasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user models.User
	if err := services.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "User not found")
		return
	}

	if !utils.CheckPassword(user.Password, request.OldPassword) {
		utils.HandleError(c, http.StatusUnauthorized, "Old password is incorrect")
		return
	}

	if len(request.NewPassword) < 6 {
		utils.HandleError(c, http.StatusBadRequest, "Password length is less than 6")
		return
	}

	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error hashing new password")
		return
	}

	user.Password = hashedPassword
	if err := services.DB.Save(&user).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error updating password")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Password updated successfully",
	})
}

// UpdateUserRole godoc
// @Summary Обновление роли пользователя на администратора
// @Description Позволяет администратору изменить роль пользователя только с "user" на "admin"
// @Tags users
// @Accept  json
// @Produce  json
// @Param Authorization header string false "Токен авторизации"
// @Param id path int true "ID пользователя"
// @Param data body models.UpdateUserRoleRequest true "Данные для обновления роли"
// @Success 200 {object} models.MessageResponse "Роль пользователя обновлена на администратора"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса или обновление роли невозможно"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /users/{id}/role [patch]
func UpdateUserRole(c *gin.Context) {
	var request models.UpdateUserRoleRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid request data")
		return
	}

	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := services.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "User not found")
		return
	}

	// Ограничение изменения роли только с "user" на "admin"
	if user.Role != "user" {
		utils.HandleError(c, http.StatusBadRequest, "Role can only be updated from 'user' to 'admin'")
		return
	}

	if request.Role != "admin" {
		utils.HandleError(c, http.StatusBadRequest, "Role can only be updated to 'admin'")
		return
	}

	// Обновление роли пользователя
	user.Role = "admin"
	if err := services.DB.Save(&user).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error updating user role")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "User role updated to admin successfully",
	})
}

// DeleteUser godoc
// @Summary Удаление пользователя с ролью "user"
// @Description Позволяет администратору удалить только пользователя с ролью "user"
// @Tags users
// @Accept  json
// @Produce  json
// @Param Authorization header string false "Токен авторизации"
// @Param id path int true "ID пользователя"
// @Success 200 {object} models.MessageResponse "Пользователь успешно удален"
// @Failure 400 {object} models.ErrorResponse "Некорректные данные запроса или удаление невозможно"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Проверка существования пользователя
	var user models.User
	if err := services.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "User not found")
		return
	}

	// Ограничение удаления только для пользователей с ролью "user"
	if user.Role != "user" {
		utils.HandleError(c, http.StatusBadRequest, "Only users with role 'user' can be deleted")
		return
	}

	tx := services.DB.Begin()

	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		utils.HandleError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	if err := tx.Where("order_id IN (SELECT id FROM orders WHERE user_id = ?)", userID).Delete(&models.OrderProduct{}).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := tx.Where("user_id = ?", userID).Delete(&models.Order{}).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Internal sever error")
		return
	}

	// Удаление пользователя
	if err := tx.Delete(&user).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error deleting user")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error deleting user and related data")
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "User and related data deleted successfully",
	})
}

// DeleteSelf godoc
// @Summary Удаление своей учетной записи
// @Description Позволяет пользователю удалить свою учетную запись. Администраторы не могут удалять себя.
// @Tags users
// @Accept  json
// @Produce  json
// @Param Authorization header string false "Токен авторизации"
// @Success 200 {object} models.MessageResponse "Учетная запись успешно удалена"
// @Failure 401 {object} models.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} models.ErrorResponse "Администратор не может удалить себя"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /users/me [delete]
func DeleteSelf(c *gin.Context) {
	// Получение идентификатора пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Проверяем, существует ли пользователь
	var user models.User
	if err := services.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "User not found")
		return
	}

	// Администратор не может удалить себя
	if user.Role == "admin" {
		utils.HandleError(c, http.StatusForbidden, "Administrators cannot delete themselves")
		return
	}

	tx := services.DB.Begin()

	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		utils.HandleError(c, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	if err := tx.Where("order_id IN (SELECT id FROM orders WHERE user_id = ?)", userID).Delete(&models.OrderProduct{}).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := tx.Where("user_id = ?", userID).Delete(&models.Order{}).Error; err != nil {
		tx.Rollback()
		utils.HandleError(c, http.StatusInternalServerError, "Internal sever error")
		return
	}

	// Удаление пользователя
	if err := tx.Delete(&user).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error deleting user")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error deleting user and related data")
		return
	}
	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Your account has been deleted successfully",
	})
}

// GetAllUsers godoc
// @Summary Получение списка всех пользователей
// @Description Возвращает данные всех пользователей.
// @Tags users
// @Accept  json
// @Produce  json
// @Param Authorization header string false "Токен авторизации"
// @Success 200 {array} models.User "Список пользователей"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /users [get]
func GetAllUsers(c *gin.Context) {
	var users []models.User

	if err := services.DB.Find(&users).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Error retrieving users")
		return
	}

	// Исключаем пароли из возвращаемых данных
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}

// GetUserByID godoc
// @Summary Получение данных пользователя по идентификатору
// @Description Возвращает данные конкретного пользователя по его ID.
// @Tags users
// @Accept  json
// @Produce  json
// @Param Authorization header string false "Токен авторизации"
// @Param id path int true "ID пользователя"
// @Success 200 {object} models.User "Данные пользователя"
// @Failure 400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка сервера"
// @Security BearerAuth
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	if err := services.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.HandleError(c, http.StatusNotFound, "User not found")
		return
	}

	// Исключаем пароль из возвращаемых данных
	user.Password = ""

	c.JSON(http.StatusOK, user)
}
