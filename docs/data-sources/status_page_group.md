---
page_title: "betterstack_status_page_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group of related status pages. It belongs to the organization, not to a team.
---

# betterstack_status_page_group (Data Source)

A group of related status pages. It belongs to the organization, not to a team.

## Example Usage

```terraform
data "betterstack_status_page_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the status_page_group to look up.

### Read-Only

- `organization_id` (Number) Organization the group belongs to, as its Better Stack organization id. Defaults to the first organization the token can access. An organization the token can't access returns a 422 error that lists the ones it can. Not returned in responses.
- `name` (String) Group name.
- `sort_index` (Number) Display order among the organization's status page groups. Writes also accept it as position.
- `created_at` (String (RFC 3339))
- `updated_at` (String (RFC 3339))

## Endpoints

- `GET /api/v2/status-page-groups`
- `GET /api/v2/status-page-groups/:id`
- `POST /api/v2/status-page-groups`
- `PATCH /api/v2/status-page-groups/:id`
- `DELETE /api/v2/status-page-groups/:id`
