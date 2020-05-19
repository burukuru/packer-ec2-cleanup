variable "trigger_ec2_cleanup_frequency" {
  default = "cron(0 * * * ? *)"
}

resource "aws_cloudwatch_event_rule" "trigger_ec2_cleanup" {
  name = "trigger_ec2_cleanup"

  schedule_expression = var.trigger_ec2_cleanup_frequency
}

resource "aws_cloudwatch_event_target" "ec2_cleanup_lambda" {
  rule = aws_cloudwatch_event_rule.trigger_ec2_cleanup.name
  arn  = aws_lambda_function.pec.arn
}

resource "aws_lambda_permission" "allow_cloudwatch_event" {
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.trigger_ec2_cleanup.arn
}
