---
page_title: "betterstack_aws_log_group_subscription Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.
---

# betterstack_aws_log_group_subscription (Resource)

An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.

-> Replaces `logtail_source_aws_log_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a log group subscription

```terraform
resource "betterstack_aws_log_group_subscription" "example" {
  source_id  = "123456"
  region     = "us-east-1"
  name       = "/aws/lambda/checkout"
  subscribed = true
}
```

### Keep a noisy log group unsubscribed

```terraform
resource "betterstack_aws_log_group_subscription" "keep_a_noisy_log_group_unsubscribed" {
  source_id  = "123456"
  region     = "eu-west-1"
  name       = "/aws/lambda/healthcheck"
  subscribed = false
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `source_id` (String, Required) The ID of the parent this source belongs to. Changing it replaces the resource.

### Required

- `region` (String, Forces replacement) AWS region of the log group.
- `name` (String, Forces replacement) CloudWatch log-group name.
- `subscribed` (Boolean) `true` subscribes the log group. `false` keeps it unsubscribed, even when automatic subscription is on.

### Optional

- `team_id` (String) Team that owns the source, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this AWS log group subscription.
- `created_at` (String (RFC 3339)) When the override was created.
- `updated_at` (String (RFC 3339)) When the override was last updated.

## Import

Import an existing AWS log group subscription by its ID as `source_id/id`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_aws_log_group_subscription.example
  id = "<source_id>/12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_aws_log_group_subscription.example <source_id>/12345
```
