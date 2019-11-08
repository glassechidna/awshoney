# `awshoney` - AWS fields for your Honeycomb.io traces

[Honeycomb.io][honeycombio] is an application performance monitoring service. They
recommend enriching spans and traces with as much metadata as you see fit. We've
found adding fields describing the AWS Lambda function or AWS ECS task our apps
very helpful.

[honeycombio]: https://honeycomb.io

## Fields added to every trace

```go
package main

import (
	"github.com/glassechidna/awshoney"
	"github.com/honeycombio/beeline-go"
)

func main() {
	beeline.Init(beeline.Config{
		WriteKey: "yourkey",
		Dataset:  "yourdataset",
	})

	awshoney.AddFieldsToClient(nil)
}
```

Traces inside AWS Lambda functions have the following fields added:

```go
map[string]string{
	"aws.env":                      "lambda",
	"aws.region":                   "us-east-1",
	"aws.lambda.handler":           "handlerName",
	"aws.lambda.name":              "functionName",
	"aws.lambda.runtime":           "go1.x",
	"aws.lambda.version":           "$LATEST",
	"aws.lambda.memory":            "128", 
	"aws.lambda.execution-context": "9d4ee6310ced4ef48754dcfc55754f82",
}
```

Traces inside AWS ECS tasks have the following fields added:

```go
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
```

## Fields added to AWS API calls

Calls to AWS APIs can be instrumented to automatically record spans, tagged with
the AWS service and action called. That works as follows:

```go
package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/glassechidna/awsctx/service/s3ctx"
	"github.com/glassechidna/awshoney"
	"github.com/honeycombio/beeline-go"
)

func main() {
	// obviously it won't be too helpful if you do this. you'd want to
	// pass in a context from a real span
	ctx, _ := beeline.StartSpan(context.Background(), "example")

	sess := session.Must(session.NewSession())
	baseApi := s3.New(sess)

	api := s3ctx.New(baseApi, awshoney.Contexter)

	// the other methods that aren't WithContext won't pass through
	// the honeycomb trace id
	_, _ = api.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: aws.String("bucket-name"),
	})

	/*
	The above will create a new span named `aws.api` with the following
	span-level fields added:

	map[string]string{
		"aws.service": "s3",
		"aws.action":  "ListObjects",
	}
	*/
}

```


## Bonus: SQS!

If you use the `sqsctx.SQS` (as described above) when performing `SendMessageWithContext`
or `SendMessageBatchWithContext` actions, messages will be annotated with the Honeycomb
trace ID for cross-system tracing. On the "receiving end", you should do:

```go
package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/glassechidna/awshoney"
	"github.com/honeycombio/beeline-go/propagation"
)

func main() {
	sess := session.Must(session.NewSession())
	baseApi := sqs.New(sess)

	// you don't need to use the sqsctx.SQS wrapper here (but it won't hurt)
	resp, _ := baseApi.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:			  aws.String("queue-url"),
		// by default sqs won't retrieve message attributes
		MessageAttributeNames: []*string{aws.String(propagation.TracePropagationHTTPHeader)},
	})

	msg := resp.Messages[0]
	ctx, _ := awshoney.StartSpanFromSqs(context.Background(), msg)
	// do something with ctx
}
```