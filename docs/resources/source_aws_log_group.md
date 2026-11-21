---
page_title: "betterstack_source_aws_log_group Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.
---

# betterstack_source_aws_log_group (Resource)

An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.

-> Replaces `logtail_source_aws_log_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_source_aws_log_group" "example" {
  source_id  = "95"
  region     = "us-east-1"
  name       = "/aws/lambda/checkout"
  subscribed = true
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `source_id` (String, Required) Id of the parent this source belongs to. Changing it replaces the resource.

### Required

- `region` (String) AWS region of the log group. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `name` (String) CloudWatch log-group name. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `subscribed` (Boolean) `true` subscribes the log group. `false` keeps it unsubscribed, even when automatic subscription is on.

### Read-Only

- `id` (String) The id of this source_aws_log_group.
- `created_at` (String (RFC 3339)) When the override was created.
- `updated_at` (String (RFC 3339)) When the override was last updated.

## Endpoints

- `GET /api/v2/sources/:source_id/aws-log-group-subscriptions`
- `GET /api/v2/sources/:source_id/aws-log-group-subscriptions/:id`
- `POST /api/v2/sources/:source_id/aws-log-group-subscriptions`
- `PATCH /api/v2/sources/:source_id/aws-log-group-subscriptions/:id`
- `DELETE /api/v2/sources/:source_id/aws-log-group-subscriptions/:id`

## Import

```shell
terraform import betterstack_source_aws_log_group.example <source_id>/12345
```
