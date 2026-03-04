package handler

import (
	"context"
	"errors"
	"gw-currency-wallet/internal/repository"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/utils/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userService *service.UserService
	jwtManager  *auth.Manager
	logger      *zap.Logger
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"required,email"`
}

func NewAuthHandler(userService *service.UserService, jwtManager *auth.Manager, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
		logger:      logger,
	}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создает нового пользователя с логином, email и паролем
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Данные для регистрации"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid registration request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Username or email already exists"})
		return
	}

	h.logger.Info("registering new user", zap.String("username", req.Username), zap.String("email", req.Email))

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("failed to hash password", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	userID := uuid.New().String()
	ctx := context.Background()

	if err := h.userService.Create(ctx, userID, req.Username, req.Email, string(hash)); err != nil {
		if errors.Is(err, repository.ErrInvalidUUID) {
			h.logger.Warn("invalid UUID during registration", zap.Error(err))
			c.JSON(400, gin.H{"error": "Invalid user data"})
			return
		}
		h.logger.Error("failed to create user", zap.String("username", req.Username), zap.Error(err))
		c.JSON(400, gin.H{"error": "Username or email already exists"})
		return
	}

	accessToken, err := h.jwtManager.GenerateAccess(userID)
	if err != nil {
		h.logger.Error("failed to generate token", zap.String("user_id", userID), zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("user registered successfully", zap.String("user_id", userID), zap.String("username", req.Username))

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
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid login request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid username or password"})
		return
	}

	h.logger.Info("login attempt", zap.String("username", req.Username))

	ctx := context.Background()
	user, err := h.userService.GetByUsername(ctx, req.Username)
	if err != nil {
		h.logger.Warn("user not found", zap.String("username", req.Username))
		c.JSON(401, gin.H{"error": "Invalid username or password"})
		return
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	) != nil {
		h.logger.Warn("invalid password", zap.String("username", req.Username))
		c.JSON(401, gin.H{"error": "Invalid username or password"})
		return
	}

	accessToken, err := h.jwtManager.GenerateAccess(user.ID)
	if err != nil {
		h.logger.Error("failed to generate token", zap.String("user_id", user.ID), zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("login successful", zap.String("user_id", user.ID), zap.String("username", req.Username))

	c.JSON(200, gin.H{
		"token": accessToken,
	})
}

// Refresh godoc
// @Summary Обновление access токена
// @Description Принимает refresh токен и возвращает новый access токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh токен"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid refresh request", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.jwtManager.Parse(req.RefreshToken)
	if err != nil {
		h.logger.Warn("invalid refresh token", zap.Error(err))
		c.JSON(401, gin.H{"error": "invalid refresh token"})
		return
	}

	accessToken, err := h.jwtManager.GenerateAccess(claims.UserID)
	if err != nil {
		h.logger.Error("failed to generate token", zap.String("user_id", claims.UserID), zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("token refreshed", zap.String("user_id", claims.UserID))

	c.JSON(200, gin.H{"access_token": accessToken})
}
