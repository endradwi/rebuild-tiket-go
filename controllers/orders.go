package controllers

import (
	"net/http"
	"strconv"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// GetSeats retrieves all seats and marks occupied ones for a showtime
// GetSeats godoc
// @Summary      Get seats for showtime
// @Description  Retrieve all seats and their occupancy status for a specific showtime
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        showtime_id  query     int     true  "Showtime ID"
// @Success      200   {object}  lib.Response{result=[]lib.SeatWithStatus}
// @Failure      500   {object}  lib.Response
// @Router       /orders/seats [get]
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

	var results []lib.SeatWithStatus
	for _, s := range seats {
		results = append(results, lib.SeatWithStatus{
			Seat:       s,
			IsOccupied: occupiedMap[s.Id],
		})
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "success", Result: results})
}

// CreateOrder godoc
// @Summary      Create an order
// @Description  Step 2: Initialize an order with selected seats
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        order  body      lib.OrderCreateRequest  true  "Order Details"
// @Success      201   {object}  lib.Response{result=lib.Order}
// @Failure      400   {object}  lib.Response
// @Failure      500   {object}  lib.Response
// @Router       /orders [post]
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
// GetOrderDetails godoc
// @Summary      Get order details
// @Description  Retrieve order summary for the Payment Page
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id    path      int     true  "Order ID"
// @Success      200   {object}  lib.Response{result=lib.Order}
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Router       /orders/{id} [get]
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

// ProcessPayment godoc
// @Summary      Process payment
// @Description  Step 3: Submit personal info and payment method to complete the order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id       path      int                 true  "Order ID"
// @Param        payment  body      lib.PaymentRequest  true  "Payment Details"
// @Success      200      {object}  lib.Response{result=lib.Payment}
// @Failure      400      {object}  lib.Response
// @Failure      500      {object}  lib.Response
// @Router       /orders/{id}/payment [post]
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

// GetTicketResult godoc
// @Summary      Get ticket result
// @Description  Retrieve the final ticket summary for a specific order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id    path      int     true  "Order ID"
// @Success      200   {object}  lib.Response{result=lib.Order}
// @Failure      400   {object}  lib.Response
// @Failure      404   {object}  lib.Response
// @Router       /orders/{id}/ticket [get]
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
