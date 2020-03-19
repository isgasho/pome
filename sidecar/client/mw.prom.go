package client

import (
	"context"
	"fmt"
	"github.com/fumeboy/pome/sidecar/middleware/prometheus"
	"time"
)

var (
	defaultMetrics = prometheus.NewClientMetrics()
)

func mw_prom(next mw_fn) mw_fn {
	return func(ctx context.Context, ctx2 *ctxT) (err error) {
		fmt.Println("client's promMiddleware")
		defaultMetrics.IncrRequest(ctx, ctx2.ServiceName, ctx2.Method)

		startTime := time.Now()
		err = next(ctx, ctx2)

		defaultMetrics.IncrCode(ctx, ctx2.ServiceName, ctx2.Method, err)
		defaultMetrics.Latency(ctx, ctx2.ServiceName,
			ctx2.Method, time.Since(startTime).Nanoseconds()/1000)
		return
	}
}

