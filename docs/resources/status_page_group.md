---
page_title: "betterstack_status_page_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group of related status pages. It belongs to the whole organization.
---

# betterstack_status_page_group (Resource)

A group of related status pages. It belongs to the whole organization.

## Example Usage

### Create a status page group

```terraform
resource "betterstack_status_page_group" "example" {
  name       = "Production Status Pages"
  sort_index = 1
}
```

### In a chosen organization

```terraform
resource "betterstack_status_page_group" "in_a_chosen_organization" {
  name            = "Customer status pages"
  organization_id = "123456"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `organization_id` (String, Forces replacement) Organization the group belongs to, such as `123`. Required when you sign in with OAuth and can access more than one organization, otherwise the token's own organization. Not returned in responses. Better Stack never returns it, so Terraform keeps the value you set.
- `sort_index` (Number) Display order among the organization's status page groups. Lower values come first. Writes also accept it as `position`.

### Read-Only

- `id` (String) The ID of this status page group.
- `created_at` (String (RFC 3339)) When the status page group was created.
- `updated_at` (String (RFC 3339)) When the status page group was last updated.

## Import

Import an existing status page group by its ID. Get the ID via [List status page groups](https://betterstack.com/docs/uptime/api/list-all-status-page-groups/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_status_page_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_status_page_group.example 12345
```
