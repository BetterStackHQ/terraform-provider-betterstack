---
page_title: "betterstack_escalation_policy Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy sets the ordered steps and people notified when an incident fires.
---

# betterstack_escalation_policy (Resource)

An escalation policy sets the ordered steps and people notified when an incident fires.

## Example Usage

### Page the on-call person, then the whole team after 10 minutes

```terraform
data "betterstack_severity" "high" {
  name = "High severity"
}

resource "betterstack_escalation_policy" "on_call" {
  name = "On-call, then the whole team"
  steps = [
    { type = "escalation", wait_before = 0, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "current_on_call" }] },
    { type = "escalation", wait_before = 600, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "entire_team" }] },
  ]
}
```

### Post to Slack right away, page the on-call person after 3 minutes and repeat

```terraform
data "betterstack_severity" "high" {
  name = "High severity"
}

resource "betterstack_escalation_policy" "slack_first" {
  name         = "Slack, then on-call"
  repeat_count = 3
  repeat_delay = 300
  steps = [
    { type = "escalation", wait_before = 0, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "all_slack_integrations" }] },
    { type = "escalation", wait_before = 180, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "current_on_call" }] },
  ]
}
```

## Schema

### Required

- `name` (String) Policy name.

### Optional

- `repeat_count` (Number) How many times the whole policy repeats if nobody acknowledges the incident. Defaults to `0`.
- `repeat_delay` (Number) Seconds until the steps repeat. Defaults to `0`.
- `escalation_policy_group_id` (String) The ID of the escalation policy group this policy is in, or `null` when it's in none. Remove the argument to clear it from its group.
- `fallback_escalation_policy_id` (String) Escalation policy to escalate to after these steps and repeats run out. Must belong to the same organization.
- `steps` (Attributes List, Computed) Ordered escalation steps. An update with steps replaces the whole list. An empty list makes a silent policy that only collects incidents. Every step has a type, a wait set by `wait_before` or by `wait_until_time` and `wait_until_timezone`, and the keys its type uses. (see [below for nested schema](#nestedatt--steps))
- `team_id` (String) Team that owns this policy, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this escalation policy.
- `incident_token` (String, Sensitive) The secret token in the link to the policy's incident report form, used to open incidents on this policy. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `team_name` (String) Name of the team that owns this policy.

<a id="nestedatt--steps"></a>
### Nested Schema for `steps`

Required:

- `type` (String) The step type.
  - `escalation`: Notifies people or integrations.
  - `time_branching`: Acts when the incident happens in a time window, and otherwise continues with the next step.
  - `metadata_branching`: Acts when the incident's metadata matches a value, and otherwise continues with the next step.
  - `instructions`: Posts instructions into the incident timeline.
  - `resolve_remove`: Resolves or removes the incident.

Optional:

- `wait_before` (Number) How long the step waits after the previous step, in seconds, from `0` to `691200`, which is 8 days. Defaults to `0`. Leave it out when the step waits until a time of day with `wait_until_time`.
- `wait_until_time` (String) The time of day the step runs, in `HH:MM` format, instead of waiting `wait_before`. Requires `wait_until_timezone`. Responses show it in place of `wait_before` on such steps.
- `wait_until_timezone` (String) The time zone of `wait_until_time`, as a Rails time zone name such as `London` or `Eastern Time (US & Canada)`.
- `wait_until_days` (List of String) The days on which `wait_until_time` applies, out of `mon`, `tue`, `wed`, `thu`, `fri`, `sat` and `sun`. Defaults to every day.
- `severity_id` (String) The ID of the severity that sets how the step's members are notified. Required on `escalation` steps. Get the severity ID via [List severities](https://betterstack.com/docs/uptime/api/list-all-severities/).
- `step_members` (Attributes List) Who or what the step notifies, at least one member. Required on `escalation` steps. Each member has a `type`, plus an `id`, `email` or `metadata_key` depending on the type. (see [below for nested schema](#nestedatt--steps--step_members))
- `timezone` (String) The time zone of the time window. Required on `time_branching` steps.
- `days` (List of String) The days on which the time window applies, out of `mon`, `tue`, `wed`, `thu`, `fri`, `sat` and `sun`. Used on `time_branching` steps and defaults to every day.
- `time_from` (String) Start of the time window, in `HH:MM` format. Required on `time_branching` steps.
- `time_to` (String) End of the time window, in `HH:MM` format. Required on `time_branching` steps.
- `metadata_key` (String) The incident metadata key the step checks. Required on `metadata_branching` steps.
- `metadata_values` (Attributes List) The metadata values that match, at least one. Required on `metadata_branching` steps. Each is `{type: "String", value}`, `{type: "no_value"}` or a catalog item reference. `no_value` matches when the key is missing or blank. (see [below for nested schema](#nestedatt--steps--metadata_values))
- `action_type` (String) What the step does. Required on `resolve_remove` steps, which accept only resolving or removing. On branching steps it applies when the incident matches. It defaults to running another policy when you set `escalation_policy_id` or `policy_metadata_key`, and to stopping otherwise.
  - `escalate_to_policy`: Runs the escalation policy in `escalation_policy_id` or `policy_metadata_key` and stops this policy.
  - `resolve_incident`: Resolves the incident and stops this policy.
  - `remove_incident`: Removes the incident for good and stops this policy.
  - `do_not_escalate`: Stops this policy.
- `escalation_policy_id` (String) The ID of another escalation policy in your organization to run when the incident matches. Used on `time_branching` and `metadata_branching` steps. Can't be combined with `policy_metadata_key`.
- `policy_metadata_key` (String) A key in the incident's metadata that holds the escalation policy to run when the incident matches, instead of `escalation_policy_id`. Used on `time_branching` and `metadata_branching` steps. If the key is missing or holds no policy, the policy stops.
- `instructions_comment` (String) The instructions posted as a comment into the incident timeline. Supports Markdown, checkboxes like `- [ ] Step 1`, and values from the incident's metadata as `{key}`, such as `{server_region}`. Required on `instructions` steps.
- `instructions_reminder_enabled` (Boolean) Whether to post a reminder periodically until every checkbox in the instructions is checked. Defaults to `false`.
- `instructions_reminder_interval_hours` (Number) How often the reminder is posted, in hours, on `instructions` steps with `instructions_reminder_enabled` on. At least `1`, and `4` by default.

<a id="nestedatt--steps--step_members"></a>
### Nested Schema for `steps.step_members`

Required:

- `type` (String) The member type. `current_on_call` notifies whoever is on call, and `entire_team` everyone in the team. `user` notifies one person, `policy` runs another escalation policy, `incident_metadata` notifies what the incident's metadata stores under `metadata_key`, and `ai_sre` starts an AI SRE investigation. A member can also be one integration, with type `webhook`, `email_integration`, `slack_integration`, `linear_integration`, `jira_integration`, `halo_psa_integration`, `microsoft_teams_integration`, `pagerduty_integration`, `zapier_webhook`, `zoom_integration` or `google_meet_integration`. Or it can be every integration of one kind, with type `all_slack_integrations`, `all_linear_integrations`, `all_jira_integrations`, `all_microsoft_teams_integrations`, `all_zapier_integrations`, `all_pagerduty_integrations`, `all_webhook_integrations`, `all_email_integrations`, `all_splunk_on_call_integrations`, `all_incident_io_integrations`, `all_zoom_integrations` or `all_google_meet_integrations`.

Optional:

- `id` (String) The ID of the member's target. Required for `user` unless `email` is set, and for `policy`, `ai_sre` and every type that notifies one integration. A `user` member is `u123` or their e-mail, and responses return `u123`. A `current_on_call` member accepts an on-call calendar to notify instead of the team's default. An `entire_team` member accepts another team in the organization, as its ID such as `t123`. Responses return it the same way. Both come back without an ID when they use the policy's own team. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `email` (String) The e-mail of a `user` member, instead of its id.
- `metadata_key` (String) The incident metadata key whose stored users, teams, on-call calendars, escalation policies and integrations an `incident_metadata` member notifies.

<a id="nestedatt--steps--metadata_values"></a>
### Nested Schema for `steps.metadata_values`

Optional:

- `type` (String) `String`, `no_value`, or a catalog item's type, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`. `no_value` matches when the incident's metadata doesn't have the key or the key is empty.
- `value` (String) The text of a `String` value.
- `item_id` (String) The catalog item's ID. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) The catalog item's name. An item can also be found by its name instead of `item_id`.
- `email` (String) A User item's e-mail, returned for a User. A user can also be found by e-mail.

-> A step's type decides which keys it uses. An `escalation` step uses `severity_id` and `step_members`, with at least one member. A `time_branching` step uses `timezone`, `days`, `time_from` and `time_to`, plus `action_type`, `escalation_policy_id` and `policy_metadata_key`. A `metadata_branching` step uses `metadata_key` and `metadata_values`, plus the same `action_type`, `escalation_policy_id` and `policy_metadata_key`. An `instructions` step uses `instructions_comment`, `instructions_reminder_enabled` and `instructions_reminder_interval_hours`. A `resolve_remove` step uses `action_type`.

-> Each step member has a type, which is `current_on_call`, `entire_team`, `user`, `policy` or `incident_metadata`. A member can also be one integration, with type `webhook`, `email_integration`, `slack_integration`, `linear_integration`, `jira_integration`, `halo_psa_integration`, `microsoft_teams_integration`, `pagerduty_integration`, `zapier_webhook`, `zoom_integration`, `google_meet_integration` or `ai_sre`. Or it can be every integration of one kind, with type `all_slack_integrations`, `all_linear_integrations`, `all_jira_integrations`, `all_microsoft_teams_integrations`, `all_zapier_integrations`, `all_pagerduty_integrations`, `all_webhook_integrations`, `all_email_integrations`, `all_splunk_on_call_integrations`, `all_incident_io_integrations`, `all_zoom_integrations` or `all_google_meet_integrations`. A `current_on_call` member accepts the id of an on-call calendar, or no id for the team's own. An `entire_team` member accepts the id of another team in the organization, returned as `t123`, or no id for the policy's own team. A `user` member accepts an id, `u123` or the user's e-mail, or an `email`, and is returned as `u123`. An `incident_metadata` member needs a `metadata_key`. Every other member type needs the id of its target. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).

-> The `metadata_values` entries are returned as `{type: "String", value}`, as `{type: "no_value"}`, or as a catalog item `{type, item_id, name}`, plus `email` for a User. `no_value` matches when the incident's metadata doesn't have the key or the key is empty. A catalog item's type is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`. A user's `item_id` is `u123` and a team's is `t123`. A catalog item is found by its `item_id`, `name` or `email`.

## Import

Import an existing escalation policy by its ID. Get the ID via [List escalation policies](https://betterstack.com/docs/uptime/api/list-all-escalation-policies/). The [`betterstack_escalation_policy` data source](../data-sources/escalation_policy.md) also finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_escalation_policy.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_escalation_policy.example 12345
```
