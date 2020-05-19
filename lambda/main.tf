variable "lambda_function_name" {
  type    = string
  default = "pec"
}

variable "terminate_tag" {
  type    = string
  default = "Name=Packer Builder"
}

variable "olderthan" {
  type    = string
  default = "60m"
  # default = "0s"
}

resource "aws_lambda_function" "pec" {
  function_name = var.lambda_function_name
  role          = aws_iam_role.iam_for_lambda.arn
  filename      = "lambda.zip"
  handler       = "lambda"

  source_code_hash = filebase64sha256("lambda.zip")

  runtime = "go1.x"

  environment {
    variables = {
      Tag       = var.terminate_tag,
      Olderthan = var.olderthan
    }
  }

}

resource "aws_iam_role" "iam_for_lambda" {
  name = "iam_for_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_iam_policy" "AmazonEC2FullAccess" {
  arn = "arn:aws:iam::aws:policy/AmazonEC2FullAccess"
}

resource "aws_iam_role_policy_attachment" "allow_ec2_fullaccess" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = data.aws_iam_policy.AmazonEC2FullAccess.arn

}


resource "aws_cloudwatch_log_group" "pec" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 7
}

resource "aws_iam_policy" "lambda_logging" {
  name        = "lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a Lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF

}

resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.iam_for_lambda.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}
