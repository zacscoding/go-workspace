package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-client-go/zipkin"
	"go-workspace/gin/middlewares/tracing"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func StartServer(addr string) error {
	tracerCloser := initGlobalTracer()
	defer tracerCloser.Close()

	e := gin.Default()
	e.Use(tracing.NewMiddleware("MyService", map[string]struct{}{
		"/v1/trace/skip": {},
	}))

	e.GET("/v1/trace", handleTrace)
	e.GET("/v1/trace/skip", handleTrace)

	e.GET("/v1/echo", handleEcho)

	return e.Run(addr)
}

func initGlobalTracer() io.Closer {
	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator(zipkin.BaggagePrefix(""))
	httpTransport := transport.NewHTTPTransport("http://localhost:14268/api/traces", transport.HTTPBatchSize(1))
	tracer, closer := jaeger.NewTracer(
		"sample-service",
		jaeger.NewConstSampler(true),
		// jaeger.NewNullReporter(),
		jaeger.NewRemoteReporter(httpTransport),
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, propagator),
		jaeger.TracerOptions.ZipkinSharedRPCSpan(true),
	)
	opentracing.SetGlobalTracer(tracer)
	return closer
}

func handleTrace(ctx *gin.Context) {
	var (
		status = http.StatusOK
		res    = gin.H{}
	)
	if statusVal := ctx.Query("status"); statusVal != "" {
		s, err := strconv.Atoi(statusVal)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
		}
		status = s
	}

	span := opentracing.SpanFromContext(ctx.Request.Context())
	if span == nil {
		res["span"] = nil
	} else {
		spanCtx := span.Context().(jaeger.SpanContext)
		spanData := gin.H{
			"x-b3-traceid":      spanCtx.TraceID().String(),
			"x-b3-parentspanid": spanCtx.ParentID().String(),
			"x-b3-spanid":       spanCtx.SpanID().String(),
		}
		if spanCtx.IsSampled() {
			spanData["x-b3-sampled"] = "1"
		} else {
			spanData["x-b3-sampled"] = "0"
		}
		res["span"] = spanData
	}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8900/v1/echo", nil)
	childSpan := tracing.StartSpan(ctx, "Call echo")
	opentracing.GlobalTracer().Inject(childSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	http.DefaultClient.Do(req)
	defer childSpan.Finish()

	ctx.JSON(status, res)
}

func handleEcho(ctx *gin.Context) {
	time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, body)
}
