package awshoney

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/glassechidna/awsctx"
	"github.com/honeycombio/beeline-go"
)

func contexter(ctx context.Context, request *awsctx.AwsRequest, inner func(ctx context.Context)) {
	ctx, span := beeline.StartSpan(ctx, fmt.Sprintf("%s %s", request.Service, request.Action))
	defer span.Send()

	span.AddField("aws.service", request.Service)
	span.AddField("aws.action", request.Action)

	if request.Service == "sqs" && (request.Action == "SendMessage" || request.Action == "SendMessageBatch") {
		contexterInsertSqsAttributes(ctx, request.Input)
	}

	inner(ctx)

	if awsErr, ok := request.Error.(awserr.Error); ok {
		span.AddField("aws.error.code", awsErr.Code())
		span.AddField("aws.error.message", awsErr.Message())
	}
}

var Contexter = awsctx.ContexterFunc(contexter)
