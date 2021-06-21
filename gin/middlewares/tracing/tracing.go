package tracing

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"runtime"
)

func NewMiddleware(componentName string, skipPaths map[string]struct{}) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if _, ok := skipPaths[ctx.FullPath()]; ok {
			ctx.Next()
			return
		}
		var (
			opname  = ctx.FullPath()
			req     = ctx.Request
			span    opentracing.Span
			carrier = opentracing.HTTPHeadersCarrier(req.Header)
			tracer  = opentracing.GlobalTracer()
		)

		wireContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			span = tracer.StartSpan(opname)
		} else {
			span = tracer.StartSpan(opname, ext.RPCServerOption(wireContext))
		}

		ext.HTTPMethod.Set(span, ctx.Request.Method)
		ext.HTTPUrl.Set(span, req.URL.String())
		ext.Component.Set(span, componentName)

		ctx.Request = req.WithContext(opentracing.ContextWithSpan(req.Context(), span))
		defer func() {
			status := ctx.Writer.Status()
			if status > 299 {
				ext.LogError(span, fmt.Errorf("status:%d", status))
			}
			ext.HTTPStatusCode.Set(span, uint16(status))
			span.Finish()
		}()
		ctx.Next()
	}
}

func StartSpan(ctx *gin.Context, opname string) opentracing.Span {
	var span opentracing.Span
	if parentSpan := opentracing.SpanFromContext(ctx.Request.Context()); parentSpan != nil {
		span = opentracing.StartSpan(opname, opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan(opname)
	}

	span.SetTag("name", opname)
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	callerDetails := fmt.Sprintf("%s#%d", frame.Function, frame.Line)
	span.SetTag("caller", callerDetails)

	return span
}
