package handler

import (
	"gw-currency-wallet/internal/middleware"
	"gw-currency-wallet/internal/utils/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	walletHandler   *WalletHandler
	authHandler     *AuthHandler
	exchangeHandler *ExchangeHandler
	jwtManager      *auth.Manager
}

func NewHandler(walletHandler *WalletHandler, authHandler *AuthHandler, exchangeHandler *ExchangeHandler, jwtManager *auth.Manager) *Handler {
	return &Handler{
		walletHandler:   walletHandler,
		authHandler:     authHandler,
		exchangeHandler: exchangeHandler,
		jwtManager:      jwtManager,
	}
}

func RegisterRoutes(router *gin.Engine, handler *Handler) {
	api := router.Group("/api/v1")
	{
		api.POST("/register", handler.authHandler.Register)
		api.POST("/login", handler.authHandler.Login)
		api.POST("/refresh", handler.authHandler.Refresh)
	}

	protected := api.Group("/")
	protected.Use(middleware.JWTMiddleware(handler.jwtManager))
	{
		protected.GET("/balance", handler.walletHandler.GetBalance)
		protected.POST("/wallet/deposit", handler.walletHandler.Deposit)
		protected.POST("/wallet/withdraw", handler.walletHandler.Withdraw)
		protected.GET("/exchange/rates", handler.exchangeHandler.GetExchangeRates)
		protected.POST("/exchange", handler.exchangeHandler.Exchange)
	}
}
