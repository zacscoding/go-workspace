package timeout

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

func StartServer(middleware ...gin.HandlerFunc) {
	gin.SetMode(gin.DebugMode)
	e := gin.New()
	e.Use(gin.Recovery(), loggingMiddleware(), firstMiddleware())
	e.Use(middleware...)

	e.GET("/sleep", handleSleep)

	server := &http.Server{
		Addr:         ":3000",
		Handler:      e,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func firstMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		now := time.Now()
		checkDoneFn := func(ctx context.Context, name string) {
			select {
			case <-ctx.Done():
				log.Printf("[Server:firstMiddleware] %s is done. elapsed: %v", name, time.Now().Sub(now))
			}
		}
		go checkDoneFn(c, "gin.Context")
		go checkDoneFn(c.Request.Context(), "gin.Request.Context")
		c.Writer.Header().Add("first", "second")
	}
}

// https://gist.github.com/montanaflynn/ef9e7b9cd21b355cfe8332b4f20163c1
// timeout middleware wraps the request context with a timeout
func timeoutMiddleware(timeout time.Duration) func(c *gin.Context) {
	return func(c *gin.Context) {
		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)

		defer func() {
			// check if context timeout was reached
			if ctx.Err() == context.DeadlineExceeded {
				log.Printf("[Server:timeoutMiddleware] context deadline exceeded")
				// write response and abort the request
				c.Writer.WriteHeader(http.StatusGatewayTimeout)
				b, _ := json.Marshal(map[string]interface{}{
					"err": "timeout!",
				})
				c.Writer.Write(b)
				c.Abort()
			}
			//cancel to clear resources after finished
			cancel()
		}()

		// replace request with context wrapped request
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func loggingMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		closed := false
		go func() {
			select {
			case <-c.Request.Context().Done():
				closed = true
			}
		}()

		// process request
		c.Next()

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		latencyValue := latency.String()
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}
		log.Printf("[GIN] %v | %3d [closed: %v] | %13v | %15s | %-7s %#v",
			timestamp.Format("2006/01/02 - 15:04:05"),
			statusCode,
			closed,
			latencyValue,
			clientIP,
			method,
			path,
		)
	}
}

type ResponseData struct {
	StatusCode int
	Body       interface{}
}

func handleRequest(c *gin.Context, f func(c *gin.Context) *ResponseData) {
	ctx := c.Request.Context()
	if _, ok := ctx.Deadline(); !ok {
		res := f(c)
		c.JSON(res.StatusCode, res.Body)
		return
	}

	doneChan := make(chan *ResponseData)
	go func() {
		doneChan <- f(c)
	}()

	select {
	case <-ctx.Done():
		log.Println("[Server:handleRequest] context in request is done")
	case res := <-doneChan:
		c.JSON(res.StatusCode, res.Body)
	}
}

func handleSleep(ctx *gin.Context) {
	handleRequest(ctx, func(c *gin.Context) *ResponseData {
		type QueryParameter struct {
			Seconds int `form:"sleep"`
		}
		var query QueryParameter
		if err := ctx.ShouldBindQuery(&query); err != nil {
			return &ResponseData{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Errorf("invalid query parameter: %w", err),
			}
		}

		log.Printf("## [Server:handleSleep] sleep %d secs", query.Seconds)
		time.Sleep(time.Duration(query.Seconds) * time.Second)
		return &ResponseData{
			StatusCode: http.StatusOK,
			Body: gin.H{
				"sleep": query.Seconds,
			},
		}
	})
}
