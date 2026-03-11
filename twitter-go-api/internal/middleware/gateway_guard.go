package middleware

import "github.com/gin-gonic/gin"

// GatewayGuard rejects requests that don't carry the expected
// X-Gateway-Secret header (injected by API Gateway).
// When secret is empty the guard is a no-op, allowing local development.
func GatewayGuard(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			c.Next()
			return
		}
		if c.GetHeader("X-Gateway-Secret") != secret {
			c.AbortWithStatus(403)
			return
		}
		c.Next()
	}
}
