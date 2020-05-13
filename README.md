# packer-ec2-cleanup
Clean up stale EC2 instances from Packer builds

# Usage
GO111MODULE=on go mod init
go run main.go -h

# TODO
- Go program to delete EC2 instance
  - filter by launch time older than
  - delete associated SSH keys
- Terraform module to deploy to AWS Lambda
