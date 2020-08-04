package circuitbreaker

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"log"
	"net/http"
	"time"
)

type Client interface {
	Execute(req func() (interface{}, error)) (interface{}, error)
}

type noopClient struct {
}

func (n *noopClient) Execute(req func() (interface{}, error)) (interface{}, error) {
	return req()
}

func NewClient() Client {
	return &noopClient{}
}

func NewCircuitBreakerClient() Client {
	st := gobreaker.Settings{
		Name:        "service1",
		MaxRequests: 0,
		Interval:    0,
		Timeout:     2 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			b, _ := json.Marshal(counts)
			log.Printf("ReadyToTrip: %s", string(b))
			rate := float32(counts.ConsecutiveFailures) / float32(counts.Requests)
			if rate >= 0.5 {
				return true
			}
			return false
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit[%s] %v -> %v", name, from, to)
		},
	}
	return gobreaker.NewCircuitBreaker(st)
}

func NewStubServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	e := gin.Default()
	e.GET("/sleep/:id/:seconds/:error", func(c *gin.Context) {
		type Uri struct {
			RequestId string `uri:"id"`
			Seconds   int    `uri:"seconds"`
			Error     bool   `uri:"error"`
		}
		var uri Uri
		err := c.ShouldBindUri(&uri)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if uri.Seconds != 0 {
			time.Sleep(time.Second * time.Duration(uri.Seconds))
		}

		if uri.Error {
			c.JSON(200, gin.H{"message": "ok"})
		} else {
			c.JSON(500, gin.H{"message": "internal error"})
		}
	})
	return e
}
