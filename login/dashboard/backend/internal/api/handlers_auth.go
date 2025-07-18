package api

import (
	"log"
	"net/http"

	"slurm-dashboard/config"
	"slurm-dashboard/internal/auth"
	"slurm-dashboard/internal/services"
	"slurm-dashboard/internal/store"

	"github.com/gin-gonic/gin"
)

type LoginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginHandler 负责处理登录逻辑
func LoginHandler(cfg *config.Config, tokenStore *store.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload LoginPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		isAuthenticated, err := auth.AuthenticateLDAP(cfg, payload.Username, payload.Password)
		if err != nil || !isAuthenticated {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		slurmToken, err := services.GetSlurmToken(payload.Username, cfg.SlurmTokenLifespanSec)
		if err != nil {
			log.Printf("Slurm token generation error for user %s: %v", payload.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Slurm token"})
			return
		}

		tokenStore.Set(payload.Username, slurmToken)
		log.Printf("Stored Slurm token for user: %s", payload.Username)

		customToken, err := auth.GenerateCustomToken(cfg, payload.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate custom token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": customToken})
	}
}
