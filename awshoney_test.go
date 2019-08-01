package awshoney

import (
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestMap_Lambda(t *testing.T) {
	defer setmap(map[string]string{
		"AWS_REGION":                      "region",
		"AWS_EXECUTION_ENV":               "AWS_Lambda_runtime",
		"AWS_LAMBDA_FUNCTION_NAME":        "name",
		"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": "memory",
		"AWS_LAMBDA_FUNCTION_VERSION":     "version",
		"_HANDLER":                        "handler",
	})()

	assert.Equal(t, map[string]string{
		"aws.env":            "lambda",
		"aws.region":         "region",
		"aws.lambda.handler": "handler",
		"aws.lambda.name":    "name",
		"aws.lambda.runtime": "runtime",
		"aws.lambda.version": "version",
		"aws.lambda.memory":  "memory",
	}, Map())
}

func TestMap_Ecs(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com/task", nil)
	assert.NoError(t, err)

	body, err := ioutil.ReadFile("testdata/ecs_metadata.json")
	assert.NoError(t, err)

	orig := MetadataClient
	defer func() {
		MetadataClient = orig
	}()

	t.Run("ecs ec2", func(t *testing.T) {
		transport := &mockTransport{}
		transport.On("RoundTrip", req).Return(&http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, nil)
		MetadataClient = &http.Client{Transport: transport}

		defer setmap(map[string]string{
			"AWS_EXECUTION_ENV":          "AWS_ECS_EC2",
			"ECS_CONTAINER_METADATA_URI": "https://example.com",
		})()

		assert.Equal(t, map[string]string{
			"aws.env":               "ecs",
			"aws.region":            "us-east-1",
			"aws.availability-zone": "us-east-1c",
			"aws.ecs.cluster":       "default",
			"aws.ecs.launchtype":    "ec2",
			"aws.ecs.task.arn":      "arn:aws:ecs:us-east-1:01234567890:task/default/3f3b08db6c984e0f98f05e5d3af242c3",
			"aws.ecs.task.family":   "worker",
			"aws.ecs.task.revision": "4",
		}, Map())
	})

	t.Run("ecs fargate", func(t *testing.T) {
		transport := &mockTransport{}
		transport.On("RoundTrip", req).Return(&http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, nil)
		MetadataClient = &http.Client{Transport: transport}

		defer setmap(map[string]string{
			"AWS_EXECUTION_ENV":          "AWS_ECS_FARGATE",
			"ECS_CONTAINER_METADATA_URI": "https://example.com",
		})()

		m := Map()
		spew.Dump(m)
		assert.Equal(t, "fargate", m["aws.ecs.launchtype"])
	})
}

func setmap(m map[string]string) func() {
	for k, v := range m {
		os.Setenv(k, v)
	}

	return func() {
		for k := range m {
			os.Unsetenv(k)
		}
	}
}

type mockTransport struct {
	mock.Mock
	http.RoundTripper
}

func (m *mockTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	f := m.Called(request)
	resp, _ := f.Get(0).(*http.Response)
	return resp, f.Error(1)
}
