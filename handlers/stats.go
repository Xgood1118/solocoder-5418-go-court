package handlers

import (
	"net/http"
	"time"

	"court-scheduler/services"
	"court-scheduler/store"

	"github.com/gin-gonic/gin"
)

func ListStats(c *gin.Context) {
	stats := store.GetAllStats()
	c.JSON(http.StatusOK, stats)
}

func GetMonthlyStats(c *gin.Context) {
	month := c.Param("month")
	if month == "" {
		now := time.Now()
		month = now.Format("2006-01")
	}

	stats, ok := store.GetStats(month)
	if !ok {
		stats = services.GenerateMonthlyStats(month)
	}

	c.JSON(http.StatusOK, stats)
}

func GenerateStats(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		now := time.Now()
		month = now.Format("2006-01")
	}

	stats := services.GenerateMonthlyStats(month)
	c.JSON(http.StatusOK, stats)
}
