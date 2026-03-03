package handler

import (
	"context"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/utils/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userService *service.UserService
	jwtManager  *auth.Manager
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"required,email"`
}

func NewAuthHandler(userService *service.UserService, jwtManager *auth.Manager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создает нового пользователя с логином, email и паролем
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} map[string]string "Пользователь успешно зарегистрирован"
// @Failure 400 {object} map[string]string "Ошибка валидации или пользователь уже существует"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	userID := uuid.New().String()
	ctx := context.Background()

	if err := h.userService.Create(ctx, userID, req.Username, req.Email, string(hash)); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := h.jwtManager.GenerateAccess(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(201, gin.H{
		"message":      "User registered successfully",
		"access_token": accessToken,
	})
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login godoc
// @Summary Авторизация пользователя
// @Description Аутентифицирует пользователя и возвращает access токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} map[string]string "Access токен"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Неверные учетные данные"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	user, err := h.userService.GetByUsername(ctx, req.Username)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	accessToken, err := h.jwtManager.GenerateAccess(user.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{
		"access_token": accessToken,
	})
}

// Refresh godoc
// @Summary Обновление access токена
// @Description Принимает refresh токен и возвращает новый access токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh токен"
// @Success 200 {object} map[string]string "Новый access токен"
// @Failure 400 {object} map[string]string "Ошибка валидации"
// @Failure 401 {object} map[string]string "Неверный refresh токен"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.jwtManager.Parse(req.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid refresh token"})
		return
	}

	accessToken, err := h.jwtManager.GenerateAccess(claims.UserID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{"access_token": accessToken})
}
