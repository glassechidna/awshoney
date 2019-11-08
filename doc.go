// Honeycomb.io is an application performance monitoring service. They recommend
// enriching spans and traces with as much metadata as you see fit. We've found
// adding fields describing the AWS Lambda function, AWS ECS task and EC2 instances
// our apps run in very helpful.
//
// Traces inside AWS ECS tasks have the following fields added:
//
//  map[string]string{
//    "aws.env":               "ecs",
//    "aws.region":            "us-east-1",
//    "aws.availability-zone": "us-east-1c",
//    "aws.ecs.cluster":       "default",
//    "aws.ecs.launchtype":    "ec2",
//    "aws.ecs.task.arn":      "arn:aws:ecs:us-east-1:01234567890:task/default/3f3b08db6c984e0f98f05e5d3af242c3",
//    "aws.ecs.task.family":   "worker",
//    "aws.ecs.task.revision": "4",
//  }
//
// Traces inside AWS EC2 instances have the following fields added:
//
//  map[string]string{
//    "aws.env":               "ecs",
//    "aws.region":            "us-east-1",
//    "aws.availability-zone": "us-east-1c",
//    "aws.ec2.image-id":      "ami-01234abc",
//    "aws.ec2.instance-type": "c5.xlarge",
//    "aws.ec2.instance-id":   "i-0012456abc",
//  }
//
// Traces inside AWS Lambda functions have the following fields added:
//
//  map[string]string{
//    "aws.env":                      "lambda",
//    "aws.region":                   "us-east-1",
//    "aws.lambda.handler":           "handlerName",
//    "aws.lambda.name":              "functionName",
//    "aws.lambda.runtime":           "go1.x",
//    "aws.lambda.version":           "$LATEST",
//    "aws.lambda.memory":            "128",
//    "aws.lambda.execution-context": "9d4ee6310ced4ef48754dcfc55754f82",
//  }
//
// awshoney.WrapLambda() adds the following fields to Lambda functions:
//
//  map[string]string{
//    "aws.lambda.cold-start":         "false",
//    "aws.lambda.invocation-counter": "65",
//    "aws.lambda.request-id":         "abcde6310ced4ef48754dcfc55754f82",
//    "aws.lambda.invoked-version":    "$LATEST (or '5', or 'live', etc)",
//  }
//
// Adding fields to AWS API calls
//
// Calls to AWS APIs can be instrumented to automatically record spans, tagged with
// the AWS service and action called. That works as follows:
//
//  func main() {
//    // obviously it won't be too helpful if you do this. you'd want to
//    // pass in a context from a real span
//    ctx, _ := beeline.StartSpan(context.Background(), "example")
//
//    sess := session.Must(session.NewSession())
//    baseApi := s3.New(sess)
//
//    api := s3ctx.New(baseApi, awshoney.Contexter)
//
//    // the other methods that aren't WithContext won't pass through
//    // the honeycomb trace id
//    _, _ = api.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
//      Bucket: aws.String("bucket-name"),
//    })
//
//    /*
//      The above will create a new span named `aws.api` with the following
//      span-level fields added:
//
//      map[string]string{
//        "aws.service": "s3",
//        "aws.action":  "ListObjects",
//      }
//    */
//  }
//
//  Bonus SQS
//
//  If you use the `sqsctx.SQS` (as described above) when performing
//  `SendMessageWithContext` or `SendMessageBatchWithContext` actions, messages
//  will be annotated with the Honeycomb trace ID for cross-system tracing. On
//  the "receiving end", you should do:
//
//  func main() {
//    sess := session.Must(session.NewSession())
//    baseApi := sqs.New(sess)
//
//    // you don't need to use the sqsctx.SQS wrapper here (but it won't hurt)
//    resp, _ := baseApi.ReceiveMessage(&sqs.ReceiveMessageInput{
//      QueueUrl:              aws.String("queue-url"),
//      // by default sqs won't retrieve message attributes
//      MessageAttributeNames: []*string{aws.String(propagation.TracePropagationHTTPHeader)},
//    })
//
//    msg := resp.Messages[0]
//    ctx, _ := awshoney.StartSpanFromSqs(context.Background(), msg)
//    // do something with ctx
//  }
//  .
package awshoney
