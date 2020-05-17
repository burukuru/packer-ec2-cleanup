package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

type MyEvent struct {
	Tag       string `json:"tag"`
	Olderthan string `json:"olderthan"`
}

func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	ec2client := pec.createClient()

	log.Print("Tag %s", name.Tag)
	instanceData, err := pec.describeInstances(ec2client, c.String("tag"), m)
	if err != nil {
		log.Fatal(err)
	}
	if len(instanceData[0]) > 0 {
		pec.terminateinstances(ec2client, instanceData[0])
		if len(instanceData[1]) > 0 {
			pec.deleteKeyPair(ec2client, instanceData[1])
		}
	} else {
		log.Print("No running instances with specified tag \"", c.String("tag"), "\" to terminate.")
	}
	return err
}

func main() {
	lambda.Start(HandleRequest)
}
