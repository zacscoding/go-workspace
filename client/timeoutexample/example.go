package timeoutexample

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

/**
Client -> Server -> ReverseProxy -> RemoteServer
Client <----------- ReverseProxy <- RemoteServer
*/
func startProxyServer(addr, targetRawUrl string, serverReadTimeout, serverWriteTimeout, proxyDialTimeout time.Duration) {
	targetUrl, _ := url.Parse(targetRawUrl)
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   proxyDialTimeout,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   1000,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	http.HandleFunc("/", reverseProxyHandler(proxy))
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func reverseProxyHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("[Proxy] Call %s %s", r.Method, r.RequestURI)
		r.Header.Set("X-Custom", "ReverseProxy")
		p.ServeHTTP(rw, r)
	}
}

func startRemoteServer(addr string) {
	e := gin.New()
	e.POST("/reverse", func(ctx *gin.Context) {
		// logging request info
		log.Printf("[Remote-Server] Call POST %s", ctx.Request.RequestURI)
		log.Println("> Headers")
		for key, values := range ctx.Request.Header {
			log.Printf(">> Key:%s, Values:%s", key, strings.Join(values, ","))
		}
		bytes, _ := ioutil.ReadAll(ctx.Request.Body)
		log.Printf("> Body : %s", string(bytes))

		// sleep if exist query
		sleepQuery := ctx.Query("sleep")
		if sleepQuery != "" {
			sleepSec, err := strconv.Atoi(sleepQuery)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"message": fmt.Sprintf("invalid sleep query: %s", sleepQuery),
				})
				return
			}
			time.Sleep(time.Duration(sleepSec) * time.Second)
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	if err := e.Run(addr); err != nil {
		log.Fatal(err)
	}
}
