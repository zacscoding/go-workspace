package ginexample

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/zipkin"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	require := assert.New(t)
	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator(zipkin.BaggagePrefix(""))
	tracer, closer := jaeger.NewTracer("sample",
		jaeger.NewConstSampler(false),
		jaeger.NewNullReporter(),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
	)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	s := NewServer(3000)
	go func() {
		if err := s.ServeHTTP(); err != nil {
			log.Fatal(err)
		}
	}()
	code, body, err := doRequest(3000, true)
	require.NoError(err)
	fmt.Println("Code: ", code, ", Body:", string(body))
}

func doRequest(port int, includeHeader bool) (int, []byte, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/root", port), nil)
	if includeHeader {
		req.Header.Set("X-Request-Id", uuid.New().String())
		req.Header.Set("X-B3-Spanid", "7cf2b60cc11b4db3")
		req.Header.Set("X-B3-Traceid", "520074fd2207a4e3")
		req.Header.Set("X-B3-Parentspanid", "520074fd2207a4e3")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, b, nil
}

func TestContext(t *testing.T) {
	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator(zipkin.BaggagePrefix(""))
	tracer, closer := jaeger.NewTracer("sample",
		jaeger.NewConstSampler(false),
		jaeger.NewNullReporter(),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
	)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	e := gin.Default()
	e.GET("/path", func(ctx *gin.Context) {
		newctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		deadline, ok := newctx.Deadline()
		log.Printf("check timeout.. deadline:%v, ok:%v", deadline, ok)

		rootspan, rootctx := opentracing.StartSpanFromContext(ctx, "span1")
		displaySpan("opentracing.StartSpanFromContext(ctx, \"span1\")", rootspan)
		displaySpan("opentracing.SpanFromContext(c1Ctx)", opentracing.SpanFromContext(rootctx))
		log.Println("--------------------------------------------------")

		c1span, c1ctx := opentracing.StartSpanFromContext(rootctx, "child1")
		displaySpan("opentracing.StartSpanFromContext(rootctx, \"child1\")", c1span)
		displaySpan(" opentracing.SpanFromContext(c1ctx)", opentracing.SpanFromContext(c1ctx))
		log.Println("--------------------------------------------------")

		c2span, c2ctx := opentracing.StartSpanFromContext(rootctx, "child2")
		displaySpan("opentracing.StartSpanFromContext(rootctx, \"c2span\")", c2span)
		displaySpan(" opentracing.SpanFromContext(c2ctx)", opentracing.SpanFromContext(c2ctx))
		log.Println("--------------------------------------------------")

	})
	go func() {
		if err := e.Run(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	_, err := http.Get("http://localhost:8080/path")
	if err != nil {
		log.Println("failed to do request", err)
		return
	}
}

func displaySpan(prefix string, span opentracing.Span) {
	if span == nil {
		log.Printf("%s > span is nil", prefix)
		return
	}
	if jctx, ok := span.Context().(jaeger.SpanContext); ok {
		log.Printf("%s > traceID: %s, spanID:%s, parentSpanID:%s",
			prefix, jctx.TraceID().String(), jctx.SpanID().String(), jctx.ParentID().String())
		return
	}
	log.Printf("%s > unknown span..:%+v", prefix, span)
}
