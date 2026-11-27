---
page_title: "betterstack_escalation_policy_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy group, a container that groups related escalation policies.
---

# betterstack_escalation_policy_group (Data Source)

An escalation policy group, a container that groups related escalation policies.

## Example Usage

```terraform
data "betterstack_escalation_policy_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the escalation policy group to look up.

### Optional

- `team_id` (String) Team that owns this group, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Group name.
- `sort_index` (Number) The group's position in your list of escalation policy groups, lower values first. `null` when no position was set.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.
- `team_name` (String) Name of the team that owns this group.
