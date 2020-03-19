package server

import (
	"context"
	"fmt"
	"github.com/fumeboy/pome/sidecar/middleware"
	"github.com/fumeboy/pome/sidecar/middleware/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/metadata"
)

func tracerMiddleware(next middleware.MiddlewareFn) middleware.MiddlewareFn {
	return func(ctx context.Context) (err error) {
		fmt.Println("server's traceMiddleware")
		//从ctx获取grpc的metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			//没有的话,	新建一个
			md = metadata.Pairs()
		} else{
			//md = md.Copy()
		}

		tracer := opentracing.GlobalTracer()
		serverMeta := getMeta(ctx)
		parentSpanContext, err := tracer.Extract(opentracing.TextMap, trace.MDReaderWriter{md})
		if err != nil {
			serverMeta.Log.Warn("trace extract failed, parsing trace information: %v", err)
		}
		//开始追踪该方法
		serverSpan := tracer.StartSpan(
			serverMeta.Method,
			opentracing.ChildOf(parentSpanContext),
			ext.RPCServerOption(parentSpanContext),
			ext.SpanKindRPCServer,
		)
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		err = next(ctx)
		//记录错误
		if err != nil {
			ext.Error.Set(serverSpan, true)
			serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}

		serverSpan.Finish()
		return
	}
}
