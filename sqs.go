package awshoney

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/honeycombio/beeline-go/propagation"
	"github.com/honeycombio/beeline-go/trace"
)

func contexterInsertSqsAttributes(ctx context.Context, input interface{}) {
	switch input := input.(type) {
	case *sqs.SendMessageInput:
		attrs := &input.MessageAttributes
		InsertTraceSqsAttribute(ctx, attrs)
	case *sqs.SendMessageBatchInput:
		for _, entry := range input.Entries {
			InsertTraceSqsAttribute(ctx, &entry.MessageAttributes)
		}
	}
}

func InsertTraceSqsAttribute(ctx context.Context, attrs *map[string]*sqs.MessageAttributeValue) {
	if attrs == nil {
		return
	}

	if *attrs == nil {
		*attrs = map[string]*sqs.MessageAttributeValue{}
	}

	span := trace.GetSpanFromContext(ctx)
	if span != nil {
		(*attrs)[propagation.TracePropagationHTTPHeader] = &sqs.MessageAttributeValue{
			StringValue: aws.String(span.SerializeHeaders()),
			DataType:    aws.String("String"),
		}
	}
}

func StartSpanFromSqs(ctx context.Context, msg *sqs.Message) (context.Context, *trace.Span) {
	hdrval := ""

	if hdr := msg.MessageAttributes[propagation.TracePropagationHTTPHeader]; hdr != nil {
		hdrval = *hdr.StringValue
	}

	ctx, t := trace.NewTrace(ctx, hdrval)
	span := t.GetRootSpan()
	span.AddTraceField("sqs.messageid", *msg.MessageId)
	return ctx, span
}
