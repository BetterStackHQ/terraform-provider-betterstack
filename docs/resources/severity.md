---
page_title: "betterstack_severity Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity: a named alerting level with its notification-channel settings.
---

# betterstack_severity (Resource)

A severity: a named alerting level with its notification-channel settings.

-> Replaces `betteruptime_severity`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_severity" "example" {
  name = "Low Severity"
}
```

## Schema

### Required

- `name` (String) Severity name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `call` (Boolean) Trigger phone-call alerts at this severity.
- `sms` (Boolean) Trigger SMS alerts at this severity.
- `email` (Boolean) Trigger e-mail alerts at this severity.
- `push` (Boolean) Trigger mobile push alerts at this severity.
- `critical_alert` (Boolean) Trigger critical push alerts that bypass silent mode.
- `urgency_group_id` (Number) Severity group this severity belongs to. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this severity.
- `team_name` (String) Name of the team that owns this severity.

## Endpoints

- `GET /api/v2/urgencies`
- `GET /api/v2/urgencies/:id`
- `POST /api/v2/urgencies`
- `PATCH /api/v2/urgencies/:id`
- `DELETE /api/v2/urgencies/:id`

## Import

```shell
terraform import betterstack_severity.example 12345
```
