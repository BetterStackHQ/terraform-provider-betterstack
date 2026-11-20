---
page_title: "betterstack_metadata Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A metadata annotation: a key and its values, attached to a monitor, heartbeat, incident, or integration.
---

# betterstack_metadata (Data Source)

A metadata annotation: a key and its values, attached to a monitor, heartbeat, incident, or integration.

## Example Usage

```terraform
data "betterstack_metadata" "example" {
  owner_type = "Monitor"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `owner_type` (String) Type of object this metadata is attached to: Monitor, Heartbeat, Incident, WebhookIntegration, EmailIntegration, IncomingWebhook or AmazonSnsIntegration. Changing it together with owner_id on update moves the key to another object.
- `owner_id` (String) ID of the object this metadata is attached to.

### Read-Only

- `team_name` (String) Name of the team that owns the object this metadata is attached to.
- `key` (String) Metadata key. Renaming it on update keeps the entry's ID and values.
- `values` (Attributes List, see [nested schema](#nestedatt--values)) Values attached to the key. The notes below describe each value's shape. Required on create. On update, the values you send replace all current values, and an empty list removes the key.
- `mode` (String) What create does when the object already has this key: `replace` overwrites its values, `merge` adds the new values and skips duplicates. Valid values: `replace`, `merge`.

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
