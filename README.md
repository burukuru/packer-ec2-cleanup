# packer-ec2-cleanup
Clean up stale EC2 instances from Packer builds

# Usage
```
GO111MODULE=on go mod init
go build
./packer-ec2-cleanup -h
```

Describe instance:
```
./packer-ec2-cleanup di
```

Terminate instance older than 30 minutes:
```
./packer-ec2-cleanup ti --olderthan 30m
```

# TODO
- Go program to delete EC2 instance
  - delete associated SSH keys
- Terraform module to deploy to AWS Lambda
