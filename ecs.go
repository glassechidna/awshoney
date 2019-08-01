package awshoney

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func ecsMap() map[string]string {
	m := map[string]string{}

	meta, err := getEcsMetadata()
	if err != nil {
		return m
	}

	m["aws.ecs.cluster"] = meta.Cluster
	m["aws.ecs.task.arn"] = meta.TaskARN
	m["aws.ecs.task.family"] = meta.Family
	m["aws.ecs.task.revision"] = meta.Revision
	m["aws.availability-zone"] = meta.AvailabilityZone
	m["aws.region"] = meta.AvailabilityZone[:len(meta.AvailabilityZone)-1]

	if os.Getenv("AWS_EXECUTION_ENV") == "AWS_ECS_FARGATE" {
		m["aws.ecs.launchtype"] = "fargate"
	} else {
		m["aws.ecs.launchtype"] = "ec2"
	}

	return m
}

type ecsTaskMetadata struct {
	Cluster       string `json:"Cluster"`
	TaskARN       string `json:"TaskARN"`
	Family        string `json:"Family"`
	Revision      string `json:"Revision"`
	DesiredStatus string `json:"DesiredStatus"`
	KnownStatus   string `json:"KnownStatus"`
	Containers    []struct {
		DockerID      string            `json:"DockerId"`
		Name          string            `json:"Name"`
		DockerName    string            `json:"DockerName"`
		Image         string            `json:"Image"`
		ImageID       string            `json:"ImageID"`
		Labels        map[string]string `json:"Labels"`
		DesiredStatus string            `json:"DesiredStatus"`
		KnownStatus   string            `json:"KnownStatus"`
		Limits        struct {
			CPU    int `json:"CPU"`
			Memory int `json:"Memory"`
		} `json:"Limits"`
		CreatedAt time.Time `json:"CreatedAt"`
		StartedAt time.Time `json:"StartedAt"`
		Type      string    `json:"Type"`
		Networks  []struct {
			NetworkMode   string   `json:"NetworkMode"`
			IPv4Addresses []string `json:"IPv4Addresses"`
		} `json:"Networks"`
		Volumes []struct {
			Source      string `json:"Source"`
			Destination string `json:"Destination"`
		} `json:"Volumes"`
	} `json:"Containers"`
	PullStartedAt    time.Time `json:"PullStartedAt"`
	PullStoppedAt    time.Time `json:"PullStoppedAt"`
	AvailabilityZone string    `json:"AvailabilityZone"`
}

func getEcsMetadata() (*ecsTaskMetadata, error) {
	url := fmt.Sprintf("%s/task", os.Getenv("ECS_CONTAINER_METADATA_URI"))
	r, err := MetadataClient.Get(url)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	data := ecsTaskMetadata{}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
