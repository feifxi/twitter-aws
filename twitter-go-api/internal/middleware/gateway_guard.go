package middleware

import "github.com/gin-gonic/gin"

func GatewayGuard() gin.HandlerFunc {
    return func(c *gin.Context) {
        secret := c.GetHeader("X-Gateway-Secret")
        if secret != "chanom-very-cool" {
            c.AbortWithStatus(403)
            return
        }
        c.Next()
    }
}