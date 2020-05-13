package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/urfave/cli/v2"
)

func getInstanceIds(r *ec2.DescribeInstancesOutput) []*string {
	instances := []*string{}
	for i := 0; i < len(r.Reservations); i++ {
		id := *r.Reservations[i].Instances[0].InstanceId
		instances = append(instances, aws.String(id))

	}
	return instances
}

func printRunningInstances(s []*string, tagkv string) {
	if len(s) < 1 {
		log.Print("No running instances with specified tag \"", tagkv, "\" found.")
	} else {
		fmt.Println("Instances running:", aws.StringValueSlice(s))
	}
}

func createClient() *ec2.EC2 {
	var region = "us-east-1"
	sess, err := session.NewSession(
		&aws.Config{Region: aws.String(region)},
	)
	if err != nil {
		fmt.Println("Error creating session ", err)
		panic(err)
	}
	return ec2.New(sess)
}

func describeInstances(ec2client *ec2.EC2, tagkv string) ([]*string, error) {
	t := strings.Split(tagkv, "=")
	tagkey := strings.Join([]string{"tag:", t[0]}, "")
	tagvalue := t[1]
	d := &ec2.DescribeInstancesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("pending"),
					aws.String("running"),
				},
			},
			&ec2.Filter{
				Name: aws.String(tagkey),
				Values: []*string{
					aws.String(tagvalue),
				},
			},
		},
	}
	reservations, err := ec2client.DescribeInstances(d)
	if err != nil {
		log.Fatal(err)
	}
	instanceIds := getInstanceIds(reservations)
	printRunningInstances(instanceIds, tagkv)

	return instanceIds, err
}

func terminateinstances(ec2client *ec2.EC2, instanceIds []*string) {
	fmt.Println("Terminating test EC2 instances: ", aws.StringValueSlice(instanceIds))
	terminateinstancesinput := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds,
	}
	_, err := ec2client.TerminateInstances(terminateinstancesinput)
	if err != nil {
		fmt.Println("Error terminating instances", err)
		panic(err)
	}
}

func main() {
	ec2client := createClient()

	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:  "tag",
			Value: "Name=Packer Builder",
			Usage: "Filter tag of EC2 instances to terminate in format: `TagName=TagValue`",
		},
	}

	app := &cli.App{
		Name:  "packer-ec2-cleanup",
		Usage: "Clean up stray EC2 instances, eg. Packer builds.",
		Commands: []*cli.Command{
			{
				Name:    "describe-instances",
				Aliases: []string{"di"},
				Usage:   "List EC2 instances in selected region",
				Flags:   cliFlags,
				Action: func(c *cli.Context) error {
					_, err := describeInstances(ec2client, c.String("tag"))
					if err != nil {
						log.Fatal(err)
					}
					return err
				},
			},
			{
				Name:    "terminate-instances",
				Aliases: []string{"ti"},
				Usage:   "Terminate EC2 instances",
				Flags:   cliFlags,
				Action: func(c *cli.Context) error {
					instanceIds, err := describeInstances(ec2client, c.String("tag"))
					if err != nil {
						log.Fatal(err)
					}
					if len(instanceIds) < 1 {
						log.Fatal("No running instances with specified tag \"", c.String("tag"), "\" to terminate.")
					}
					terminateinstances(ec2client, instanceIds)
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
