---
page_title: "betterstack_pagerduty_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to a PagerDuty service.
---

# betterstack_pagerduty_integration (Resource)

Forwards Better Stack incidents to a PagerDuty service.

-> Replaces `betteruptime_pagerduty_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_pagerduty_integration" "example" {
  name     = "PagerDuty alerts"
  key      = "routing-key-placeholder"
  severity = "critical"
}
```

## Schema

### Required

- `name` (String) Integration name.
- `key` (String, Sensitive) PagerDuty integration key, also called the routing key.
- `severity` (String) PagerDuty severity sent with the event. Valid values: `info`, `warning`, `error`, `critical`.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `notify_alongside_primary_responder` (Boolean) Also notify the primary on-call responder. Defaults to `true`.

### Read-Only

- `id` (String) The id of this pagerduty_integration.
- `team_name` (String) Name of the team that owns this integration.

## Endpoints

- `GET /api/v3/pagerduty-integrations`
- `GET /api/v3/pagerduty-integrations/:id`
- `POST /api/v3/pagerduty-integrations`
- `PATCH /api/v3/pagerduty-integrations/:id`
- `DELETE /api/v3/pagerduty-integrations/:id`

## Import

```shell
terraform import betterstack_pagerduty_integration.example 12345
```
