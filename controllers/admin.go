package controllers

import (
	"net/http"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// GetDashboardStats godoc
// @Summary      Get dashboard statistics
// @Description  Retrieve sales and ticket statistics for the admin dashboard
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200   {object}  lib.Response{result=lib.DashboardStats}
// @Failure      500   {object}  lib.Response
// @Router       /admin/stats [get]
func GetDashboardStats(c *gin.Context) {
	stats, err := models.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Failed to fetch stats: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: stats})
}
