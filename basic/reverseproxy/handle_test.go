package reverseproxy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-workspace/logger/zaplogger"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"testing"
	"time"
)

// tests httputil.ReverseProxy's ModifyResponse and ErrorHandler
func Test1(t *testing.T) {
	go startRemoteServer(":8000")
	go startReverseProxyServer("http://localhost:8000", ":8800")

	cli := http.Client{
		//Timeout: time.Millisecond * 2000,
	}
	u := fmt.Sprintf("http://localhost:8800/call?sleep=%d", 5000)

	start := time.Now()
	resp, err := cli.Get(u)
	elapsed := time.Now().Sub(start).Milliseconds()
	if err != nil {
		log.Printf("[Client] Error: %v [%d ms]", err, elapsed)
	} else {
		defer resp.Body.Close()
		log.Printf("[Client] StatusCode: %d [%d ms]", resp.StatusCode, elapsed)
	}
	time.Sleep(30 * time.Second)
}

func startReverseProxyServer(targetRawUrl, addr string) {
	e := gin.New()
	e.Use(loggingMiddleware("ReverseProxy"), gin.Recovery())
	targetUrl, _ := url.Parse(targetRawUrl)
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxConnsPerHost:       2,
		MaxIdleConns:          2,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		if tctx, ok := FromTraceContext(resp.Request.Context()); ok {
			elapsed := time.Now().Sub(tctx.RequestTime).Milliseconds()
			zaplogger.DefaultLogger().Infof("[ReverseProxy::ModifyResponse] exist key in ctx:%s [%d ms]", tctx.RequestId, elapsed)
		} else {
			zaplogger.DefaultLogger().Info("[ReverseProxy::ModifyResponse] not exist trace context in ctx")
		}
		return nil
	}
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		if tctx, ok := FromTraceContext(request.Context()); ok {
			elapsed := time.Now().Sub(tctx.RequestTime).Milliseconds()
			zaplogger.DefaultLogger().Infof("[ReverseProxy::ErrorHandler] exist key in ctx:%s, err:%v [%d ms]", tctx.RequestId, err, elapsed)
		} else {
			zaplogger.DefaultLogger().Infof("[ReverseProxy::ErrorHandler] not exist key in ctx. err:%v", err)
		}
		writer.WriteHeader(http.StatusBadGateway)
	}

	e.GET("call", func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	})

	if err := e.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func startRemoteServer(addr string) {
	e := gin.New()
	e.Use(loggingMiddleware("RemoteServer"), gin.Recovery())
	e.GET("call", func(ctx *gin.Context) {
		sleep := ctx.Query("sleep")
		if sleep != "" {
			sleepMills, err := strconv.Atoi(sleep)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"message": fmt.Sprintf("invalid sleep query:%v", err),
				})
				return
			}
			zaplogger.DefaultLogger().Infof("[Remote Server] sleep %d [ms]", sleepMills)
			time.Sleep(time.Duration(sleepMills) * time.Millisecond)
		}

		statusCodeVal := ctx.Query("statusCode")
		if statusCodeVal == "" {
			statusCodeVal = "200"
		}
		statusCode, err := strconv.Atoi(statusCodeVal)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("invalid statusCode query:%v", err),
			})
			return
		}
		ctx.JSON(statusCode, gin.H{
			"status": "ok",
		})
	})

	if err := e.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func loggingMiddleware(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceCtx := TraceContext{
			RequestTime: time.Now(),
			RequestId:   uuid.New().String(),
		}
		c.Request = c.Request.WithContext(WithTraceContext(c.Request.Context(), &traceCtx))

		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		cancelled := false
		doneCh := make(chan struct{}, 1)
		go func() {
			select {
			case <-c.Request.Context().Done():
				cancelled = true
			case <-doneCh:
				return
			}
		}()
		defer close(doneCh)

		// process request
		c.Next()

		timestamp := time.Now()
		latency := timestamp.Sub(traceCtx.RequestTime)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}
		template := "[GIN-%s] %v | %3d | %13v | %15s | %-7s %#v | cancelled:%v | key: %s"
		zaplogger.DefaultLogger().Infof(template,
			name,
			timestamp.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency.String(),
			clientIP,
			method,
			path,
			cancelled,
			traceCtx.RequestId,
		)
	}
}
