package controllers

import (
	"net/http"
	"project/models"
	"project/services"
	"project/utils"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Login godoc
// @Summary      Авторизация пользователя
// @Description  Эндпоинт для авторизации пользователя. При успешной авторизации возвращает JWT-токен.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body models.Credentials true "Учетные данные пользователя"
// @Success      200 {object} models.TokenResponse "Возвращает jwt-токен"
// @Failure      400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure      401 {object} models.ErrorResponse "Некорректное имя пользователя"
// @Failure      401 {object} models.ErrorResponse "Некорректный пароль"
// @Failure      500 {object} models.ErrorResponse "Невозможно создать токен"
// @Router       /login [post]
func Login(c *gin.Context) {
	var creds models.Credentials
	if err := c.BindJSON(&creds); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "invalid request")
		return
	}

	// Ищем пользователя
	var user models.User
	if err := services.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
		utils.HandleError(c, http.StatusUnauthorized, "invalid username")
		return
	}

	// Проверяем пароль
	if !utils.CheckPassword(user.Password, creds.Password) {
		utils.HandleError(c, http.StatusUnauthorized, "invalid password")
		return
	}

	// Генерация токена с ролью пользователя
	token, err := services.GenerateToken(int(user.ID), user.Username, user.Role)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "could not create token")
		return
	}
	c.JSON(http.StatusOK, models.TokenResponse{
		Token: token,
	})
}

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Эндпоинт для регистрации нового пользователя. Возвращает сообщение об успешной регистрации.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials body models.Credentials true "Учетные данные пользователя (username, password, optional: role)"
// @Success      201 {object} models.MessageResponse "Пользователь успешно зарегистрирован"
// @Failure      400 {object} models.ErrorResponse "Некорректный запрос"
// @Failure      409 {object} models.ErrorResponse "Пользователь уже существует"
// @Failure      500 {object} models.ErrorResponse "Невозможно зарегистрировать пользователя"
// @Router       /register [post]
func Register(c *gin.Context) {
	var creds models.Credentials
	if err := c.BindJSON(&creds); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if len(creds.Username) < 2 {
		utils.HandleError(c, http.StatusBadRequest, "Username length is less than 2")
		return
	}

	if len(creds.Password) < 6 {
		utils.HandleError(c, http.StatusBadRequest, "Password length is less than 6")
		return
	}

	var existingUser models.User
	if err := services.DB.Where("username = ?", creds.Username).First(&existingUser).Error; err == nil {
		utils.HandleError(c, http.StatusConflict, "user already exists")
	}

	hashedPassword, err := utils.HashPassword(creds.Password)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "failed to register user")
	}

	// Регистрируем пользователя
	newUser := models.User{
		Username: creds.Username,
		Password: hashedPassword,
		Role:     "user",
	}

	if err := services.DB.Create(&newUser).Error; err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "failed to register user")
		return
	}
	c.JSON(http.StatusCreated, models.MessageResponse{
		Message: "user registered successfully",
	})
}

// Refresh godoc
// @Summary      Обновление токена
// @Description  Эндпоинт для обновления JWT токена. Генерирует новый токен, если исходный почти истек.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "токен"
// @Success      200 {object} models.TokenResponse "новый JWT токен"
// @Failure      400 {object} models.ErrorResponse "Токен еще не истек"
// @Failure      401 {object} models.ErrorResponse "Пользователь не авторизирован"
// @Failure      500 {object} models.ErrorResponse "Невозможно создать токен"
// @Router       /refresh [post]
func Refresh(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	claims := &models.Claims{}

	// Парсим исходный токен
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return services.JwtKey, nil
	})

	if err != nil || !token.Valid {
		utils.HandleError(c, http.StatusUnauthorized, "unauthorized")

		return
	}

	// Проверяем, не истек ли срок действия токена
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 120*time.Second {
		utils.HandleError(c, http.StatusBadRequest, "token not expired enough")
		return
	}

	// Генерация нового токена с теми же данными (пользователь и роль), но с новым временем истечения
	newToken, err := services.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "token not create token")
		return
	}
	c.JSON(http.StatusOK, models.TokenResponse{
		Token: newToken,
	})
}
