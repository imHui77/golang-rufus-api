package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/api/rufus", handleRufusRequest)
}

func handleRufusRequest(c *gin.Context) {
	// 實現Rufus API的邏輯
	c.JSON(200, gin.H{
		"message": "Hello from Rufus API",
	})
}