package handler

import "github.com/gin-gonic/gin"

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.JSON(200, gin.H{"message": "login endpoint"})
}
