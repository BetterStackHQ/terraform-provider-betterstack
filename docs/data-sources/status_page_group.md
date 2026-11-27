---
page_title: "betterstack_status_page_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group of related status pages. It belongs to the whole organization.
---

# betterstack_status_page_group (Data Source)

A group of related status pages. It belongs to the whole organization.

## Example Usage

```terraform
data "betterstack_status_page_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the status page group to look up.

### Read-Only

- `organization_id` (String) Organization the group belongs to, such as `123`. Required when you sign in with OAuth and can access more than one organization, otherwise the token's own organization. Not returned in responses.
- `name` (String) Group name.
- `sort_index` (Number) Display order among the organization's status page groups. Lower values come first. Writes also accept it as `position`.
- `created_at` (String (RFC 3339)) When the status page group was created.
- `updated_at` (String (RFC 3339)) When the status page group was last updated.
