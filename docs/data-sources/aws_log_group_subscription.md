---
page_title: "betterstack_aws_log_group_subscription Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.
---

# betterstack_aws_log_group_subscription (Data Source)

An override that subscribes a CloudWatch log group or keeps it unsubscribed, even before the group is discovered.

## Example Usage

```terraform
data "betterstack_aws_log_group_subscription" "example" {
  source_id = "12345"
  id        = "12345"
}
```

## Schema

### Address

- `source_id` (String, Required) The ID of the parent this source belongs to. Changing it replaces the resource.

### Required

- `id` (String) The ID of the AWS log group subscription to look up.

### Optional

- `team_id` (String) Team that owns the source, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `region` (String) AWS region of the log group.
- `name` (String) CloudWatch log-group name.
- `subscribed` (Boolean) `true` subscribes the log group. `false` keeps it unsubscribed, even when automatic subscription is on.
- `created_at` (String (RFC 3339)) When the override was created.
- `updated_at` (String (RFC 3339)) When the override was last updated.
