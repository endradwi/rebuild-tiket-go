package middleware

import (
	"strings"
	"tiket/lib"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" {
			c.JSON(401, lib.Response{
				Status:  401,
				Message: "Unauthorized",
			})
			c.Abort()
			return
		}

		// Standard format for Authorization header: "Bearer <token>"
		parts := strings.Split(tokenHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, lib.Response{
				Status:  401,
				Message: "Unauthorized: Invalid token format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Verify the token
		claims, err := lib.VerifyTokenJwt(tokenString)
		if err != nil {
			c.JSON(401, lib.Response{
				Status:  401,
				Message: "Unauthorized: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Set the userId in the context so it can be accessed in subsequent handlers
		// Note: The userId was set as dbUser.Id (int) during generation, 
		// but jwt-go decodes numeric claims to float64 by default.
		c.Set("userId", claims["userId"])
		c.Set("role", claims["role"])

		c.Next()
	}
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != requiredRole {
			c.JSON(403, gin.H{
				"error": "Forbidden: insufficient permissions",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
