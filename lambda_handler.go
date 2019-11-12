package awshoney

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
	"strings"
)

type wrapper struct {
	inner    lambda.Handler
	spanName string
	warm     bool
	counter  int
}

func (w *wrapper) Invoke(ctx context.Context, payload []byte) (output []byte, err error) {
	// a span might have already been created for us, in which case we're just adding
	// useful fields to the trace
	span := trace.GetSpanFromContext(ctx)
	if span == nil {
		ctx, span = beeline.StartSpan(ctx, w.spanName)
		defer beeline.Flush(ctx)
		defer span.Send()
	}

	span.AddTraceField("aws.lambda.invocation-counter", w.counter)
	w.counter++

	if !w.warm {
		span.AddTraceField("aws.lambda.cold-start", "true")
		w.warm = true
	} else {
		span.AddTraceField("aws.lambda.cold-start", "false")
	}

	if lctx, ok := lambdacontext.FromContext(ctx); ok {
		span.AddTraceField("aws.lambda.request-id", lctx.AwsRequestID)

		version := "$LATEST"
		split := strings.Split(lctx.InvokedFunctionArn, ":")
		if len(split) == 8 {
			version = split[7]
		}

		span.AddTraceField("aws.lambda.invoked-version", version)
	}

	return w.inner.Invoke(ctx, payload)
}

// WrapLambda takes a root span name and an "inner" lambda handler to wrap.
// This inner handler can either be a plain function (i.e. one compatible
// with lambda.Start()) or a type implementing the lambda.Handler interface.
//
// If the context does not already have a trace, a new one is started. Either
// way, the current or new span has additional Lambda fields added.
//
// It returns a lambda.Handler that can be passed to lambda.StartHandler().
func WrapLambda(spanName string, inner interface{}) lambda.Handler {
	if spanName == "" {
		spanName = "lambda"
	}

	var innerH lambda.Handler
	if h, ok := inner.(lambda.Handler); ok {
		innerH = h
	} else {
		innerH = lambda.NewHandler(inner)
	}

	return &wrapper{inner: innerH, spanName: spanName}
}
