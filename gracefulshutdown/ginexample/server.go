package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	srv         *http.Server
	quitChannel chan struct{}
	running     sync.WaitGroup
}

func NewServer(addr string) *Server {
	gin.SetMode(gin.ReleaseMode)
	e := gin.Default()
	e.GET("/request", func(ctx *gin.Context) {
		log.Println("Request /request")
		s := ctx.Query("sleep")
		if s == "" {
			s = "5s"
		}

		d, err := time.ParseDuration(s)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		time.Sleep(d)
		log.Println("Complete /request", "ctx.error", ctx.Request.Context().Err())
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Sleep %s", s),
		})
	})

	srv := http.Server{
		Addr:    addr,
		Handler: e,
	}

	s := Server{
		srv:         &srv,
		quitChannel: make(chan struct{}, 1),
	}
	go s.loopMain()
	return &s
}

func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	err := s.srv.Shutdown(ctx)
	close(s.quitChannel)
	//s.running.Wait()
	return err
}

func (s *Server) loopMain() {
	done := true
	s.running.Add(1)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer func() {
		log.Println("loopMain task done:", done)
		ticker.Stop()
		s.running.Done()
	}()

	for {
		select {
		case <-ticker.C:
			done = false
			time.Sleep(time.Second)
			done = true
		case <-s.quitChannel:
			log.Println("Terminate work..!")
			return
		}
	}
}
