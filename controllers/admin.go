package controllers

import (
	"net/http"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// GetDashboardStats returns statistics for the admin dashboard
func GetDashboardStats(c *gin.Context) {
	stats, err := models.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to fetch stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: stats})
}
