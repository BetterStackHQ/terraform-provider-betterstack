resource "betterstack_aws_log_group_subscription" "example" {
  source_id  = "123456"
  region     = "us-east-1"
  name       = "/aws/lambda/checkout"
  subscribed = true
}
