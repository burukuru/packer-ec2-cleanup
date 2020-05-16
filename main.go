package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/urfave/cli/v2"
)

func getInstanceData(r *ec2.DescribeInstancesOutput, instanceAge time.Duration) [][]*string {
	instances := []*string{}
	keynames := []*string{}
	for i := 0; i < len(r.Reservations); i++ {
		launchTime := *r.Reservations[i].Instances[0].LaunchTime
		if time.Now().UTC().Sub(launchTime).Minutes() > instanceAge.Minutes() {
			instances = append(instances, aws.String(*r.Reservations[i].Instances[0].InstanceId))
			// Check if instance has SSH key
			if r.Reservations[i].Instances[0].KeyName != nil &&
				len(*r.Reservations[i].Instances[0].KeyName) > 0 {
				keynames = append(keynames, aws.String(*r.Reservations[i].Instances[0].KeyName))
			}
		}
	}
	return [][]*string{instances, keynames}
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

func describeInstances(ec2client *ec2.EC2, tagkv string, instanceAge time.Duration) ([][]*string, error) {
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
	instanceData := getInstanceData(reservations, instanceAge)
	printRunningInstances(instanceData[0], tagkv)

	return instanceData, err
}

func deleteKeyPair(ec2client *ec2.EC2, KeyName []*string) {
	fmt.Println("Deleting SSH key pair: ", aws.StringValueSlice(KeyName))
	deleteKeypairInput := &ec2.DeleteKeyPairInput{
		KeyName: KeyName[0],
	}
	_, err := ec2client.DeleteKeyPair(deleteKeypairInput)
	if err != nil {
		fmt.Println("Error deleting keypair", err)
		panic(err)
	}
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
		&cli.StringFlag{
			Name:  "olderthan",
			Value: "60m",
			Usage: "Minimum age of instance that will be terminated, in minutes",
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
					m, err := time.ParseDuration(c.String("olderthan"))
					if err != nil {
						log.Fatal(err)
					}
					_, err = describeInstances(ec2client, c.String("tag"), m)
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
					m, err := time.ParseDuration(c.String("olderthan"))
					if err != nil {
						log.Fatal(err)
					}
					instanceData, err := describeInstances(ec2client, c.String("tag"), m)
					if err != nil {
						log.Fatal(err)
					}
					if len(instanceData[0]) > 0 {
						terminateinstances(ec2client, instanceData[0])
						if len(instanceData[1]) > 0 {
							deleteKeyPair(ec2client, instanceData[1])
						}
					} else {
						log.Print("No running instances with specified tag \"", c.String("tag"), "\" to terminate.")
					}
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
