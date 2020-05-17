package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/burukuru/packer-ec2-cleanup/pec"
)

type MyEvent struct {
	Tag       string `json:"tag"`
	Olderthan string `json:"olderthan"`
}

func HandleRequest(ctx context.Context, event MyEvent) error {
	ec2client := pec.CreateClient()

	log.Print("Tag %s", event.Tag)
	m, err := time.ParseDuration(event.Olderthan)
	instanceData, err := pec.DescribeInstances(ec2client, event.Tag, m)
	if err != nil {
		log.Fatal(err)
	}
	if len(instanceData[0]) > 0 {
		pec.Terminateinstances(ec2client, instanceData[0])
		if len(instanceData[1]) > 0 {
			pec.DeleteKeyPair(ec2client, instanceData[1])
		}
	} else {
		log.Print("No running instances with specified tag \"", event.Tag, "\" to terminate.")
	}
	return err
}

func main() {
	lambda.Start(HandleRequest)
}
