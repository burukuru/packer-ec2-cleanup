package pec

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetInstanceData(r *ec2.DescribeInstancesOutput, instanceAge time.Duration) [][]*string {
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

func PrintRunningInstances(s []*string, tagkv string) {
	if len(s) < 1 {
		log.Print("No running instances with specified tag \"", tagkv, "\" found.")
	} else {
		fmt.Println("Instances running:", aws.StringValueSlice(s))
	}
}

func CreateClient() *ec2.EC2 {
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

func DescribeInstances(ec2client *ec2.EC2, tagkv string, instanceAge time.Duration) ([][]*string, error) {
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
	instanceData := GetInstanceData(reservations, instanceAge)
	PrintRunningInstances(instanceData[0], tagkv)

	return instanceData, err
}

func DeleteKeyPair(ec2client *ec2.EC2, KeyName []*string) {
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

func Terminateinstances(ec2client *ec2.EC2, instanceIds []*string) {
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
