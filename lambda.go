package awshoney

import (
	"os"
	"strings"
)

func lambdaMap() map[string]string {
	m := map[string]string{}

	m["aws.region"] = os.Getenv("AWS_REGION")
	m["aws.lambda.runtime"] = strings.TrimPrefix(os.Getenv("AWS_EXECUTION_ENV"), "AWS_Lambda_")
	m["aws.lambda.name"] = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	m["aws.lambda.memory"] = os.Getenv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE")
	m["aws.lambda.version"] = os.Getenv("AWS_LAMBDA_FUNCTION_VERSION")
	m["aws.lambda.handler"] = os.Getenv("_HANDLER")

	return m
}
