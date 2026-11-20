---
page_title: "betterstack_team Resource - terraform-provider-betterstack"
subcategory: "Better Stack"
description: |-
  A team in the organization. Everything you create in any Better Stack product belongs to a team.
---

# betterstack_team (Resource)

A team in the organization. Everything you create in any Better Stack product belongs to a team.

## Example Usage

```terraform
resource "betterstack_team" "example" {
  name = "Platform team"
}
```

## Schema

### Required

- `name` (String) The team's name. Unique within the organization, ignoring letter case.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `incident_grouping_period` (Number) Seconds within which a new incident is grouped with an open one, from 0 to 2147483647. New teams start at 300.
- `incident_grouping_metadata_scope` (List of String) Metadata keys whose values must match for incidents to be grouped. New teams start with ["Origin"]. Defaults to `["Origin"]`.
- `notify_about_all_incidents_in_a_group` (Boolean) Whether every incident in a group alerts, rather than only the first. New teams start at true.
- `reporting_enabled` (Boolean) Whether reporting is on. New teams start at false.
- `opened_emails_tracking_enabled` (Boolean) Whether the incident timeline records opened alert e-mails. New teams start at true.
- `on_call_reminder_lead_time` (Number) Seconds before an on-call shift starts to remind the person: one of 0 (off), 3600, 21600, 43200, 86400, 172800 or 604800. A value other than 0 needs a plan with on-call scheduling. New teams start at 86400.
- `on_call_reminder_sms` (Boolean) Send on-call reminders by SMS. Turning it on needs a plan with on-call scheduling and SMS alerts. New teams start at false.
- `on_call_reminder_email` (Boolean) Send on-call reminders by e-mail. Turning it on needs a plan with on-call scheduling. New teams start at true.
- `on_call_reminder_push` (Boolean) Send on-call reminders as push notifications. Turning it on needs a plan with on-call scheduling and push alerts. New teams start at true.

### Read-Only

- `id` (String) The id of this team.
- `created_at` (String (RFC 3339)) When the team was created.
- `updated_at` (String (RFC 3339)) When the team was last updated.

## Endpoints

- `GET /api/v2/teams`
- `GET /api/v2/teams/:id`
- `POST /api/v2/teams`
- `PATCH /api/v2/teams/:id`
- `DELETE /api/v2/teams/:id`

-> A team's `id` is a string with a `t` prefix ("t123"), the same form the other Better Stack APIs use for `team_id`. Every path also accepts the bare number (123).

-> Team settings are checked, then applied in Incidents before the change is saved, so a settings change Incidents refuses changes nothing. When Incidents refuses the settings of a new team, the team is removed again.

## Import

```shell
terraform import betterstack_team.example 12345
```
