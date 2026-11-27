---
page_title: "betterstack_escalation_policy_group Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy group, a container that groups related escalation policies.
---

# betterstack_escalation_policy_group (Resource)

An escalation policy group, a container that groups related escalation policies.

-> Replaces `betteruptime_policy_group`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create an escalation policy group

```terraform
resource "betterstack_escalation_policy_group" "example" {
  name = "Platform on-call"
}
```

### Listed first

```terraform
resource "betterstack_escalation_policy_group" "listed_first" {
  name       = "Critical escalations"
  sort_index = 0
}
```

## Schema

### Required

- `name` (String) Group name.

### Optional

- `sort_index` (Number) The group's position in your list of escalation policy groups, lower values first. `null` when no position was set.
- `team_id` (String) Team that owns this group, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this escalation policy group.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this group.

## Import

Import an existing escalation policy group by its ID. Get the ID via [List escalation policy groups](https://betterstack.com/docs/uptime/api/list-all-existing-escalation-policy-groups/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_escalation_policy_group.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_escalation_policy_group.example 12345
```
