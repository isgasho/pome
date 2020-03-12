package trace

import (
	"context"
	"encoding/base64"
	"github.com/uber/jaeger-client-go/config"
	"strings"

	"fmt"

	"github.com/fumeboy/pome/util/logs"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"google.golang.org/grpc/metadata"
)

const (
	TraceID      = "trace_id"
	binHdrSuffix = "-bin"
)

// metadataTextMap extends a metadata.MD to be an opentracing textmap
type MetadataTextMap metadata.MD

// Set is a opentracing.TextMapReader interface that extracts values.
func (m MetadataTextMap) Set(key, val string) {
	// gRPC allows for complex binary values to be written.
	encodedKey, encodedVal := encodeKeyValue(key, val)
	// The metadata object is a multimap, and previous values may exist, but for opentracing headers, we do not append
	// we just override.
	m[encodedKey] = []string{encodedVal}
}

// ForeachKey is a opentracing.TextMapReader interface that extracts values.
func (m MetadataTextMap) ForeachKey(callback func(key, val string) error) error {
	for k, vv := range m {
		for _, v := range vv {
			if decodedKey, decodedVal, err := metadata.DecodeKeyValue(k, v); err == nil {
				if err = callback(decodedKey, decodedVal); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("failed decoding opentracing from gRPC metadata: %v", err)
			}
		}
	}
	return nil
}

// encodeKeyValue encodes key and value qualified for transmission via gRPC.
// note: copy pasted from private values of grpc.metadata
func encodeKeyValue(k, v string) (string, string) {
	k = strings.ToLower(k)
	if strings.HasSuffix(k, binHdrSuffix) {
		val := base64.StdEncoding.EncodeToString([]byte(v))
		v = string(val)
	}
	return k, v
}

func InitTrace(serviceName, reportAddr, sampleType string, rate float64) (err error) {
	cfg := &config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  sampleType, //const = 固定采样
			Param: rate,       // 1=全采样、0=不采样
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: reportAddr,
		},
	}
	tracer, closer, err := cfg.NewTracer(
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		logs.Error(context.TODO(), "ERROR: cannot init Jaeger: %v\n", err)
		return
	}

	_ = closer
	opentracing.SetGlobalTracer(tracer)
	return
}
