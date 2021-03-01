package nethttp

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestReadTimeout(t *testing.T) {
	e := gin.Default()
	e.GET("/path", func(ctx *gin.Context) {
		time.Sleep(time.Second)
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})
	appServer := &http.Server{
		Addr:         ":8801",
		Handler:      e,
		ReadTimeout:  time.Nanosecond * 50,
		WriteTimeout: time.Millisecond * 2000,
	}
	http.TimeoutHandler()

	//appServer.ConnState = func(conn net.Conn, state http.ConnState) {
	//	log.Println("------------------------------------------------------------------")
	//	log.Println("[ConnState] conn.LocalAddr().String():", conn.LocalAddr().String())
	//	log.Println("[ConnState] conn.LocalAddr().Network():", conn.LocalAddr().Network())
	//	log.Println("[ConnState] conn.RemoteAddr().String():", conn.RemoteAddr().String())
	//	log.Println("[ConnState] conn.RemoteAddr().Network():", conn.RemoteAddr().Network())
	//	log.Println("[ConnState] state.String():", state.String())
	//}
	go func() {
		if err := appServer.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	start := time.Now()
	resp, err := http.DefaultClient.Get("http://localhost:8801/path")
	elapsed := time.Now().Sub(start)
	if err != nil {
		log.Println("Err:", err, ", Elapsed:", elapsed)
	} else {
		log.Println("Code:", resp.StatusCode, ", Elapsed:", elapsed)
	}
}
