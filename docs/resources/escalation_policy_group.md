---
page_title: "betterstack_escalation_policy_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy group: a container that groups related escalation policies.
---

# betterstack_escalation_policy_group (Resource)

An escalation policy group: a container that groups related escalation policies.

-> Replaces `betteruptime_policy_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_escalation_policy_group" "example" {
  name = "Example"
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `sort_index` (Number) Display order of the group among the team's escalation policy groups. When writing, you can also send it as position. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this escalation_policy_group.
- `team_name` (String) Name of the team that owns this group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v2/policy-groups`
- `GET /api/v2/policy-groups/:id`
- `POST /api/v2/policy-groups`
- `PATCH /api/v2/policy-groups/:id`
- `DELETE /api/v2/policy-groups/:id`

## Import

```shell
terraform import betterstack_escalation_policy_group.example 12345
```
