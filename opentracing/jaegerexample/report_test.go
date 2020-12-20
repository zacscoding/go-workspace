package jaegerexample

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-client-go/zipkin"
	"strconv"
	"testing"
	"time"
)

func TestSpanWithReport(t *testing.T) {
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
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	spanA := opentracing.StartSpan("SpanA")
	PrintSpan(spanA, "SpanA")

	spanB := opentracing.GlobalTracer().StartSpan("SpanB", opentracing.ChildOf(spanA.Context()))
	PrintSpan(spanB, "SpanB")
	spanB.Finish()

	spanC := opentracing.GlobalTracer().StartSpan("SpanC", opentracing.FollowsFrom(spanB.Context()))
	PrintSpan(spanC, "SpanC")
	spanC.Finish()
	spanA.Finish()
	time.Sleep(1 * time.Minute)
}

func PrintSpan(span opentracing.Span, operationName string) {
	spanCtx := span.Context().(jaeger.SpanContext)
	fmt.Println("--------------------------------")
	fmt.Println("Span::", operationName)
	fmt.Printf("TraceID:%d(%s)\n", MustParseUint(spanCtx.TraceID().String()), spanCtx.TraceID().String())
	fmt.Printf("SpanID:%d(%s)\n", MustParseUint(spanCtx.SpanID().String()), spanCtx.SpanID().String())
	fmt.Printf("ParentSpanID:%d(%s)\n", MustParseUint(spanCtx.ParentID().String()), spanCtx.ParentID().String())
}

func MustParseUint(hex string) uint64 {
	v, err := strconv.ParseUint(hex, 16, 64)
	if err != nil {
		panic(err)
	}
	return v
}
