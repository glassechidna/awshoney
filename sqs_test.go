package awshoney

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/glassechidna/awsctx/service/sqsctx"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/propagation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSqs_SendMessageWithContext(t *testing.T) {
	ctx, _ := beeline.StartSpan(context.Background(), "test")

	m := &mockSqs{}
	m.On("SendMessageWithContext", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		input := args.Get(1).(*sqs.SendMessageInput)
		attrs := input.MessageAttributes[propagation.TracePropagationHTTPHeader]
		assert.NotEmpty(t, *attrs.StringValue)
	}).Return(nil, nil)

	s := sqsctx.New(m, Contexter())
	_, err := s.SendMessageWithContext(ctx, &sqs.SendMessageInput{})
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestSqs_SendMessageBatchWithContext(t *testing.T) {
	ctx, _ := beeline.StartSpan(context.Background(), "test")

	m := &mockSqs{}
	m.On("SendMessageBatchWithContext", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		input := args.Get(1).(*sqs.SendMessageBatchInput)
		attrs := input.Entries[0].MessageAttributes[propagation.TracePropagationHTTPHeader]
		assert.NotEmpty(t, *attrs.StringValue)
	}).Return(nil, nil)

	s := sqsctx.New(m, Contexter())
	_, err := s.SendMessageBatchWithContext(ctx, &sqs.SendMessageBatchInput{
		Entries: []*sqs.SendMessageBatchRequestEntry{
			{},
		},
	})
	assert.NoError(t, err)
	m.AssertExpectations(t)
}

func TestInsertTraceSqsAttribute(t *testing.T) {
	t.Run("no nil deref", func(t *testing.T) {
		assert.NotPanics(t, func() {
			InsertTraceSqsAttribute(context.Background(), nil)
		})
	})

	t.Run("nil attributes", func(t *testing.T) {
		msg := &sqs.SendMessageInput{}
		ctx, span := beeline.StartSpan(context.Background(), "test")
		InsertTraceSqsAttribute(ctx, &msg.MessageAttributes)
		assert.Equal(t, &sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"X-Honeycomb-Trace": {
					DataType:    aws.String("String"),
					StringValue: aws.String(span.SerializeHeaders()),
				},
			},
		}, msg)
	})

	t.Run("some attributes", func(t *testing.T) {
		msg := &sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"extant": {
					DataType:    aws.String("String"),
					StringValue: aws.String("hello world"),
				},
			},
		}

		ctx, span := beeline.StartSpan(context.Background(), "test")
		InsertTraceSqsAttribute(ctx, &msg.MessageAttributes)
		assert.Equal(t, &sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"X-Honeycomb-Trace": {
					DataType:    aws.String("String"),
					StringValue: aws.String(span.SerializeHeaders()),
				},
				"extant": {
					DataType:    aws.String("String"),
					StringValue: aws.String("hello world"),
				},
			},
		}, msg)
	})
}

func TestStartSpanFromSqs(t *testing.T) {
	t.Run("no existing span", func(t *testing.T) {
		m := &mockSqs{}
		m.
			On("ReceiveMessageWithContext", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput"), mock.AnythingOfType("[]request.Option")).
			Return(&sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{
					{
						MessageId: aws.String("msg-id"),
					},
				},
			}, nil)

		s := sqsctx.New(m, Contexter())
		resp, err := s.ReceiveMessageWithContext(context.Background(), &sqs.ReceiveMessageInput{})
		assert.NoError(t, err)

		msg := resp.Messages[0]
		_, msgspan := StartSpanFromSqs(context.Background(), msg)

		assert.NotEmpty(t, msgspan.SerializeHeaders())
		m.AssertExpectations(t)
	})

	t.Run("existing span", func(t *testing.T) {
		ctx, span := beeline.StartSpan(context.Background(), "test")

		m := &mockSqs{}
		m.
			On("ReceiveMessageWithContext", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput"), mock.AnythingOfType("[]request.Option")).
			Return(&sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{
					{
						MessageId: aws.String("msg-id"),
						MessageAttributes: map[string]*sqs.MessageAttributeValue{
							"X-Honeycomb-Trace": {
								DataType:    aws.String("String"),
								StringValue: aws.String(span.SerializeHeaders()),
							},
						},
					},
				},
			}, nil)

		s := sqsctx.New(m, Contexter())
		resp, err := s.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{})
		assert.NoError(t, err)

		msg := resp.Messages[0]
		_, msgspan := StartSpanFromSqs(context.Background(), msg)

		rootSpanProps, err := propagation.UnmarshalTraceContext(span.SerializeHeaders())
		assert.NoError(t, err)

		msgSpanProps, err := propagation.UnmarshalTraceContext(msgspan.SerializeHeaders())
		assert.NoError(t, err)

		assert.Equal(t, rootSpanProps.TraceID, msgSpanProps.TraceID)
		m.AssertExpectations(t)
	})
}

type mockSqs struct {
	mock.Mock
	sqsiface.SQSAPI
}

func (m *mockSqs) SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, opts ...request.Option) (resp *sqs.SendMessageOutput, err error) {
	f := m.Called(ctx, input, opts)
	ret, _ := f.Get(0).(*sqs.SendMessageOutput)
	return ret, f.Error(1)
}

func (m *mockSqs) SendMessageBatchWithContext(ctx aws.Context, input *sqs.SendMessageBatchInput, opts ...request.Option) (resp *sqs.SendMessageBatchOutput, err error) {
	f := m.Called(ctx, input, opts)
	ret, _ := f.Get(0).(*sqs.SendMessageBatchOutput)
	return ret, f.Error(1)
}

func (m *mockSqs) ReceiveMessageWithContext(ctx aws.Context, input *sqs.ReceiveMessageInput, opts ...request.Option) (*sqs.ReceiveMessageOutput, error) {
	f := m.Called(ctx, input, opts)
	ret, _ := f.Get(0).(*sqs.ReceiveMessageOutput)
	return ret, f.Error(1)
}
