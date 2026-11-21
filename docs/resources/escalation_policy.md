---
page_title: "betterstack_escalation_policy Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy: the ordered steps and people notified when an incident fires.
---

# betterstack_escalation_policy (Resource)

An escalation policy: the ordered steps and people notified when an incident fires.

## Example Usage

```terraform
resource "betterstack_escalation_policy" "example" {
  name = "Standard Escalation Policy"
}
```

## Schema

### Required

- `name` (String) Policy name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `repeat_count` (Number) How many times to repeat the steps if the incident stays unacknowledged. Defaults to `0`.
- `repeat_delay` (Number) Seconds to wait before repeating the steps. Defaults to `0`.
- `escalation_policy_group_id` (String) Escalation policy group this policy belongs to. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `fallback_policy_id` (String) Escalation policy to escalate to after these steps and repeats run out. Must belong to the same organization. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `steps` (Attributes List, see [nested schema](#nestedatt--steps)) Ordered escalation steps. Sending steps replaces the whole list. An empty list makes a silent policy that only collects incidents. Every step has a type, a wait set by wait_before or by wait_until_time and wait_until_timezone, and the keys its type uses. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.

### Read-Only

- `id` (String) The id of this escalation_policy.
- `team_name` (String) Name of the team that owns this policy.
- `incident_token` (String, Sensitive) Token used to open incidents against this policy. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.

<a id="nestedatt--steps"></a>
### Nested Schema for `steps`

- `type` (String) Step type. Valid values: `escalation`, `time_branching`, `metadata_branching`, `instructions`, `resolve_remove`.
- `wait_before` (Number) Seconds to wait before this step runs.
- `wait_until_time` (String) Time of day to wait until, as `HH:MM`, instead of wait_before.
- `wait_until_timezone` (String) Timezone for wait_until_time.
- `wait_until_days` (List of String) Weekdays wait_until_time applies to.
- `severity_id` (String) Severity to alert with. Used by escalation steps.
- `step_members` (Attributes List, see [nested schema](#nestedatt--steps--step_members)) Who to notify. Each member has a type, plus an id, email or metadata_key depending on the type. The notes below list every member type.
- `timezone` (String) Timezone the time branch condition is evaluated in.
- `days` (List of String) Weekdays the time branch applies to, from `mon` to `sun`.
- `time_from` (String) Start of the time branch window, as `HH:MM`.
- `time_to` (String) End of the time branch window, as `HH:MM`.
- `metadata_key` (String) Incident metadata key that a metadata branching step checks.
- `metadata_values` (Attributes List, see [nested schema](#nestedatt--steps--metadata_values)) Metadata values to match. Each is `{type: "String", value}`, `{type: "no_value"}` or a catalog item reference. You can also send a value as a plain string. See the notes below.
- `action_type` (String) The action of a time_branching or metadata_branching step, or of a resolve_remove step. A branch without escalation_policy_id or policy_metadata_key is saved as `do_not_escalate`. Valid values: `escalate_to_policy`, `resolve_incident`, `remove_incident`, `do_not_escalate`.
- `escalation_policy_id` (String) Escalation policy the branch escalates to when action_type is `escalate_to_policy`.
- `policy_metadata_key` (String) Incident metadata key whose value names the escalation policy to escalate to, instead of a fixed escalation_policy_id.
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

- `GET /api/v5/escalation-policies`
- `GET /api/v5/escalation-policies/:id`
- `POST /api/v5/escalation-policies`
- `PATCH /api/v5/escalation-policies/:id`
- `DELETE /api/v5/escalation-policies/:id`

-> A step's type decides which keys it uses. An escalation step uses severity_id and step_members, with at least one member. A time_branching step uses timezone, days, time_from and time_to, plus action_type, escalation_policy_id and policy_metadata_key. A metadata_branching step uses metadata_key and metadata_values, plus the same action_type, escalation_policy_id and policy_metadata_key. An instructions step uses instructions_comment, instructions_reminder_enabled and instructions_reminder_interval_hours. A resolve_remove step uses action_type.

-> Each step member has a type: current_on_call, entire_team, user, policy or incident_metadata. A member can also be one integration: webhook, email_integration, slack_integration, linear_integration, jira_integration, halo_psa_integration, microsoft_teams_integration, pagerduty_integration, zapier_webhook, zoom_integration, google_meet_integration, ai_sre or ai_sre_decision. Or it can be every integration of one kind: all_slack_integrations, all_linear_integrations, all_jira_integrations, all_microsoft_teams_integrations, all_zapier_integrations, all_pagerduty_integrations, all_webhook_integrations, all_email_integrations, all_splunk_on_call_integrations, all_incident_io_integrations, all_zoom_integrations or all_google_meet_integrations. A current_on_call member takes the id of an on-call calendar, or no id for the team's own. An entire_team member takes the id of another team in the organization, returned as `t123`, or no id for the policy's own team. A user member takes an id or an email. An incident_metadata member takes a metadata_key. Every other member type takes the id of its target. team_id is an older way to send a team or calendar's team id. It is accepted on write and never returned.

-> metadata_values entries are returned as `{type: "String", value}`, as `{type: "no_value"}`, or as a catalog item `{type, item_id, name}`, plus email for a User. `no_value` matches when the incident's metadata key is missing or empty. A catalog item's type is one of User, Team, Policy, Schedule, Endpoint, SlackIntegration, LinearIntegration, JiraIntegration, MicrosoftTeamsWebhook, ZapierWebhook, NativeWebhook, PagerDutyWebhook or Source. A team's item_id starts with `t`, such as `t123`. When writing, an entry can also be a plain string, and a catalog item is found by its item_id, name or email.

## Import

```shell
terraform import betterstack_escalation_policy.example 12345
```
