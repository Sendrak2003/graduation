package http

import (
	"gw-currency-wallet/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	walletHandler *WalletHandler
}

func NewHandler(walletHandler *WalletHandler) *Handler {
	return &Handler{
		walletHandler: walletHandler,
	}
}

func RegisterRoutes(router *gin.Engine, handler *Handler) {

	api := router.Group("/api/v1")
	{
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)
		api.POST("/refresh", handler.Refresh)
	}

	protected := api.Group("/")
	protected.Use(middleware.JWTMiddleware())
	{
		protected.GET("/balance", handler.GetBalance)
		protected.POST("/wallet/deposit", handler.Deposit)
		protected.POST("/wallet/withdraw", handler.Withdraw)
		protected.GET("/exchange/rates", handler.GetRates)
		protected.POST("/exchange", handler.Exchange)
	}
}

// Заглушки для методов Handler (будут реализованы позже)

func (h *Handler) Register(c *gin.Context) {
	c.JSON(200, gin.H{"message": "register endpoint"})
}

func (h *Handler) Login(c *gin.Context) {
	c.JSON(200, gin.H{"message": "login endpoint"})
}

func (h *Handler) Refresh(c *gin.Context) {
	c.JSON(200, gin.H{"message": "refresh endpoint"})
}

func (h *Handler) GetBalance(c *gin.Context) {
	c.JSON(200, gin.H{"message": "balance endpoint"})
}

func (h *Handler) Deposit(c *gin.Context) {
	c.JSON(200, gin.H{"message": "deposit endpoint"})
}

func (h *Handler) Withdraw(c *gin.Context) {
	c.JSON(200, gin.H{"message": "withdraw endpoint"})
}

func (h *Handler) GetRates(c *gin.Context) {
	c.JSON(200, gin.H{"message": "rates endpoint"})
}

func (h *Handler) Exchange(c *gin.Context) {
	c.JSON(200, gin.H{"message": "exchange endpoint"})
}
