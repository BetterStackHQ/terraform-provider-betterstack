resource "betterstack_source_aws_log_group" "example" {
  source_id  = "95"
  region     = "us-east-1"
  name       = "/aws/lambda/checkout"
  subscribed = true
}
