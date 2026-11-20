---
page_title: "betterstack_severity_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity group: a set of related severities.
---

# betterstack_severity_group (Data Source)

A severity group: a set of related severities.

## Example Usage

```terraform
data "betterstack_severity_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the severity_group to look up.

### Read-Only

- `team_name` (String) Name of the team that owns this group.
- `name` (String) Group name.
- `sort_index` (Number) Display order of the group among the team's severity groups.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/urgency-groups`
- `GET /api/v2/urgency-groups/:id`
- `POST /api/v2/urgency-groups`
- `PATCH /api/v2/urgency-groups/:id`
- `DELETE /api/v2/urgency-groups/:id`
