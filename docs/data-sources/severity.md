---
page_title: "betterstack_severity Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity: a named alerting level with its notification-channel settings.
---

# betterstack_severity (Data Source)

A severity: a named alerting level with its notification-channel settings.

## Example Usage

```terraform
data "betterstack_severity" "example" {
  name = "Low Severity"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Severity name.

### Read-Only

- `team_name` (String) Name of the team that owns this severity.
- `call` (Boolean) Trigger phone-call alerts at this severity.
- `sms` (Boolean) Trigger SMS alerts at this severity.
- `email` (Boolean) Trigger e-mail alerts at this severity.
- `push` (Boolean) Trigger mobile push alerts at this severity.
- `critical_alert` (Boolean) Trigger critical push alerts that bypass silent mode.
- `severity_group_id` (String) Severity group this severity belongs to.

## Endpoints

- `GET /api/v3/severities`
- `GET /api/v3/severities/:id`
- `POST /api/v3/severities`
- `PATCH /api/v3/severities/:id`
- `DELETE /api/v3/severities/:id`
