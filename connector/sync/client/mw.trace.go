package client

import (
	"context"
	"fmt"
	"github.com/fumeboy/llog"
	"github.com/fumeboy/pome/manager/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/metadata"
)

func mw_trace(next mw_fn) mw_fn {
	return func(ctx context.Context, ctx2 *ctxT) (err error) {
		fmt.Println("client's traceMiddleware")
		var parentSpanCtx opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentSpanCtx = parent.Context()
		}
		tracer := opentracing.GlobalTracer()
		opts := []opentracing.StartSpanOption{
			opentracing.ChildOf(parentSpanCtx),
			ext.SpanKindRPCClient,
		}
		span := tracer.StartSpan(ctx2.method, opts...)
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.Pairs()
		} else {
			//如果对metadata进行修改，那么需要用拷贝的副本进行修改。（FromIncomingContext的注释）
			//md = md.Copy()
		}
		if err := tracer.Inject(span.Context(), opentracing.TextMap, trace.MDReaderWriter{md}); err != nil {
			llog.Debug("grpc_opentracing: failed serializing trace information: %v", err)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)
		err = next(ctx,ctx2)
		//记录错误
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
		}

		span.Finish()
		return
	}
}