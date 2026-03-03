package http

import (
	"gw-currency-wallet/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	service *service.WalletService
}

func NewWalletHandler(service *service.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

type BalanceRequest struct {
	Currency string `json:"currency" binding:"required,oneof=USD EUR RUB"`
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
// @Description Возвращает баланс указанной валюты в кошельке пользователя
// @Tags wallet
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param currency query string true "Валюта" Enums(USD, EUR, RUB)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /balance [get]
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	currency := c.Query("currency")
	if currency == "" {
		c.JSON(400, gin.H{"error": "currency parameter is required"})
		return
	}

	if currency != "USD" && currency != "EUR" && currency != "RUB" {
		c.JSON(400, gin.H{"error": "currency must be one of: USD, EUR, RUB"})
		return
	}

	balance, err := h.service.GetBalance(c.Request.Context(), userID.(string), currency)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"currency": currency,
		"balance":  balance,
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
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Deposit(c.Request.Context(), userID.(string), req.Currency, req.Amount)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":  "Deposit successful",
		"currency": req.Currency,
		"amount":   req.Amount,
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
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := h.service.Withdraw(c.Request.Context(), userID.(string), req.Currency, req.Amount)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message":  "Withdrawal successful",
		"currency": req.Currency,
		"amount":   req.Amount,
	})
}
