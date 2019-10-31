package awshoney

import (
	"net/http"
	"os"
	"time"

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
	return "unknown"
}

func Map() map[string]string {
	m := map[string]string{}

	env := execEnv()
	m["aws.env"] = env

	var envm map[string]string
	if env == "ecs" {
		envm = ecsMap()
	} else if env == "lambda" {
		envm = lambdaMap()
	}

	for k, v := range envm {
		m[k] = v
	}

	return m
}

var MetadataClient = &http.Client{
	Timeout: 3 * time.Second,
}
