package awshoney

import (
	"context"
	"fmt"
	"github.com/glassechidna/awsctx"
	"github.com/honeycombio/beeline-go"
)

func contexter(ctx context.Context, request *awsctx.AwsRequest, inner func(ctx context.Context)) {
	ctx, span := beeline.StartSpan(ctx, fmt.Sprintf("%s %s", request.Service, request.Action))
	span.AddField("aws.service", request.Service)
	span.AddField("aws.action", request.Action)

	if request.Service == "sqs" && (request.Action == "SendMessage" || request.Action == "SendMessageBatch") {
		contexterInsertSqsAttributes(ctx, request.Input)
	}

	defer span.Send()
	inner(ctx)
}

var Contexter = awsctx.ContexterFunc(contexter)
