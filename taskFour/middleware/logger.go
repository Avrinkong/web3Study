package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path

		log.Printf("| %3d | %13v | %15s | %-7s %s",
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
	}
}
