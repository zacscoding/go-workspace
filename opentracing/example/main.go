package main

import (
	"context"
	"errors"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	"io"
	"sync"
	"time"
)

func main() {
	tracer, closer := initJaeger()
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)
	startSpan1(context.Background())
}

func startSpan1(ctx context.Context) {
	span2, span2Ctx := opentracing.StartSpanFromContext(ctx, "OperateSpan2")
	defer span2.Finish()
	startSpan2(span2Ctx)
}

func startSpan2(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		span3, span3Ctx := opentracing.StartSpanFromContext(ctx, "OperateSpan3")
		span3.SetTag("event", "span3") // Tag

		defer func() {
			wg.Done()
			span3.Finish()
		}()
		if err := startSpan3(span3Ctx); err != nil {
			span3.LogKV("err", err)
		}
	}()
	go func() {
		span4, span3Ctx := opentracing.StartSpanFromContext(ctx, "OperateSpan4")
		span4.SetTag("event", "span4") // Tag

		defer func() {
			wg.Done()
			span4.Finish()
		}()
		if err := startSpan4(span3Ctx); err != nil {
			ext.LogError(span4, err)
		}
	}()
	wg.Wait()
}

func startSpan3(ctx context.Context) error {
	return errors.New("force error")
}

func startSpan4(ctx context.Context) error {
	return nil
}

func initJaeger() (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		ServiceName: "MyApplication",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			CollectorEndpoint:   "http://localhost:14268/api/traces",
			LogSpans:            true,
			BufferFlushInterval: time.Second,
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}
