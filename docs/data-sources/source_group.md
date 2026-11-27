---
page_title: "betterstack_source_group Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A group that organizes related telemetry sources together.
---

# betterstack_source_group (Data Source)

A group that organizes related telemetry sources together.

## Example Usage

```terraform
data "betterstack_source_group" "example" {
  name = "Backend sources"
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

- `sort_index` (Number) Display order among the team's source groups. Lower values come first.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns the group.
