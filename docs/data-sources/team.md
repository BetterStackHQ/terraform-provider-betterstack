---
page_title: "betterstack_team Data Source - terraform-provider-betterstack"
subcategory: "Better Stack"
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

- `id` (String) The id of the team to look up.

### Read-Only

- `name` (String) The team's name. Unique within the organization, ignoring letter case.
- `incident_grouping_period` (Number) Seconds within which a new incident is grouped with an open one, from 0 to 2147483647. New teams start at 300.
- `incident_grouping_metadata_scope` (List of String) Metadata keys whose values must match for incidents to be grouped. New teams start with ["Origin"].
- `notify_about_all_incidents_in_a_group` (Boolean) Whether every incident in a group alerts, rather than only the first. New teams start at true.
- `reporting_enabled` (Boolean) Whether reporting is on. New teams start at false.
- `opened_emails_tracking_enabled` (Boolean) Whether the incident timeline records opened alert e-mails. New teams start at true.
- `on_call_reminder_lead_time` (Number) Seconds before an on-call shift starts to remind the person: one of 0 (off), 3600, 21600, 43200, 86400, 172800 or 604800. A value other than 0 needs a plan with on-call scheduling. New teams start at 86400.
- `on_call_reminder_sms` (Boolean) Send on-call reminders by SMS. Turning it on needs a plan with on-call scheduling and SMS alerts. New teams start at false.
- `on_call_reminder_email` (Boolean) Send on-call reminders by e-mail. Turning it on needs a plan with on-call scheduling. New teams start at true.
- `on_call_reminder_push` (Boolean) Send on-call reminders as push notifications. Turning it on needs a plan with on-call scheduling and push alerts. New teams start at true.
- `created_at` (String (RFC 3339)) When the team was created.
- `updated_at` (String (RFC 3339)) When the team was last updated.

## Endpoints

- `GET /api/v2/teams`
- `GET /api/v2/teams/:id`
- `POST /api/v2/teams`
- `PATCH /api/v2/teams/:id`
- `DELETE /api/v2/teams/:id`
