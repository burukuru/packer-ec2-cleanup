package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/burukuru/packer-ec2-cleanup/pec"
)

func HandleRequest(ctx context.Context) error {
	ec2client := pec.CreateClient()

	tag := os.Getenv("Tag")
	olderthan := os.Getenv("Olderthan")

	log.Printf("Tag %s", tag)
	m, err := time.ParseDuration(olderthan)
	instanceData, err := pec.DescribeInstances(ec2client, tag, m)
	if err != nil {
		log.Fatal(err)
	}
	if len(instanceData[0]) > 0 {
		pec.Terminateinstances(ec2client, instanceData[0])
		if len(instanceData[1]) > 0 {
			pec.DeleteKeyPair(ec2client, instanceData[1])
		}
	} else {
		log.Print("No running instances with specified tag \"", tag, "\" to terminate.")
	}
	return err
}

func main() {
	lambda.Start(HandleRequest)
}
