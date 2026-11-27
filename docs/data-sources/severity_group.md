---
page_title: "betterstack_severity_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity group, a set of related severities.
---

# betterstack_severity_group (Data Source)

A severity group, a set of related severities.

## Example Usage

```terraform
data "betterstack_severity_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the severity group to look up.

### Optional

- `team_id` (String) Team that owns this group, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Group name.
- `sort_index` (Number) Display order of the group among the team's severity groups. Lower values come first.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this group.
