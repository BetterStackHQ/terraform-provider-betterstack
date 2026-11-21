---
page_title: "betterstack_status_page_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A group of related status pages. It belongs to the organization, not to a team.
---

# betterstack_status_page_group (Resource)

A group of related status pages. It belongs to the organization, not to a team.

## Example Usage

```terraform
resource "betterstack_status_page_group" "example" {
  name = "Production Status Pages"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `organization_id` (String) Organization the group belongs to, as its Better Stack organization id. Defaults to the first organization the token can access. An organization the token can't access returns a 422 error that lists the ones it can. Not returned in responses. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `sort_index` (Number) Display order among the organization's status page groups. Writes also accept it as position. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this status_page_group.
- `created_at` (String (RFC 3339))
- `updated_at` (String (RFC 3339))

## Endpoints

- `GET /api/v2/status-page-groups`
- `GET /api/v2/status-page-groups/:id`
- `POST /api/v2/status-page-groups`
- `PATCH /api/v2/status-page-groups/:id`
- `DELETE /api/v2/status-page-groups/:id`

## Import

```shell
terraform import betterstack_status_page_group.example 12345
```
