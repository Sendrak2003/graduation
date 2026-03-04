package handler

import (
	"errors"
	"gw-currency-wallet/internal/repository"
	"gw-currency-wallet/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WalletHandler struct {
	service *service.WalletService
	logger  *zap.Logger
}

func NewWalletHandler(service *service.WalletService, logger *zap.Logger) *WalletHandler {
	return &WalletHandler{
		service: service,
		logger:  logger,
	}
}

type DepositRequest struct {
	Currency string  `json:"currency" binding:"required,oneof=USD EUR RUB"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}

type WithdrawRequest struct {
	Currency string  `json:"currency" binding:"required,oneof=USD EUR RUB"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}

// GetBalance godoc
// @Summary Получить баланс кошелька
// @Description Возвращает баланс всех валют в кошельке пользователя
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /balance [get]
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("unauthorized access attempt")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr := userID.(string)
	h.logger.Info("getting balance", zap.String("user_id", userIDStr))

	balances, err := h.service.GetAllBalances(c.Request.Context(), userIDStr)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidUUID) {
			h.logger.Warn("invalid user UUID", zap.String("user_id", userIDStr), zap.Error(err))
			c.JSON(400, gin.H{"error": "Invalid user ID"})
			return
		}
		h.logger.Error("failed to get balance", zap.String("user_id", userIDStr), zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to retrieve balance"})
		return
	}

	h.logger.Info("balance retrieved successfully", zap.String("user_id", userIDStr))
	c.JSON(200, gin.H{
		"balance": balances,
	})
}

// Deposit godoc
// @Summary Пополнить кошелек
// @Description Пополняет баланс кошелька указанной валютой
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DepositRequest true "Данные для пополнения"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /wallet/deposit [post]
func (h *WalletHandler) Deposit(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("unauthorized deposit attempt")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid deposit request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid amount or currency"})
		return
	}

	userIDStr := userID.(string)
	h.logger.Info("processing deposit",
		zap.String("user_id", userIDStr),
		zap.String("currency", req.Currency),
		zap.Float64("amount", req.Amount))

	err := h.service.Deposit(c.Request.Context(), userIDStr, req.Currency, req.Amount)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidUUID) {
			h.logger.Warn("invalid user UUID for deposit", zap.String("user_id", userIDStr), zap.Error(err))
			c.JSON(400, gin.H{"error": "Invalid user ID"})
			return
		}
		h.logger.Error("deposit failed",
			zap.String("user_id", userIDStr),
			zap.String("currency", req.Currency),
			zap.Float64("amount", req.Amount),
			zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to process deposit"})
		return
	}

	balances, err := h.service.GetAllBalances(c.Request.Context(), userIDStr)
	if err != nil {
		h.logger.Error("failed to get updated balance after deposit", zap.String("user_id", userIDStr), zap.Error(err))
		c.JSON(200, gin.H{
			"message": "Account topped up successfully",
		})
		return
	}

	h.logger.Info("deposit successful",
		zap.String("user_id", userIDStr),
		zap.String("currency", req.Currency),
		zap.Float64("amount", req.Amount))

	c.JSON(200, gin.H{
		"message":     "Account topped up successfully",
		"new_balance": balances,
	})
}

// Withdraw godoc
// @Summary Снять средства с кошелька
// @Description Снимает указанную сумму в указанной валюте с кошелька
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body WithdrawRequest true "Данные для снятия"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /wallet/withdraw [post]
func (h *WalletHandler) Withdraw(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("unauthorized withdraw attempt")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid withdraw request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid amount or currency"})
		return
	}

	userIDStr := userID.(string)
	h.logger.Info("processing withdrawal",
		zap.String("user_id", userIDStr),
		zap.String("currency", req.Currency),
		zap.Float64("amount", req.Amount))

	err := h.service.Withdraw(c.Request.Context(), userIDStr, req.Currency, req.Amount)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidUUID) {
			h.logger.Warn("invalid user UUID for withdraw", zap.String("user_id", userIDStr), zap.Error(err))
			c.JSON(400, gin.H{"error": "Invalid user ID"})
			return
		}
		if errors.Is(err, repository.ErrWalletNotFound) || errors.Is(err, repository.ErrInsufficientFunds) {
			h.logger.Warn("withdrawal failed",
				zap.String("user_id", userIDStr),
				zap.String("currency", req.Currency),
				zap.Float64("amount", req.Amount),
				zap.Error(err))
			c.JSON(400, gin.H{"error": "Insufficient funds or invalid amount"})
			return
		}
		h.logger.Error("withdrawal failed",
			zap.String("user_id", userIDStr),
			zap.String("currency", req.Currency),
			zap.Float64("amount", req.Amount),
			zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to process withdrawal"})
		return
	}

	balances, err := h.service.GetAllBalances(c.Request.Context(), userIDStr)
	if err != nil {
		h.logger.Error("failed to get updated balance after withdrawal", zap.String("user_id", userIDStr), zap.Error(err))
		c.JSON(200, gin.H{
			"message": "Withdrawal successful",
		})
		return
	}

	h.logger.Info("withdrawal successful",
		zap.String("user_id", userIDStr),
		zap.String("currency", req.Currency),
		zap.Float64("amount", req.Amount))

	c.JSON(200, gin.H{
		"message":     "Withdrawal successful",
		"new_balance": balances,
	})
}
