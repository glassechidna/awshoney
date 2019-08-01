package awshoney_test

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/glassechidna/awshoney"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/propagation"
)

func ExampleSqs() {
	// obviously it won't be too helpful if you do this. you'd want to
	// pass in a context from a real span
	ctx, _ := beeline.StartSpan(context.Background(), "example")

	sess := session.Must(session.NewSession())
	baseApi := sqs.New(sess)
	api := &awshoney.Sqs{SQSAPI: baseApi}

	// the other methods that aren't WithContext won't pass through
	// the honeycomb trace id
	_, _ = api.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String("queue-url"),
		MessageBody: aws.String("body"),
	})

	// the other methods that aren't WithContext won't pass through
	// the honeycomb trace id
	_, _ = api.SendMessageBatchWithContext(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String("queue-url"),
		Entries: []*sqs.SendMessageBatchRequestEntry{
			{
				MessageBody: aws.String("body"),
			},
		},
	})

	// you don't need to use the &awshoney.Sqs{} wrapper here (but it won't hurt)
	resp, _ := baseApi.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String("queue-url"),
		// by default sqs won't retrieve message attributes
		MessageAttributeNames: []*string{aws.String(propagation.TracePropagationHTTPHeader)},
	})

	msg := resp.Messages[0]
	ctx, _ = awshoney.StartSpanFromSqs(context.Background(), msg)
}
