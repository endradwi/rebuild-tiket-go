package controllers

import (
	"net/http"
	"os"
	"tiket/lib"
	"tiket/models"

	"github.com/gin-gonic/gin"
)

// XenditWebhook handles callbacks from Xendit when an invoice status changes
func XenditWebhook(c *gin.Context) {
	// Security: Verify X-Callback-Token if set in .env
	callbackToken := os.Getenv("XENDIT_CALLBACK_TOKEN")
	if callbackToken != "" {
		headerToken := c.GetHeader("x-callback-token")
		if headerToken != callbackToken {
			c.JSON(http.StatusUnauthorized, lib.Response{Status: 401, Message: "Unauthorized: Invalid callback token"})
			return
		}
	}

	var req lib.XenditWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, lib.Response{Status: 400, Message: "Error binding JSON: " + err.Error()})
		return
	}

	// ExternalId in Xendit is our OrderNumber
	err := models.UpdatePaymentStatusByExternalId(req.ExternalId, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, lib.Response{Status: 500, Message: "Error updating status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, lib.Response{Status: 200, Message: "Webhook processed successfully"})
}
