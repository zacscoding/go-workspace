package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-workspace/logger/zaplogger"
	"go.uber.org/zap"
	"net/http"
)

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
	s.Use(s.NewTraceIdMiddleware())
	s.GET("/persons", s.getPersons)
	return s
}

func (s *Server) NewTraceIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceIdValues, ok := c.Request.Header["trace-id"]
		var traceId string
		if ok {
			traceId = traceIdValues[0]
		} else {
			traceId = uuid.New().String()
		}
		c.Set("trace-id", traceId)
		zaplogger.WithGinLoggwer(c, s.logger.With("traceId", traceId))
	}
}

func (s *Server) getPersons(ctx *gin.Context) {
	logger := zaplogger.FromGinContext(ctx)
	logger.Info("gerPerson() is called..")
	ctx.JSON(http.StatusOK, getPersons1(ctx))
}

func getPersons1(ctx *gin.Context) map[string]interface{} {
	logger := zaplogger.FromGinContext(ctx)
	logger.Info("Service1:getPerson1()")

	return getPerson2(ctx)
}

func getPerson2(ctx *gin.Context) map[string]interface{} {
	logger := zaplogger.FromGinContext(ctx)
	logger.Info("Service2:getPerson2()")

	return map[string]interface{}{
		"Name": "zaccoding",
		"Age":  10,
	}
}
