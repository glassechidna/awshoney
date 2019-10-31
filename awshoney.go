package awshoney

import (
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/honeycombio/beeline-go/client"
	"github.com/honeycombio/libhoney-go"
)

// Adds aws.* fields to all traces and spans recorded by c. If c is nil,
// the default client will be used. Usually you will invoke this right after
// beeline.Init()
func AddFieldsToClient(c *libhoney.Client) {
	if c == nil {
		c = client.Get()
	}

	c.Add(Map())
}

func execEnv() string {
	if os.Getenv("ECS_CONTAINER_METADATA_URI") != "" {
		return "ecs"
	} else if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		return "lambda"
	}
	sess, err := session.NewSession()
	if err != nil {
		return "unknown"
	}
	metadata := ec2metadata.New(sess)
	if metadata.Available() {
		return "ec2"
	}
	return "unknown"
}

func Map() map[string]string {
	var m map[string]string

	env := execEnv()
	switch env {
	case "ecs":
		m = ecsMap()
	case "lambda":
		m = lambdaMap()
	case "ec2":
		m = ec2Map()
	}
	m["aws.env"] = env

	return m
}

var MetadataClient = &http.Client{
	Timeout: 3 * time.Second,
}
