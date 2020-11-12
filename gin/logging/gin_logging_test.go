package logging

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-workspace/logger/zaplogger"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestLogging(t *testing.T) {
	logger := zaplogger.DefaultLogger()
	r := gin.New()
	r.Use(gin.Recovery(), NewLoggingMiddleware(logger), gin.Logger())

	// Example ping request.
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	r.GET("/resource/:resourceId", func(c *gin.Context) {
		c.String(400, "Invalid resource id")
	})

	// Example when panic happen.
	r.GET("/panic", func(c *gin.Context) {
		panic("An unexpected error happen!")
	})

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

func NewLoggingMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// process request
		c.Next()

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		message := fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %#v\n%s",
			timestamp.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency.String(),
			clientIP,
			method,
			path,
			c.Errors.String(),
		)

		fmt.Printf("## latency:%s, clientIP:%s, method:%s, statusCode:%d, path:%s, raw:%s\n",
			latency.String(), clientIP, method, statusCode, path, raw)

		switch {
		case statusCode >= 400 && statusCode <= 499:
			{
				logger.Warn(message)
			}
		case statusCode >= 500:
			{
				logger.Error(message)
			}
		default:
			logger.Info(message)
		}
	}
}
