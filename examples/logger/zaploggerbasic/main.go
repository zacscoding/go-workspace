package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-workspace/logger/zaplogger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	s := NewServer()
	go func() {
		if err := s.Run(":3000"); err != nil {
			panic(err)
		}
	}()

	cancel := make(chan struct{})
	for i := 0; i < 2; i++ {
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-cancel:
					return
				case <-ticker.C:
					// just call
					http.Get("http://localhost:3000/persons")
				}
			}

		}()
	}

	time.Sleep(1 * time.Minute)
	cancel <- struct{}{}
	cancel <- struct{}{}
}

type Server struct {
	*gin.Engine
	logger *zap.SugaredLogger
}

func NewServer() *Server {
	gin.SetMode(gin.TestMode)
	s := &Server{
		Engine: gin.Default(),
		logger: zaplogger.DefaultLogger(),
	}
	s.GET("/persons", s.getPersons)
	return s
}

func (s *Server) getPersons(ctx *gin.Context) {
	traceId := uuid.New().String()
	c := zaplogger.WithLogger(ctx, s.logger.With("traceId", traceId))
	ctx.JSON(http.StatusOK, getPersons1(c, traceId))
}

func getPersons1(ctx context.Context, traceId string) map[string]interface{} {
	logger := zaplogger.FromContext(ctx)
	logger.Info("Service1:getPerson1()-traceId: " + traceId)

	return getPerson2(ctx, traceId)
}

func getPerson2(ctx context.Context, traceId string) map[string]interface{} {
	logger := zaplogger.FromContext(ctx)
	logger.Info("Service2:getPerson2()-traceId: " + traceId)

	return map[string]interface{}{
		"Name": "zaccoding",
		"Age":  10,
	}
}
