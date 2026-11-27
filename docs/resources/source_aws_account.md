---
page_title: "betterstack_source_aws_account Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  The AWS account connected to an `aws` source, through the IAM role from Better Stack's CloudFormation stack.
---

# betterstack_source_aws_account (Resource)

The AWS account connected to an `aws` source, through the IAM role from Better Stack's CloudFormation stack.

This resource takes over an object that already exists in Better Stack, named by its `source_id`, and
applies the arguments you set to it. It never creates one.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

-> Replaces `logtail_source_aws_account`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Connect an AWS account through Better Stack's CloudFormation stack

```terraform
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
```

### Reuse an AWS account another source connected

```terraform
resource "betterstack_source" "aws_staging" {
  name     = "AWS staging"
  platform = "aws"
}

resource "betterstack_source_aws_account" "aws_staging" {
  source_id      = betterstack_source.aws_staging.id
  aws_account_id = "1234"
}
```

## Schema

Configure exactly one of these: `aws_role_arn` with `aws_external_id`, or `aws_account_id`. The provider checks this when it plans the resource, before it creates or replaces anything.

### Required

- `source_id` (String) ID of the object in Better Stack this resource takes over. Changing it makes Terraform take over that object instead. The previous one keeps its settings in Better Stack.

### Optional

- `aws_account_id` (String) Better Stack's own ID for an AWS account that another `aws` source in your organization connected, to reuse that connection. It differs from the 12-digit AWS account ID. At least one of `aws_role_arn` or `aws_account_id` must be configured. Better Stack never returns it, so Terraform keeps the value you set.
- `aws_role_arn` (String, Sensitive) ARN of the IAM role from Better Stack's CloudFormation stack. At least one of `aws_role_arn` or `aws_account_id` must be configured. Better Stack never returns it, so Terraform keeps the value you set.
- `aws_external_id` (String, Sensitive, Computed) External ID from Better Stack's CloudFormation stack. Better Stack never returns it, so Terraform keeps the value you set.
- `aws_auto_sub_log_groups` (Boolean, Computed) Whether to subscribe new CloudWatch log groups automatically when they're discovered.
- `team_id` (String) Team that owns the source, such as `t123`. Defaults to the provider's `team_id`. Set it when the object is in another team. Changing it replaces the resource, which takes over the object in that team instead.

## Import

Import an existing source AWS account by its ID. Get the ID via [List sources](https://betterstack.com/docs/logs/api/list-all-existing-sources/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_source_aws_account.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_source_aws_account.example 12345
```
