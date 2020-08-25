package main

import (
	"context"
	"github.com/opentracing/opentracing-go"
)

func main() {
	ctx := context.Background()
	span, ctx := opentracing.StartSpanFromContext(ctx, "operation_name")

}
