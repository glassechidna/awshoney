package awshoney

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/honeycombio/beeline-go"
	"strings"
)

type wrapper struct {
	inner    lambda.Handler
	spanName string
	warm     bool
	counter  int
}

func (w *wrapper) Invoke(ctx context.Context, payload []byte) (output []byte, err error) {
	ctx, span := beeline.StartSpan(ctx, w.spanName)
	defer beeline.Flush(ctx)
	defer span.Send()

	span.AddField("aws.lambda.invocation-counter", w.counter)
	w.counter++

	if !w.warm {
		span.AddField("aws.lambda.cold-start", "true")
		w.warm = true
	} else {
		span.AddField("aws.lambda.cold-start", "false")
	}

	if lctx, ok := lambdacontext.FromContext(ctx); ok {
		span.AddField("aws.lambda.request-id", lctx.AwsRequestID)

		version := "$LATEST"
		split := strings.Split(lctx.InvokedFunctionArn, ":")
		if len(split) == 8 {
			version = split[7]
		}

		span.AddField("aws.lambda.invoked-version", version)
	}

	return w.inner.Invoke(ctx, payload)
}

// WrapLambda takes a root span name and an "inner" lambda handler to wrap.
// This inner handler can either be a plain function (i.e. one compatible
// with lambda.Start()) or a type implementing the lambda.Handler interface.
//
// In addition to starting a trace, additional fields are added to the root
// span.
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
