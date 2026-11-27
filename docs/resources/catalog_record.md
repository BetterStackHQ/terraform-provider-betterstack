---
page_title: "betterstack_catalog_record Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A catalog record is one entry in a catalog relation, with values for its attributes.
---

# betterstack_catalog_record (Resource)

A catalog record is one entry in a catalog relation, with values for its attributes.

## Example Usage

### Name the team that owns a service

```terraform
resource "betterstack_catalog_record" "name_the_team_that_owns_a_service" {
  catalog_relation_id = "67"
  attributes          = [{ attribute = { id = "56", name = "Service" }, values = [{ type = "String", value = "checkout-api" }] }, { attribute = { id = "57", name = "Owner" }, values = [{ type = "Team", item_id = "t123" }] }]
}
```

### Route a service to its escalation policy and tech lead

```terraform
resource "betterstack_catalog_record" "route_a_service_to_its_escalation_policy_and_tech_lead" {
  catalog_relation_id = "67"
  attributes          = [{ attribute = { id = "56", name = "Service" }, values = [{ type = "String", value = "payments-api" }] }, { attribute = { id = "58", name = "Escalation policy" }, values = [{ type = "EscalationPolicy", name = "Payments on-call" }] }, { attribute = { id = "59", name = "Tech lead" }, values = [{ type = "User", email = "jane.cooper@example.com" }] }]
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `catalog_relation_id` (String, Required) The ID of the parent this catalog relation belongs to. Changing it replaces the resource.

### Required

- `attributes` (Attributes List) The record's values, one entry per attribute that has values. Each entry sets one catalog attribute, by id or name, to one or more values. Required on create and update, and an update replaces all of them. (see [below for nested schema](#nestedatt--attributes))

### Read-Only

- `id` (String) The ID of this catalog record.

<a id="nestedatt--attributes"></a>
### Nested Schema for `attributes`

Required:

- `attribute` (Attributes) The catalog attribute to set, by id or name. Responses include both. (see [below for nested schema](#nestedatt--attributes--attribute))
- `values` (Attributes List) Values to set for this attribute. (see [below for nested schema](#nestedatt--attributes--values))

<a id="nestedatt--attributes--attribute"></a>
### Nested Schema for `attributes.attribute`

Optional:

- `id` (String) The catalog attribute's ID.
- `name` (String) The catalog attribute's name.

<a id="nestedatt--attributes--values"></a>
### Nested Schema for `attributes.values`

Optional:

- `type` (String) `String`, or the type of the referenced item, which is one of User, Team, EscalationPolicy, OnCallSchedule, Monitor, SlackIntegration, LinearIntegration, JiraIntegration, MicrosoftTeamsIntegration, ZapierZap, OutgoingWebhook, PagerDutyIntegration or Source. Required except for `{value}`.
- `value` (String) The text of a String value. Required when `type` is `String`.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). A reference needs one of `item_id`, `name` or `email`.
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`. A monitor, a Slack integration and a Zapier Zap can't be looked up by name.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.

~> The API also serves this object under `/api/v4/catalog/relations/{catalog_relation_id}/records/batch`. One Terraform resource maps to one endpoint family; use `terraform import` or the data source for objects created elsewhere.

-> A value is `{type: "String", value: "..."}` or a reference to an item in the organization. Write a reference as `{type: <ItemType>, item_id: <id>}`, or look the item up with `{type:, name:}` or `{type: "User", email:}`. Responses always return `{type:, item_id:, name:}`, plus `email` for a User. Reference item types are User, Team, EscalationPolicy, OnCallSchedule, Monitor, SlackIntegration, LinearIntegration, JiraIntegration, MicrosoftTeamsIntegration, ZapierZap, OutgoingWebhook, PagerDutyIntegration, Source. A user's `item_id` is `u123` or their e-mail, and a team's is `t123`, in responses and requests. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).

## Import

Import an existing catalog record by its ID as `catalog_relation_id/id`. Get the ID via [List catalog records](https://betterstack.com/docs/uptime/api/list-all-catalog-records/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_catalog_record.example
  id = "<catalog_relation_id>/12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_catalog_record.example <catalog_relation_id>/12345
```
