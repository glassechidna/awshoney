package awshoney

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

func ec2Map() map[string]string {
	m := map[string]string{}

	sess, err := session.NewSession()
	if err != nil {
		return m
	}
	metadata := ec2metadata.New(sess)

	id, err := metadata.GetInstanceIdentityDocument()
	if err != nil {
		return m
	}

	m["aws.ec2.image-id"] = id.ImageID
	m["aws.ec2.instance-type"] = id.InstanceType
	m["aws.ec2.instance-id"] = id.InstanceID
	m["aws.availability-zone"] = id.AvailabilityZone
	m["aws.region"] = id.Region

	return m
}
