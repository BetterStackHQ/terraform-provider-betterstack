---
page_title: "betterstack_escalation_policy Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy: the ordered steps and people notified when an incident fires.
---

# betterstack_escalation_policy (Data Source)

An escalation policy: the ordered steps and people notified when an incident fires.

## Example Usage

```terraform
data "betterstack_escalation_policy" "example" {
  name = "Standard Escalation Policy"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Policy name.

### Read-Only

- `team_name` (String) Name of the team that owns this policy.
- `repeat_count` (Number) How many times to repeat the steps if the incident stays unacknowledged.
- `repeat_delay` (Number) Seconds to wait before repeating the steps.
- `policy_group_id` (Number) Escalation policy group this policy belongs to.
- `fallback_policy_id` (Number) Escalation policy to escalate to after these steps and repeats run out. Must belong to the same organization.
- `incident_token` (String, Sensitive) Token used to open incidents against this policy.
- `steps` (Attributes List, see [nested schema](#nestedatt--steps)) Ordered escalation steps. Sending steps replaces the whole list. An empty list makes a silent policy that only collects incidents. Every step has a type, a wait set by wait_before or by wait_until_time and wait_until_timezone, and the keys its type uses.

<a id="nestedatt--steps"></a>
### Nested Schema for `steps`

- `type` (String) Step type. Valid values: `escalation`, `time_branching`, `metadata_branching`, `instructions`, `resolve_remove`.
- `wait_before` (Number) Seconds to wait before this step runs.
- `wait_until_time` (String) Time of day to wait until, as `HH:MM`, instead of wait_before.
- `wait_until_timezone` (String) Timezone for wait_until_time.
- `wait_until_days` (List of String) Weekdays wait_until_time applies to.
- `urgency_id` (Number) Severity to alert with. Used by escalation steps.
- `step_members` (Attributes List, see [nested schema](#nestedatt--steps--step_members)) Who to notify. Each member has a type, plus an id, email or metadata_key depending on the type. The notes below list every member type.
- `timezone` (String) Timezone the time branch condition is evaluated in.
- `days` (List of String) Weekdays the time branch applies to, from `mon` to `sun`.
- `time_from` (String) Start of the time branch window, as `HH:MM`.
- `time_to` (String) End of the time branch window, as `HH:MM`.
- `metadata_key` (String) Incident metadata key that a metadata branching step checks.
- `metadata_values` (Attributes List, see [nested schema](#nestedatt--steps--metadata_values)) Metadata values to match. Each is `{type: "String", value}`, `{type: "no_value"}` or a catalog item reference. You can also send a value as a plain string. See the notes below.
- `action_type` (String) The action of a time_branching or metadata_branching step, or of a resolve_remove step. A branch without policy_id or policy_metadata_key is saved as `do_not_escalate`. Valid values: `escalate_to_policy`, `resolve_incident`, `remove_incident`, `do_not_escalate`.
- `policy_id` (Number) Escalation policy the branch escalates to when action_type is `escalate_to_policy`.
- `policy_metadata_key` (String) Incident metadata key whose value names the escalation policy to escalate to, instead of a fixed policy_id.
- `instructions_comment` (String) Text of an instructions step.
- `instructions_reminder_enabled` (Boolean) Whether to remind about unacknowledged instructions.
- `instructions_reminder_interval_hours` (Number) Reminder interval in hours.

<a id="nestedatt--steps--step_members"></a>
### Nested Schema for `steps.step_members`

- `type` (String)
- `id` (String)
- `team_id` (String)
- `email` (String)
- `metadata_key` (String)

<a id="nestedatt--steps--metadata_values"></a>
### Nested Schema for `steps.metadata_values`

- `type` (String)
- `value` (String)
- `item_id` (String)
- `name` (String)
- `email` (String)

## Endpoints

- `GET /api/v4/policies`
- `GET /api/v4/policies/:id`
- `POST /api/v4/policies`
- `PATCH /api/v4/policies/:id`
- `DELETE /api/v4/policies/:id`
