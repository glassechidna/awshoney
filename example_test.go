package awshoney_test

import (
	"github.com/glassechidna/awshoney"
	"github.com/honeycombio/beeline-go"
)

func ExampleUsage() {
	beeline.Init(beeline.Config{
		WriteKey:    "yourkey",
		Dataset:     "yourdataset",
		PresendHook: awshoney.PresendHook,
	})

	// alternatively, if you have other presend hooks:
	beeline.Init(beeline.Config{
		WriteKey: "yourkey",
		Dataset:  "yourdataset",
		PresendHook: awshoney.ComposePresendHooks(awshoney.PresendHook, func(m map[string]interface{}) {
			// ...
		}),
	})

	/*
		Traces inside AWS Lambda functions have the following fields added:

		map[string]string{
			"aws.env":            "lambda",
			"aws.region":         "us-east-1",
			"aws.lambda.handler": "handlerName",
			"aws.lambda.name":    "functionName",
			"aws.lambda.runtime": "go1.x",
			"aws.lambda.version": "$LATEST",
			"aws.lambda.memory":  "128",
		}

		Traces inside AWS ECS tasks have the following fields added:

		map[string]string{
			"aws.env":               "ecs",
			"aws.region":            "us-east-1",
			"aws.availability-zone": "us-east-1c",
			"aws.ecs.cluster":       "default",
			"aws.ecs.launchtype":    "ec2",
			"aws.ecs.task.arn":      "arn:aws:ecs:us-east-1:01234567890:task/default/3f3b08db6c984e0f98f05e5d3af242c3",
			"aws.ecs.task.family":   "worker",
			"aws.ecs.task.revision": "4",
		}
	*/
}
