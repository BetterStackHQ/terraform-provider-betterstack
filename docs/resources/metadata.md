---
page_title: "betterstack_metadata Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A metadata annotation is a key and its values, attached to a monitor, heartbeat, incident, or integration.
---

# betterstack_metadata (Resource)

A metadata annotation is a key and its values, attached to a monitor, heartbeat, incident, or integration.

## Example Usage

### Create metadata

```terraform
resource "betterstack_metadata" "example" {
  key        = "Key 1"
  values     = [{ value = "Value 1" }, { value = "Value 2" }]
  owner_id   = "2"
  owner_type = "Monitor"
}
```

### Typed values

```terraform
resource "betterstack_metadata" "typed_values" {
  key        = "Key 1"
  values     = [{ type = "User", item_id = "u123" }, { type = "User", email = "alice@example.com" }]
  owner_id   = "2"
  owner_type = "Monitor"
}
```

### Merge values

```terraform
resource "betterstack_metadata" "merge_values" {
  key        = "Key 1"
  values     = [{ value = "Value 3" }]
  mode       = "merge"
  owner_id   = "2"
  owner_type = "Monitor"
}
```

## Schema

### Required

- `owner_type` (String) Type of object this metadata is attached to, one of `Monitor`, `Heartbeat`, `Incident`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`, `AmazonSnsIntegration` or `CallRouting`. Create and update accept every type except `CallRouting`. Changing it together with `owner_id` moves the key to another object.
- `owner_id` (String) ID of the object this metadata is attached to.
- `key` (String) Metadata key. Renaming it keeps the entry's ID and values.
- `values` (Attributes List) Values attached to the key. Required on create. An update replaces all current values, and an empty list removes the key. (see [below for nested schema](#nestedatt--values))

### Optional

- `mode` (String, Forces replacement) What create does when the object already has this key. Defaults to `replace`. Better Stack never returns it, so Terraform keeps the value you set.
  - `replace`: Overwrites the key's values.
  - `merge`: Adds the new values to the key and skips duplicates.
- `team_id` (String) Team that owns the object this metadata is attached to, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this metadata.
- `team_name` (String) Name of the team that owns the object this metadata is attached to.

<a id="nestedatt--values"></a>
### Nested Schema for `values`

Optional:

- `type` (String) `String`, or the type of the referenced item, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`.
- `value` (String) The text of a `String` value.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.

-> The `owner_id` attribute is the object's ID in its own resource. An `owner_type` doesn't always match a resource name in this API. For example, WebhookIntegration covers integrations such as Datadog, Grafana, Prometheus and AWS CloudWatch.

-> Writing the metadata attribute of a monitor, heartbeat or integration replaces all of its keys, these included. Edit an object's metadata either here or through that attribute, not both.

-> A value is `{type: "String", value: "..."}` or a reference to an item in the organization. An item is looked up by `{type: <ItemType>, item_id: <id>}`, `{type:, name:}` or `{type: "User", email:}`. Every response returns a reference as `{type:, item_id:, name:}`, plus `email` for a User. Item types are User, Team, EscalationPolicy, OnCallSchedule, Monitor, SlackIntegration, LinearIntegration, JiraIntegration, MicrosoftTeamsIntegration, ZapierZap, OutgoingWebhook, PagerDutyIntegration and Source. A team's `item_id` is its ID starting with `t`, such as `t123`, and is accepted the same way.

## Import

Import an existing metadata by its ID. Get the ID from the [`betterstack_metadata` data source](../data-sources/metadata.md), which finds it by `owner_type` or `owner_id`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_metadata.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_metadata.example 12345
```
