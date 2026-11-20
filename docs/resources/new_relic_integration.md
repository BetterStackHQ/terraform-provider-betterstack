---
page_title: "betterstack_new_relic_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from New Relic and opens Better Stack incidents.
---

# betterstack_new_relic_integration (Resource)

Receives alerts from New Relic and opens Better Stack incidents.

-> Replaces `betteruptime_new_relic_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_new_relic_integration" "example" {
}
```

## Schema

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Integration name.
- `policy_id` (Number) ID of the escalation policy used for alerts. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `alerting_rule` (String) Which New Relic alert levels open incidents: alert only, or alert and warn. Valid values: `alert`, `alert_and_warn`. Defaults to `alert`.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents. Defaults to `false`.
- `call` (Boolean) Alert by phone call. Defaults to `false`.
- `sms` (Boolean) Alert by SMS. Defaults to `false`.
- `email` (Boolean) Alert by e-mail. Defaults to `true`.
- `push` (Boolean) Alert by push notification. Defaults to `false`.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode. Defaults to `false`.
- `team_wait` (Number) Seconds before escalating to the whole team. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `recovery_period` (Number) Seconds a resolved alert must hold before the incident resolves. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this new_relic_integration.
- `team_name` (String) Name of the team that owns this integration.
- `webhook_url` (String, Sensitive) The URL to paste into New Relic. Alerts New Relic sends to it open incidents. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.

## Endpoints

- `GET /api/v2/new-relic-integrations`
- `GET /api/v2/new-relic-integrations/:id`
- `POST /api/v2/new-relic-integrations`
- `PATCH /api/v2/new-relic-integrations/:id`
- `DELETE /api/v2/new-relic-integrations/:id`

## Import

```shell
terraform import betterstack_new_relic_integration.example 12345
```
