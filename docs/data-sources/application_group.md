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

### Read-Only

- `sort_index` (Number) Display order among the team's application groups. `null` when never set.
- `team_name` (String) Name of the team that owns the group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v1/application-groups`
- `GET /api/v1/application-groups/:id`
- `POST /api/v1/application-groups`
- `PATCH /api/v1/application-groups/:id`
- `DELETE /api/v1/application-groups/:id`
