---
page_title: "betterstack_source_aws_log_group Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.
---

# betterstack_source_aws_log_group (Data Source)

An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.

## Example Usage

```terraform
data "betterstack_source_aws_log_group" "example" {
  source_id = "12345"
  id        = "12345"
}
```

## Schema

### Address

- `source_id` (String, Required) Id of the parent this source belongs to. Changing it replaces the resource.

### Required

- `id` (String) The id of the source_aws_log_group to look up.

### Read-Only

- `region` (String) AWS region of the log group.
- `name` (String) CloudWatch log-group name.
- `subscribed` (Boolean) `true` subscribes the log group. `false` keeps it unsubscribed, even when automatic subscription is on.
- `created_at` (String (RFC 3339)) When the override was created.
- `updated_at` (String (RFC 3339)) When the override was last updated.

## Endpoints

- `GET /api/v2/sources/:source_id/aws-log-group-subscriptions`
- `GET /api/v2/sources/:source_id/aws-log-group-subscriptions/:id`
- `POST /api/v2/sources/:source_id/aws-log-group-subscriptions`
- `PATCH /api/v2/sources/:source_id/aws-log-group-subscriptions/:id`
- `DELETE /api/v2/sources/:source_id/aws-log-group-subscriptions/:id`
