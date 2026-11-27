---
page_title: "betterstack_application_group Data Source - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  A group that organizes related error-tracking applications together.
---

# betterstack_application_group (Data Source)

A group that organizes related error-tracking applications together.

## Example Usage

```terraform
data "betterstack_application_group" "example" {
  name = "production-apps"
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

- `sort_index` (Number) Display order among the team's application groups. Lower values come first.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns the group.
