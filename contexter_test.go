package awshoney

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/glassechidna/awsctx/service/s3ctx"
	"github.com/honeycombio/beeline-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mocks3 struct {
	mock.Mock
	s3iface.S3API
}

func (m *mocks3) PutObjectWithContext(ctx context.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	f := m.Called(ctx, input, opts)
	out, _ := f.Get(0).(*s3.PutObjectOutput)
	return out, f.Error(1)
}

func TestContexter(t *testing.T) {
	api := &mocks3{}
	ctxual := s3ctx.New(api, Contexter())

	api.
		On("PutObjectWithContext", mock.Anything, &s3.PutObjectInput{Bucket: aws.String("bucket")}, mock.Anything).
		Return(&s3.PutObjectOutput{ETag: aws.String("etag")}, nil)

	honeycombTriggered := false
	beeline.Init(beeline.Config{
		STDOUT: true,
		PresendHook: func(ev map[string]interface{}) {
			if ev["aws.service"] == "s3" && ev["aws.action"] == "PutObject" {
				honeycombTriggered = true
			}
		},
	})

	ctx, _ := beeline.StartSpan(context.Background(), "test")

	out, err := ctxual.PutObjectWithContext(ctx, &s3.PutObjectInput{Bucket: aws.String("bucket")})
	assert.NoError(t, err)
	assert.Equal(t, "etag", *out.ETag)
	assert.True(t, honeycombTriggered)
}
