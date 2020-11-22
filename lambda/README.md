# Usage

go build   
zip lambda.zip lambda  
terraform apply

# Overview
The Terraform code will add:
- a Lambda function that will delete EC2 instances on a regular basis
- IAM permissions for the Lambda to do full EC2 access and create CloudWatch log groups/streams


