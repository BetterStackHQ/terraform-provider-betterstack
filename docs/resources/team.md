---
page_title: "betterstack_team Resource - terraform-provider-betterstack"
subcategory: "Account"
description: |-
  A team in the organization. Everything you create in any Better Stack product belongs to a team.
---

# betterstack_team (Resource)

A team in the organization. Everything you create in any Better Stack product belongs to a team.

~> Destroying a team deletes its data for good, so `deletion_protection` is on by default: Terraform refuses to destroy or replace it until you set `deletion_protection = false` and apply.

## Example Usage

### Create a team

```terraform
resource "betterstack_team" "example" {
  name                             = "Platform team"
  incident_grouping_period         = 600
  incident_grouping_metadata_scope = ["environment"]
}
```

### Remind on-call people an hour ahead by push only

```terraform
resource "betterstack_team" "remind_on_call_people_an_hour_ahead_by_push_only" {
  name                       = "Support team"
  on_call_reminder_lead_time = 3600
  on_call_reminder_email     = false
  on_call_reminder_push      = true
}
```

### Alert only for the first incident in a group

```terraform
resource "betterstack_team" "alert_only_for_the_first_incident_in_a_group" {
  name                                  = "Payments team"
  incident_grouping_period              = 1800
  notify_about_all_incidents_in_a_group = false
}
```

## Schema

### Required

- `name` (String) The team's name. It must be unique in the organization, ignoring letter case.

### Optional

- `deletion_protection` (Boolean) Whether destroying or replacing this team is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.
- `incident_grouping_period` (Number) Seconds within which a new incident is grouped with an earlier one that's still open, for example `300` for 5 minutes. `0` never groups incidents. From `0` to `2147483647`, and new teams start at `300`. Defaults to `300`.
- `incident_grouping_metadata_scope` (List of String) Metadata keys whose values must match for incidents to be grouped. New teams start with `["Origin"]`. `null` groups the same way as `["Origin"]`. Defaults to `["Origin"]`.
- `notify_about_all_incidents_in_a_group` (Boolean) Whether every incident in a group alerts, rather than only the first. New teams start at true. Defaults to `true`.
- `reporting_enabled` (Boolean) Whether reporting is on, which unlocks the reports in Incidents. New teams start at `false`. Defaults to `false`.
- `opened_emails_tracking_enabled` (Boolean) Whether the incident timeline records when a team member opens the e-mail about a new incident. New teams start at `true`. Defaults to `true`.
- `on_call_reminder_lead_time` (Number, Computed) Seconds before an on-call shift starts to remind the person, or `0` for no reminder. The other values are `3600`, `21600`, `43200`, `86400`, `172800` and `604800`. New teams start at `86400`.
- `on_call_reminder_sms` (Boolean) Send on-call reminders by SMS. New teams start at `false`. Defaults to `false`.
- `on_call_reminder_email` (Boolean, Computed) Send on-call reminders by e-mail. New teams start at `true`.
- `on_call_reminder_push` (Boolean, Computed) Send on-call reminders as push notifications. New teams start at `true`.
- `logo_remote_url` (String) The URL the team's logo was fetched from. It accepts an image URL, and we fetch the logo from it. Remove the argument to remove the logo.

### Read-Only

- `id` (String) The ID of this team.
- `organization_id` (String) ID of the organization the team belongs to.
- `created_at` (String (RFC 3339)) When the team was created.
- `updated_at` (String (RFC 3339)) When the team was last updated.

-> The `logo_url` and `logo_remote_url` attributes must be absolute URLs, or the change is refused. A URL that redirects isn't followed.

-> A team's `id` is a string with a `t` prefix, such as `"t123"`, the same form the other Better Stack APIs use for `team_id`.

-> Team settings are checked, then applied in Incidents before the change is saved, so a settings change Incidents refuses changes nothing. When Incidents refuses the settings of a new team, the team is removed again.

## Import

Import an existing team by its ID. Get the ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_team.example
  id = "t123"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_team.example t123
```
