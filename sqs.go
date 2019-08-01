package awshoney

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/honeycombio/beeline-go/propagation"
	"github.com/honeycombio/beeline-go/trace"
)

type Sqs struct {
	sqsiface.SQSAPI
}

func (s *Sqs) SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, opts ...request.Option) (resp *sqs.SendMessageOutput, err error) {
	InsertTraceSqsAttribute(ctx, &input.MessageAttributes)
	return s.SQSAPI.SendMessageWithContext(ctx, input, opts...)
}

func (s *Sqs) SendMessageBatchWithContext(ctx aws.Context, input *sqs.SendMessageBatchInput, opts ...request.Option) (resp *sqs.SendMessageBatchOutput, err error) {
	for _, entry := range input.Entries {
		InsertTraceSqsAttribute(ctx, &entry.MessageAttributes)
	}

	return s.SQSAPI.SendMessageBatchWithContext(ctx, input, opts...)
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
