---
page_title: "betterstack_outgoing_webhook_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Sends a webhook to an external URL when incidents, on-call, or monitors change.
---

# betterstack_outgoing_webhook_integration (Resource)

Sends a webhook to an external URL when incidents, on-call, or monitors change.

## Example Usage

```terraform
resource "betterstack_outgoing_webhook_integration" "example" {
  url          = "https://example.com/webhooks/better-stack"
  trigger_type = "incident_change"
}
```

## Schema

### Required

- `url` (String) Destination URL.
- `trigger_type` (String) What triggers the webhook. Fixed once the integration is created. Valid values: `incident_change`, `on_call_change`, `monitor_change`. Can only be set when creating the resource; changing it afterwards replaces the resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Integration name. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `metadata_api_version` (String) How template variables show metadata. With v2, a key shows its last plain-text value. With v3, it shows every value with its type. Valid values: `v2`, `v3`. Defaults to `v3`.
- `on_incident_started` (Boolean) Fire when an incident starts. Defaults to `true` when `trigger_type` is `incident_change`; other trigger types don't take it.
- `on_incident_acknowledged` (Boolean) Fire when an incident is acknowledged. Defaults to `true` when `trigger_type` is `incident_change`; other trigger types don't take it.
- `on_incident_resolved` (Boolean) Fire when an incident is resolved. Defaults to `true` when `trigger_type` is `incident_change`; other trigger types don't take it.
- `on_incident_reopened` (Boolean) Fire when an incident is reopened. Defaults to `true` when `trigger_type` is `incident_change`; other trigger types don't take it.
- `on_incident_comment` (Boolean) Fire on a new incident comment. Defaults to `true` when `trigger_type` is `incident_change`; other trigger types don't take it.
- `notify_alongside_primary_responder` (Boolean) Also notify the primary on-call responder. Defaults to `true` when `trigger_type` is `incident_change`; other trigger types don't take it.
- `custom_webhook_template_attributes` (Attributes, see [nested schema](#nestedatt--custom_webhook_template_attributes)) Custom request method, headers, body and basic authentication. Leave it out to keep the saved template unchanged. Write-only: the API never returns it; Terraform keeps the configured value. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.

### Read-Only

- `id` (String) The id of this outgoing_webhook_integration.
- `team_name` (String) Name of the team that owns this integration.
- `custom_webhook_template` (Attributes, see [nested schema](#nestedatt--custom_webhook_template)) The saved request template, returned under `included`. `null` when there is none.

<a id="nestedatt--custom_webhook_template_attributes"></a>
### Nested Schema for `custom_webhook_template_attributes`

- `id` (String) ID of the saved template, to edit it in place.
- `http_method` (String) HTTP method for the outgoing request. Valid values: `head`, `get`, `post`, `put`, `patch`, `options`.
- `headers_template` (Attributes List, see [nested schema](#nestedatt--custom_webhook_template_attributes--headers_template)) Custom request headers, as a list of `{name, value}` objects. Values can use template variables.
- `body_template` (Dynamic) Request body template, as a string or a JSON object. It can use template variables.
- `auth_username` (String) Basic authentication username sent with the request. Set both username and password, or neither.
- `auth_password` (String) Basic authentication password sent with the request. Set both username and password, or neither.

<a id="nestedatt--custom_webhook_template_attributes--headers_template"></a>
### Nested Schema for `custom_webhook_template_attributes.headers_template`

- `name` (String)
- `value` (String)

<a id="nestedatt--custom_webhook_template"></a>
### Nested Schema for `custom_webhook_template`

- `http_method` (String) HTTP method for the outgoing request. Valid values: `head`, `get`, `post`, `put`, `patch`, `options`.
- `headers_template` (Attributes List, see [nested schema](#nestedatt--custom_webhook_template--headers_template)) Custom request headers, as a list of `{name, value}` objects.
- `body_template` (String) Request body template, as a string.
- `auth_username` (String) Basic-auth username.
- `auth_password` (String) Basic-auth password.

<a id="nestedatt--custom_webhook_template--headers_template"></a>
### Nested Schema for `custom_webhook_template.headers_template`

- `name` (String)
- `value` (String)

## Endpoints

- `GET /api/v2/outgoing-webhooks`
- `GET /api/v2/outgoing-webhooks/:id`
- `POST /api/v2/outgoing-webhooks`
- `PATCH /api/v2/outgoing-webhooks/:id`
- `DELETE /api/v2/outgoing-webhooks/:id`

-> An ID that belongs to a Splunk On-Call or incident.io integration returns a 404 error here.

## Import

```shell
terraform import betterstack_outgoing_webhook_integration.example 12345
```
