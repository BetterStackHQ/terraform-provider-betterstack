---
page_title: "betterstack_metadata Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A metadata annotation: a key and its values, attached to a monitor, heartbeat, incident, or integration.
---

# betterstack_metadata (Resource)

A metadata annotation: a key and its values, attached to a monitor, heartbeat, incident, or integration.

## Example Usage

```terraform
resource "betterstack_metadata" "example" {
  owner_type = "Monitor"
  owner_id   = "2"
  key        = "Owner"
  values     = [{ type = "example" }]
}
```

## Schema

### Required

- `owner_type` (String) Type of object this metadata is attached to: Monitor, Heartbeat, Incident, WebhookIntegration, EmailIntegration, IncomingWebhook or AmazonSnsIntegration. Changing it together with owner_id on update moves the key to another object.
- `owner_id` (String) ID of the object this metadata is attached to.
- `key` (String) Metadata key. Renaming it on update keeps the entry's ID and values.
- `values` (Attributes List, see [nested schema](#nestedatt--values)) Values attached to the key. The notes below describe each value's shape. Required on create. On update, the values you send replace all current values, and an empty list removes the key.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `mode` (String) What create does when the object already has this key: `replace` overwrites its values, `merge` adds the new values and skips duplicates. Valid values: `replace`, `merge`. Defaults to `replace`. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.

### Read-Only

- `id` (String) The id of this metadata.
- `team_name` (String) Name of the team that owns the object this metadata is attached to.

<a id="nestedatt--values"></a>
### Nested Schema for `values`

- `type` (String) `String`, or the type of the referenced item.
- `value` (String) The text of a String value.
- `item_id` (String) ID of the referenced item: a number, or for a team its ID starting with `t`, such as `t123`.
- `name` (String) Referenced item's name.
- `email` (String) Referenced user's e-mail.

## Endpoints

- `GET /api/v4/metadata`
- `GET /api/v4/metadata/:id`
- `POST /api/v4/metadata`
- `PATCH /api/v4/metadata/:id`
- `DELETE /api/v4/metadata/:id`

-> owner_id is the object's ID in its own resource. An owner_type doesn't always match a resource name in this API. For example, WebhookIntegration covers integrations such as Datadog, Grafana, Prometheus and AWS CloudWatch.

-> Writing the metadata attribute of a monitor, heartbeat or integration replaces all of its keys, these included. Manage an object's metadata either here or through that attribute, not both.

-> A value is a plain string, `{type: "String", value: "..."}`, or a reference to an item in the organization. To look up an item, send `{type: <ItemType>, item_id: <id>}`, `{type:, name:}` or `{type: "User", email:}`. Every response returns a reference as `{type:, item_id:, name:}`, plus email for a User. Item types: User, Team, Policy, Schedule, Endpoint, SlackIntegration, LinearIntegration, JiraIntegration, MicrosoftTeamsWebhook, ZapierWebhook, NativeWebhook, PagerDutyWebhook and Source. A team's item_id is its ID starting with `t`, such as `t123`, and is accepted the same way.

## Import

```shell
terraform import betterstack_metadata.example 12345
```
