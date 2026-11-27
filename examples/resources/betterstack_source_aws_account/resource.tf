# Connect an AWS account through Better Stack's CloudFormation stack
resource "betterstack_source" "aws" {
  name     = "AWS production"
  platform = "aws"
}

resource "aws_cloudformation_stack" "better_stack" {
  name         = "better-stack-integration"
  template_url = "https://better-stack-cloudformation.s3.amazonaws.com/better-stack-full.yaml"
  capabilities = ["CAPABILITY_NAMED_IAM"]

  parameters = {
    ClusterId   = betterstack_source.aws.cluster
    SourceToken = betterstack_source.aws.token
    SourceId    = betterstack_source.aws.id
  }

  lifecycle {
    ignore_changes = [parameters["SourceToken"]]
  }
}

resource "betterstack_source_aws_account" "aws" {
  source_id               = betterstack_source.aws.id
  aws_role_arn            = aws_cloudformation_stack.better_stack.outputs["IntegrationRoleArn"]
  aws_external_id         = aws_cloudformation_stack.better_stack.outputs["ExternalId"]
  aws_auto_sub_log_groups = true
}

# Reuse an AWS account another source connected
resource "betterstack_source" "aws_staging" {
  name     = "AWS staging"
  platform = "aws"
}

resource "betterstack_source_aws_account" "aws_staging" {
  source_id      = betterstack_source.aws_staging.id
  aws_account_id = "1234"
}
