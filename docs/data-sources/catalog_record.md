---
page_title: "betterstack_catalog_record Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog record is one entry in a catalog relation, with values for its attributes.
---

# betterstack_catalog_record (Data Source)

A catalog record is one entry in a catalog relation, with values for its attributes.

## Example Usage

```terraform
data "betterstack_catalog_record" "example" {
  catalog_relation_id = "12345"
  id                  = "12345"
}
```

## Schema

### Address

- `catalog_relation_id` (String, Required) The ID of the parent this catalog relation belongs to. Changing it replaces the resource.

### Required

- `id` (String) The ID of the catalog record to look up.

### Read-Only

- `attributes` (Attributes List) The record's values, one entry per attribute that has values. Each entry sets one catalog attribute, by id or name, to one or more values. Required on create and update, and an update replaces all of them. (see [below for nested schema](#nestedatt--attributes))

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Read-Only:

- `attribute` (Attributes) The catalog attribute to set, by id or name. Responses include both. (see [below for nested schema](#nestedatt--attributes--attribute))
- `values` (Attributes List) Values to set for this attribute. (see [below for nested schema](#nestedatt--attributes--values))

<a id="nestedatt--attributes--attribute"></a>
### Nested Schema for `attributes.attribute`

Read-Only:

- `id` (String) The catalog attribute's ID.
- `name` (String) The catalog attribute's name.

<a id="nestedatt--attributes--values"></a>
### Nested Schema for `attributes.values`

Read-Only:

- `type` (String) `String`, or the type of the referenced item, which is one of User, Team, EscalationPolicy, OnCallSchedule, Monitor, SlackIntegration, LinearIntegration, JiraIntegration, MicrosoftTeamsIntegration, ZapierZap, OutgoingWebhook, PagerDutyIntegration or Source. Required except for `{value}`.
- `value` (String) The text of a String value. Required when `type` is `String`.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). A reference needs one of `item_id`, `name` or `email`.
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`. A monitor, a Slack integration and a Zapier Zap can't be looked up by name.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.
