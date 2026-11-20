---
page_title: "betterstack_role Data Source - terraform-provider-betterstack"
subcategory: "Better Stack"
description: |-
  A role in the organization: a system role, or a custom role built in the role editor.
---

# betterstack_role (Data Source)

A role in the organization: a system role, or a custom role built in the role editor.

-> The API describes no create/delete for this object, so the data source is its only Terraform surface.

## Example Usage

```terraform
data "betterstack_role" "example" {
  name = "On-call lead"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Role name.

### Read-Only

- `role` (String) The built-in role (admin, billing_admin, team_lead, responder, member), or "custom" for a role the organization made itself. A team member's `role` takes the same values.

## Endpoints

- `GET /api/v2/roles`
