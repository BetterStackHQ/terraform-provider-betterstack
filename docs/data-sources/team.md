---
page_title: "betterstack_team Data Source - terraform-provider-betterstack"
subcategory: "Account"
description: |-
  A team in the organization. Everything you create in any Better Stack product belongs to a team.
---

# betterstack_team (Data Source)

A team in the organization. Everything you create in any Better Stack product belongs to a team.

## Example Usage

```terraform
data "betterstack_team" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the team to look up.

### Read-Only

- `name` (String) The team's name. It must be unique in the organization, ignoring letter case.
- `organization_id` (String) ID of the organization the team belongs to.
- `incident_grouping_period` (Number) Seconds within which a new incident is grouped with an earlier one that's still open, for example `300` for 5 minutes. `0` never groups incidents. From `0` to `2147483647`, and new teams start at `300`.
- `incident_grouping_metadata_scope` (List of String) Metadata keys whose values must match for incidents to be grouped. New teams start with `["Origin"]`. `null` groups the same way as `["Origin"]`.
- `notify_about_all_incidents_in_a_group` (Boolean) Whether every incident in a group alerts, rather than only the first. New teams start at true.
- `reporting_enabled` (Boolean) Whether reporting is on, which unlocks the reports in Incidents. New teams start at `false`.
- `opened_emails_tracking_enabled` (Boolean) Whether the incident timeline records when a team member opens the e-mail about a new incident. New teams start at `true`.
- `on_call_reminder_lead_time` (Number) Seconds before an on-call shift starts to remind the person, or `0` for no reminder. The other values are `3600`, `21600`, `43200`, `86400`, `172800` and `604800`. New teams start at `86400`.
- `on_call_reminder_sms` (Boolean) Send on-call reminders by SMS. New teams start at `false`.
- `on_call_reminder_email` (Boolean) Send on-call reminders by e-mail. New teams start at `true`.
- `on_call_reminder_push` (Boolean) Send on-call reminders as push notifications. New teams start at `true`.
- `logo_remote_url` (String) The URL the team's logo was fetched from.
- `created_at` (String (RFC 3339)) When the team was created.
- `updated_at` (String (RFC 3339)) When the team was last updated.
- `deletion_protection` (Boolean) Whether destroying or replacing this team is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.
