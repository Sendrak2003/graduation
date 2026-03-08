package handler

import (
	"errors"
	"gw-currency-wallet/internal/repository"
	"gw-currency-wallet/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ExchangeHandler struct {
	service *service.ExchangeService
	logger  *zap.Logger
}

func NewExchangeHandler(service *service.ExchangeService, logger *zap.Logger) *ExchangeHandler {
	return &ExchangeHandler{
		service: service,
		logger:  logger,
	}
}

type ExchangeRequest struct {
	FromCurrency string  `json:"from_currency" binding:"required,oneof=USD EUR RUB"`
	ToCurrency   string  `json:"to_currency" binding:"required,oneof=USD EUR RUB"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
}

// GetExchangeRates godoc
// @Summary Получить курсы валют
// @Description Возвращает актуальные курсы обмена всех валют
// @Tags exchange
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /exchange/rates [get]
func (h *ExchangeHandler) GetExchangeRates(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("unauthorized access attempt")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	rates, err := h.service.GetRates(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get exchange rates", zap.Error(err))
		c.JSON(503, gin.H{"error": "Exchange service temporarily unavailable"})
		return
	}

	c.JSON(200, gin.H{
		"rates": rates,
	})
}

// Exchange godoc
// @Summary Обменять валюту
// @Description Обменивает одну валюту на другую по текущему курсу
// @Tags exchange
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ExchangeRequest true "Данные для обмена"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /exchange [post]
func (h *ExchangeHandler) Exchange(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Warn("unauthorized exchange attempt")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid exchange request", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	if req.FromCurrency == req.ToCurrency {
		h.logger.Warn("same currency exchange attempt")
		c.JSON(400, gin.H{"error": "Cannot exchange same currency"})
		return
	}

	userIDStr := userID.(string)

	exchangedAmount, err := h.service.Exchange(
		c.Request.Context(),
		userIDStr,
		req.FromCurrency,
		req.ToCurrency,
		req.Amount,
	)

	if err != nil {
		if errors.Is(err, repository.ErrInvalidUUID) {
			h.logger.Warn("invalid user UUID for exchange", zap.String("user_id", userIDStr), zap.Error(err))
			c.JSON(400, gin.H{"error": "Invalid user ID"})
			return
		}
		if errors.Is(err, repository.ErrWalletNotFound) || errors.Is(err, repository.ErrInsufficientFunds) {
			h.logger.Warn("exchange failed",
				zap.String("user_id", userIDStr),
				zap.Error(err))
			c.JSON(400, gin.H{"error": "Insufficient funds or invalid currencies"})
			return
		}
		h.logger.Error("exchange failed",
			zap.String("user_id", userIDStr),
			zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to process exchange"})
		return
	}

	h.logger.Info("exchange successful",
		zap.String("user_id", userIDStr),
		zap.String("from", req.FromCurrency),
		zap.String("to", req.ToCurrency),
		zap.Float64("amount", req.Amount),
		zap.Float64("exchanged_amount", exchangedAmount))

	c.JSON(200, gin.H{
		"message":          "Exchange successful",
		"exchanged_amount": exchangedAmount,
		"new_balance": gin.H{
			req.FromCurrency: "updated",
			req.ToCurrency:   "updated",
		},
	})
}
