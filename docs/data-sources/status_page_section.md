---
page_title: "betterstack_status_page_section Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A section that groups resources on a status page.
---

# betterstack_status_page_section (Data Source)

A section that groups resources on a status page.

## Example Usage

```terraform
data "betterstack_status_page_section" "example" {
  status_page_id = "12345"
  id             = "12345"
}
```

## Schema

### Address

- `status_page_id` (String, Required) Id of the parent this status page belongs to. Changing it replaces the resource.

### Required

- `id` (String) The id of the status_page_section to look up.

### Read-Only

- `name` (String) Section name.
- `position` (Number) 0-based display order on the page. Inserting at a position shifts the sections at or after it down one. With fixed_position true, which Terraform always sends, the position is written as given and nothing shifts. A position past the end lands last. Left out, the new one is added at the end. Must be a non-negative integer.

## Endpoints

- `GET /api/v2/status-pages/:status_page_id/sections`
- `GET /api/v2/status-pages/:status_page_id/sections/:id`
- `POST /api/v2/status-pages/:status_page_id/sections`
- `PATCH /api/v2/status-pages/:status_page_id/sections/:id`
- `DELETE /api/v2/status-pages/:status_page_id/sections/:id`
