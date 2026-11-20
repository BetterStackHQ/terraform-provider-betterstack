---
page_title: "betterstack_application_group Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  A group that organizes related error-tracking applications together.
---

# betterstack_application_group (Resource)

A group that organizes related error-tracking applications together.

-> Replaces `logtail_errors_application_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_application_group" "example" {
  name = "production-apps"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `sort_index` (Number) Display order among the team's application groups. `null` when never set. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this application_group.
- `team_name` (String) Name of the team that owns the group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v1/application-groups`
- `GET /api/v1/application-groups/:id`
- `POST /api/v1/application-groups`
- `PATCH /api/v1/application-groups/:id`
- `DELETE /api/v1/application-groups/:id`

## Import

```shell
terraform import betterstack_application_group.example 12345
```
