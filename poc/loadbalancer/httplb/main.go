package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*http.Server
}

func NewServer(addr string) *Server {
	e := gin.Default()
	e.GET("/health/liveness", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	e.GET("/call", func(ctx *gin.Context) {
		log.Printf("[%s] GET /call called", addr)
	})
	return &Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: e,
		},
	}
}

func (srv *Server) start() error {
	return srv.ListenAndServe()
}

func (srv *Server) stop() {
	srv.Shutdown(context.Background())
}

type Host struct {
	Address string
	Port    int
}

func (h Host) String() string {
	return h.Address + ":" + strconv.Itoa(h.Port)
}

type BalancedTripper struct {
	counter int64
	hosts   []Host
	rt      http.RoundTripper
}

func (bt *BalancedTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	index := atomic.AddInt64(&bt.counter, 1) % int64(len(bt.hosts))
	req.URL.Host = bt.hosts[index].String()
	log.Printf("Host: %s, URI: %s", req.URL.Host, req.RequestURI)
	return bt.rt.RoundTrip(req)
}

func main() {
	srv1 := NewServer(":8080")
	go srv1.start()
	defer srv1.stop()
	srv2 := NewServer(":8081")
	go srv2.start()
	defer srv2.stop()
	srv3 := NewServer(":8082")
	go srv3.start()
	defer srv3.stop()

	cli := http.Client{
		Transport: &BalancedTripper{
			hosts: []Host{
				{Address: "localhost", Port: 8080},
				{Address: "localhost", Port: 8081},
				{Address: "localhost", Port: 8082},
			},
			rt: http.DefaultTransport,
		},
	}

	for i := 0; i < 5; i++ {
		_, err := cli.Get("http://localhost:8080/call")
		if err != nil {
			log.Printf("failed to call. err: %v", err)
			continue
		}
	}
}
