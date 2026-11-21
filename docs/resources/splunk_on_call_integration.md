---
page_title: "betterstack_splunk_on_call_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.
---

# betterstack_splunk_on_call_integration (Resource)

Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.

-> Replaces `betteruptime_splunk_oncall_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_splunk_on_call_integration" "example" {
  url = "https://alert.victorops.com/integrations/generic/20131114/alert/api-key-placeholder/routing-key-placeholder"
}
```

## Schema

### Required

- `url` (String, Sensitive) Splunk On-Call REST endpoint URL.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Integration name. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `notify_alongside_primary_responder` (Boolean) Also notify the primary on-call responder. Defaults to `true`.

### Read-Only

- `id` (String) The id of this splunk_on_call_integration.
- `team_name` (String) Name of the team that owns this integration.

## Endpoints

- `GET /api/v2/splunk-on-calls`
- `GET /api/v2/splunk-on-calls/:id`
- `POST /api/v2/splunk-on-calls`
- `PATCH /api/v2/splunk-on-calls/:id`
- `DELETE /api/v2/splunk-on-calls/:id`

-> Splunk On-Call integrations and outgoing webhooks share one set of ids. The list returns only Splunk On-Call integrations, but get, update and delete also find an outgoing webhook by its id, so pass only an id this list returned.

## Import

```shell
terraform import betterstack_splunk_on_call_integration.example 12345
```
