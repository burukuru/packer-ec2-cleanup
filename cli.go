package main

import (
	"log"
	"os"
	"time"

	"github.com/burukuru/packer-ec2-cleanup/pec"
	"github.com/urfave/cli/v2"
)

func main() {
	ec2client := pec.CreateClient()

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
					_, err = pec.DescribeInstances(ec2client, c.String("tag"), m)
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
					instanceData, err := pec.DescribeInstances(ec2client, c.String("tag"), m)
					if err != nil {
						log.Fatal(err)
					}
					if len(instanceData[0]) > 0 {
						pec.Terminateinstances(ec2client, instanceData[0])
						if len(instanceData[1]) > 0 {
							pec.DeleteKeyPair(ec2client, instanceData[1])
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
