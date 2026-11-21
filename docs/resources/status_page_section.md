---
page_title: "betterstack_status_page_section Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A section that groups resources on a status page.
---

# betterstack_status_page_section (Resource)

A section that groups resources on a status page.

-> Replaces `betteruptime_status_page_section`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_status_page_section" "example" {
  status_page_id = "12345"
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `status_page_id` (String, Required) Id of the parent this status page belongs to. Changing it replaces the resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Section name. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `position` (Number) 0-based display order on the page. Inserting at a position shifts the sections at or after it down one. With fixed_position true, which Terraform always sends, the position is written as given and nothing shifts. A position past the end lands last. Left out, the new one is added at the end. Must be a non-negative integer. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.

### Read-Only

- `id` (String) The id of this status_page_section.

## Endpoints

- `GET /api/v2/status-pages/:status_page_id/sections`
- `GET /api/v2/status-pages/:status_page_id/sections/:id`
- `POST /api/v2/status-pages/:status_page_id/sections`
- `PATCH /api/v2/status-pages/:status_page_id/sections/:id`
- `DELETE /api/v2/status-pages/:status_page_id/sections/:id`

## Import

```shell
terraform import betterstack_status_page_section.example <status_page_id>/12345
```
