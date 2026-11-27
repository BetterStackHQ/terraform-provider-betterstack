---
page_title: "betterstack_escalation_policy Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An escalation policy sets the ordered steps and people notified when an incident fires.
---

# betterstack_escalation_policy (Data Source)

An escalation policy sets the ordered steps and people notified when an incident fires.

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

### Optional

- `team_id` (String) Team that owns this policy, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `repeat_count` (Number) How many times the whole policy repeats if nobody acknowledges the incident.
- `repeat_delay` (Number) Seconds until the steps repeat.
- `escalation_policy_group_id` (String) The ID of the escalation policy group this policy is in, or `null` when it's in none.
- `fallback_escalation_policy_id` (String) Escalation policy to escalate to after these steps and repeats run out. Must belong to the same organization.
- `incident_token` (String, Sensitive) The secret token in the link to the policy's incident report form, used to open incidents on this policy.
- `steps` (Attributes List) Ordered escalation steps. An update with steps replaces the whole list. An empty list makes a silent policy that only collects incidents. Every step has a type, a wait set by `wait_before` or by `wait_until_time` and `wait_until_timezone`, and the keys its type uses. (see [below for nested schema](#nestedatt--steps))
- `team_name` (String) Name of the team that owns this policy.

<a id="nestedatt--steps"></a>
### Nested Schema for `steps`

Read-Only:

- `type` (String) The step type.
  - `escalation`: Notifies people or integrations.
  - `time_branching`: Acts when the incident happens in a time window, and otherwise continues with the next step.
  - `metadata_branching`: Acts when the incident's metadata matches a value, and otherwise continues with the next step.
  - `instructions`: Posts instructions into the incident timeline.
  - `resolve_remove`: Resolves or removes the incident.
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

Read-Only:

- `type` (String) The member type. `current_on_call` notifies whoever is on call, and `entire_team` everyone in the team. `user` notifies one person, `policy` runs another escalation policy, `incident_metadata` notifies what the incident's metadata stores under `metadata_key`, and `ai_sre` starts an AI SRE investigation. A member can also be one integration, with type `webhook`, `email_integration`, `slack_integration`, `linear_integration`, `jira_integration`, `halo_psa_integration`, `microsoft_teams_integration`, `pagerduty_integration`, `zapier_webhook`, `zoom_integration` or `google_meet_integration`. Or it can be every integration of one kind, with type `all_slack_integrations`, `all_linear_integrations`, `all_jira_integrations`, `all_microsoft_teams_integrations`, `all_zapier_integrations`, `all_pagerduty_integrations`, `all_webhook_integrations`, `all_email_integrations`, `all_splunk_on_call_integrations`, `all_incident_io_integrations`, `all_zoom_integrations` or `all_google_meet_integrations`.
- `id` (String) The ID of the member's target. Required for `user` unless `email` is set, and for `policy`, `ai_sre` and every type that notifies one integration. A `user` member is `u123` or their e-mail, and responses return `u123`. A `current_on_call` member accepts an on-call calendar to notify instead of the team's default. An `entire_team` member accepts another team in the organization, as its ID such as `t123`. Responses return it the same way. Both come back without an ID when they use the policy's own team. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `email` (String) The e-mail of a `user` member, instead of its id.
- `metadata_key` (String) The incident metadata key whose stored users, teams, on-call calendars, escalation policies and integrations an `incident_metadata` member notifies.

<a id="nestedatt--steps--metadata_values"></a>
### Nested Schema for `steps.metadata_values`

Read-Only:

- `type` (String) `String`, `no_value`, or a catalog item's type, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`. `no_value` matches when the incident's metadata doesn't have the key or the key is empty.
- `value` (String) The text of a `String` value.
- `item_id` (String) The catalog item's ID. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) The catalog item's name. An item can also be found by its name instead of `item_id`.
- `email` (String) A User item's e-mail, returned for a User. A user can also be found by e-mail.
