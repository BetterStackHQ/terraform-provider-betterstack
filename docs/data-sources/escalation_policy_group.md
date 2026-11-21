---
page_title: "betterstack_escalation_policy_group Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy group: a container that groups related escalation policies.
---

# betterstack_escalation_policy_group (Data Source)

An escalation policy group: a container that groups related escalation policies.

## Example Usage

```terraform
data "betterstack_escalation_policy_group" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the escalation_policy_group to look up.

### Read-Only

- `name` (String) Group name.
- `team_name` (String) Name of the team that owns this group.
- `sort_index` (Number) Display order of the group among the team's escalation policy groups. When writing, you can also send it as position.
- `created_at` (String (RFC 3339)) When the group was created.
- `updated_at` (String (RFC 3339)) When the group was last updated.

## Endpoints

- `GET /api/v3/escalation-policy-groups`
- `GET /api/v3/escalation-policy-groups/:id`
- `POST /api/v3/escalation-policy-groups`
- `PATCH /api/v3/escalation-policy-groups/:id`
- `DELETE /api/v3/escalation-policy-groups/:id`
