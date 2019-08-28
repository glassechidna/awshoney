package awshoney

import (
	"github.com/honeycombio/beeline-go/client"
	"github.com/honeycombio/libhoney-go"
	"net/http"
	"os"
	"strings"
	"time"
)

const execEnvEcs = "AWS_ECS_"
const execEnvLambda = "AWS_Lambda_"

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
	env := os.Getenv("AWS_EXECUTION_ENV")
	if strings.HasPrefix(env, execEnvEcs) {
		return "ecs"
	} else if strings.HasPrefix(env, execEnvLambda) {
		return "lambda"
	} else {
		return "unknown"
	}
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
