---
page_title: "betterstack_aws_cloudwatch_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives alerts from AWS CloudWatch and opens Better Stack incidents.
---

# betterstack_aws_cloudwatch_integration (Resource)

Receives alerts from AWS CloudWatch and opens Better Stack incidents.

-> Replaces `betteruptime_aws_cloudwatch_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_aws_cloudwatch_integration" "example" {
}
```

## Schema

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Integration name.
- `policy_id` (Number) ID of the escalation policy used for alerts. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents. Defaults to `false`.
- `call` (Boolean) Alert by phone call. Defaults to `false`.
- `sms` (Boolean) Alert by SMS. Defaults to `false`.
- `email` (Boolean) Alert by e-mail. Defaults to `true`.
- `push` (Boolean) Alert by push notification. Defaults to `false`.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode. Defaults to `false`.
- `team_wait` (Number) Seconds before escalating to the whole team. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `recovery_period` (Number) Seconds a resolved alert must hold before the incident resolves. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.

### Read-Only

- `id` (String) The id of this aws_cloudwatch_integration.
- `team_name` (String) Name of the team that owns this integration.
- `webhook_url` (String, Sensitive) The URL to paste into AWS CloudWatch. Alerts AWS CloudWatch sends to it open incidents. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.

## Endpoints

- `GET /api/v2/aws-cloudwatch-integrations`
- `GET /api/v2/aws-cloudwatch-integrations/:id`
- `POST /api/v2/aws-cloudwatch-integrations`
- `PATCH /api/v2/aws-cloudwatch-integrations/:id`
- `DELETE /api/v2/aws-cloudwatch-integrations/:id`

-> alerting_rule is accepted on write but not read back.

## Import

```shell
terraform import betterstack_aws_cloudwatch_integration.example 12345
```
