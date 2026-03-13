package controllers

import (
	"net/http"
	"strconv"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// GetSeats retrieves all seats and marks occupied ones for a showtime
func GetSeats(c *gin.Context) {
	showtimeIdStr := c.Query("showtime_id")
	showtimeId, _ := strconv.Atoi(showtimeIdStr)

	seats, err := models.GetAllSeats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: err.Error()})
		return
	}

	occupiedIds, _ := models.GetOccupiedSeats(showtimeId)
	occupiedMap := make(map[int]bool)
	for _, id := range occupiedIds {
		occupiedMap[id] = true
	}

	type SeatWithStatus struct {
		lib.Seat
		IsOccupied bool `json:"is_occupied"`
	}

	var results []SeatWithStatus
	for _, s := range seats {
		results = append(results, SeatWithStatus{
			Seat:       s,
			IsOccupied: occupiedMap[s.Id],
		})
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: results})
}

// CreateOrder handles Step 2 of the flow (Seat Selection -> Checkout)
func CreateOrder(c *gin.Context) {
	userIdAny, exists := c.Get("userId")
	var userId int
	if exists {
		userId = int(userIdAny.(float64))
	}

	var req lib.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: err.Error()})
		return
	}

	order, err := models.CreateOrder(userId, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lib.Response{Status: 201, Message: "Order created successfully", Result: order})
}

// GetOrderDetails handles fetching order summary for the Payment Page
func GetOrderDetails(c *gin.Context) {
	orderIdStr := c.Param("id")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid order ID"})
		return
	}

	order, err := models.GetOrderById(orderId)
	if err != nil {
		c.JSON(http.StatusNotFound, lib.Response{Status: 404, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: order})
}

// ProcessPayment handles Step 3 of the flow (Personal Info + Payment Method -> Pay now)
func ProcessPayment(c *gin.Context) {
	orderIdStr := c.Param("id")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid order ID"})
		return
	}

	var req lib.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: err.Error()})
		return
	}

	payment, err := models.CreatePayment(orderId, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "Payment successful", Result: payment})
}

// GetTicketResult retrieves the ticket summary for a specific order
func GetTicketResult(c *gin.Context) {
	orderIdStr := c.Param("id")
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Invalid order ID"})
		return
	}

	order, err := models.GetOrderById(orderId)
	if err != nil {
		c.JSON(http.StatusNotFound, lib.Response{Status: 404, Message: err.Error()})
		return
	}

	// We can reuse GetOrderById since it already joins movie, cinema, showtimes, and seats.
	// The design calls for a QR code which we stored in the payment table.
	// If payment exists, we should probably include it.
	
	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: order})
}
