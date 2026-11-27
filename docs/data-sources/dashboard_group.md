---
page_title: "betterstack_dashboard_group Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related dashboards together.
---

# betterstack_dashboard_group (Data Source)

A group that organizes related dashboards together.

## Example Usage

```terraform
data "betterstack_dashboard_group" "example" {
  name = "Production Monitoring"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Group name.

### Optional

- `team_id` (String) Team that owns the group, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns the group.
