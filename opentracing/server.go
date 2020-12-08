package opentracing

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	Port int
}

func NewServer(port int) *Server {
	return &Server{
		Port: port,
	}
}

func (s *Server) ServeHTTP() error {
	e := gin.Default()
	e.Use(NewTracingMiddleware())
	e.GET("/root", func(c *gin.Context) {
		sb := strings.Builder{}
		sb.WriteString("## /root\n")
		for k, v := range c.Request.Header {
			sb.WriteString(fmt.Sprintf("Header key:%s, value:%v\n", k, v))
		}
		fmt.Println(sb.String())
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/child", s.Port), nil)
		if span := opentracing.SpanFromContext(c.Request.Context()); span != nil {
			tracer := opentracing.GlobalTracer()
			err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
			if err != nil {
				fmt.Println("Failed to inject child span:", err)
			}
		}
		http.DefaultClient.Do(req)
		c.String(http.StatusOK, "hello")
	})

	e.GET("/child", func(c *gin.Context) {
		sb := strings.Builder{}
		sb.WriteString("## /child\n")
		for k, v := range c.Request.Header {
			sb.WriteString(fmt.Sprintf("## Header key:%s, value:%v\n", k, v))
		}
		fmt.Println(sb.String())
		c.String(http.StatusOK, "hello")
	})
	return e.Run(fmt.Sprintf(":%d", s.Port))
}

///////////////////////////
// middleware
func NewTracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := ExtractSpan(c)
		fmt.Printf("## %s ExtractSpan exist\n", c.FullPath())
		jSpanCtx, ok := span.Context().(jaeger.SpanContext)
		sb := strings.Builder{}
		if ok {
			sb.WriteString("## " + c.FullPath() + "\n")
			sb.WriteString(fmt.Sprintf("TraceId:%d\n", hexToUint64(jSpanCtx.TraceID().String())))
			sb.WriteString(fmt.Sprintf("SanId:%d\n", hexToUint64(jSpanCtx.SpanID().String())))
			sb.WriteString(fmt.Sprintf("ParentSpanId:%d\n", hexToUint64(jSpanCtx.ParentID().String())))
		} else {
			sb.WriteString(fmt.Sprintf("## %s could not parse jaeger span context", c.FullPath()))
		}
		fmt.Println(sb.String())
		defer span.Finish()

		requestId := c.Request.Header.Get("X-Request-Id")
		if requestId == "" {
			requestId = uuid.New().String()
		}
		span.SetBaggageItem("X-Request-Id", requestId)
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))
		c.Next()
	}
}

///////////////////////////
// trace
func ExtractSpan(c *gin.Context) opentracing.Span {
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	tracer := opentracing.GlobalTracer()
	wireContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		fmt.Printf("## %s ExtractSpan err:%s\n", c.FullPath(), err)
		return opentracing.StartSpan(c.Request.URL.Path)
	}
	return opentracing.StartSpan(c.Request.URL.Path, opentracing.ChildOf(wireContext))
}

func hexToUint64(v string) uint64 {
	u, _ := strconv.ParseUint(v, 16, 64)
	return u
}
