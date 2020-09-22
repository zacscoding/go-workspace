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
	e := gin.New()
	e.Use(gin.Recovery(), firstMiddleware())
	e.Use(middleware...)

	e.GET("/sleep1", handleSleep)

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
		log.Println("timeout occur in handleRequest()")
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

		log.Printf("## [Server] sleep %d secs", query.Seconds)
		time.Sleep(time.Duration(query.Seconds) * time.Second)
		return &ResponseData{
			StatusCode: http.StatusOK,
			Body: gin.H{
				"sleep": query.Seconds,
			},
		}
	})
}
