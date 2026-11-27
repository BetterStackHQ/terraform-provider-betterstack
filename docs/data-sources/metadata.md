---
page_title: "betterstack_metadata Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A metadata annotation is a key and its values, attached to a monitor, heartbeat, incident, or integration.
---

# betterstack_metadata (Data Source)

A metadata annotation is a key and its values, attached to a monitor, heartbeat, incident, or integration.

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
- `owner_type` (String) Type of object this metadata is attached to, one of `Monitor`, `Heartbeat`, `Incident`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`, `AmazonSnsIntegration` or `CallRouting`. Create and update accept every type except `CallRouting`. Changing it together with `owner_id` moves the key to another object.
- `owner_id` (String) ID of the object this metadata is attached to.

### Optional

- `team_id` (String) Team that owns the object this metadata is attached to, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `key` (String) Metadata key. Renaming it keeps the entry's ID and values.
- `values` (Attributes List) Values attached to the key. (see [below for nested schema](#nestedatt--values))
- `mode` (String) What create does when the object already has this key.
  - `replace`: Overwrites the key's values.
  - `merge`: Adds the new values to the key and skips duplicates.
- `team_name` (String) Name of the team that owns the object this metadata is attached to.

<a id="nestedatt--values"></a>
### Nested Schema for `values`

Read-Only:

- `type` (String) `String`, or the type of the referenced item, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`.
- `value` (String) The text of a `String` value.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.
