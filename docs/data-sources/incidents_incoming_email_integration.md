---
page_title: "betterstack_incidents_incoming_email_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An inbound e-mail address that turns received e-mails into incidents, using field and rule extraction.
---

# betterstack_incidents_incoming_email_integration (Data Source)

An inbound e-mail address that turns received e-mails into incidents, using field and rule extraction.

## Example Usage

```terraform
data "betterstack_incidents_incoming_email_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the incoming e-mail integration to look up.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) The name of the e-mail integration shown in the dashboard.
- `escalation_policy_id` (String) ID of the escalation policy for alerts, or `null` to alert the current on-call person directly. Returned as a string.
- `email_address` (String) The address to forward your e-mails to. E-mails sent to it open incidents.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents.
- `call` (Boolean) Whether to call the on-call person. Ignored while an escalation policy is set.
- `sms` (Boolean) Whether to send an SMS to the on-call person. Ignored while an escalation policy is set.
- `email` (Boolean) Whether to send an e-mail to the on-call person. Ignored while an escalation policy is set.
- `push` (Boolean) Whether to send a push notification to the on-call person. Ignored while an escalation policy is set.
- `critical_alert` (Boolean) Whether to send a critical push notification that ignores the mute switch and Do Not Disturb mode. Ignored while an escalation policy is set.
- `team_wait` (Number) Seconds to wait before escalating to the whole team. `0` alerts the whole team right away, and `null` never escalates to the team. Ignored while an escalation policy is set.
- `recovery_period` (Number) Seconds a matching resolved e-mail must hold before the incident resolves. It applies only while `resolved_rule_type` isn't `unused`. Valid values: `0`, `60`, `180`, `300`, `900`, `1800`, `3600`, `7200`.
- `started_rule_type` (String) When e-mails open an incident. Every value other than `unused` needs at least one rule. When `started_alert_id_field` finds an alert ID that matches an open incident, the e-mail goes to that incident instead of opening a new one.
  - `unused`: Every e-mail that doesn't acknowledge or resolve an incident opens one.
  - `any`: An e-mail opens an incident when at least one of `started_rules` matches.
  - `all`: An e-mail opens an incident when all of `started_rules` match.
- `acknowledged_rule_type` (String) When e-mails acknowledge an incident. Any value other than `unused` needs `started_alert_id_field`, `acknowledged_alert_id_field` and at least one rule.
  - `unused`: E-mails never acknowledge an incident.
  - `any`: An e-mail acknowledges an incident when at least one of `acknowledged_rules` matches.
  - `all`: An e-mail acknowledges an incident when all of `acknowledged_rules` match.
- `resolved_rule_type` (String) When e-mails resolve an incident. Any value other than `unused` needs `started_alert_id_field`, `resolved_alert_id_field` and at least one rule. Resolving is checked first, then acknowledging, then opening, and an e-mail that matches no open incident can still open one.
  - `unused`: E-mails never resolve an incident.
  - `any`: An e-mail resolves an incident when at least one of `resolved_rules` matches.
  - `all`: An e-mail resolves an incident when all of `resolved_rules` match.
- `cause_field` (Attributes) How to extract the incident cause from the e-mail that starts an incident. Defaults to the whole subject. (see [below for nested schema](#nestedatt--cause_field))
- `title_field` (Attributes) Field that extracts the incident title, with the same keys as `cause_field`. (see [below for nested schema](#nestedatt--title_field))
- `started_alert_id_field` (Attributes) How to extract the alert ID from the e-mail that starts an incident, with the same keys as `cause_field`. Later e-mails acknowledge or resolve the open incident with the same alert ID. Required when `acknowledged_rule_type` or `resolved_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--started_alert_id_field))
- `acknowledged_alert_id_field` (Attributes) How to extract the alert ID from an e-mail that acknowledges an incident, with the same keys as `cause_field`. Required when `acknowledged_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--acknowledged_alert_id_field))
- `resolved_alert_id_field` (Attributes) How to extract the alert ID from an e-mail that resolves an incident, with the same keys as `cause_field`. Required when `resolved_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--resolved_alert_id_field))
- `other_started_fields` (Attributes List) Custom fields extracted from a started e-mail, stored as incident metadata. (see [below for nested schema](#nestedatt--other_started_fields))
- `other_acknowledged_fields` (Attributes List) Custom fields extracted from an acknowledging e-mail. (see [below for nested schema](#nestedatt--other_acknowledged_fields))
- `other_resolved_fields` (Attributes List) Custom fields extracted from a resolving e-mail. (see [below for nested schema](#nestedatt--other_resolved_fields))
- `started_rules` (Attributes List) The conditions `started_rule_type` checks. At least one rule is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--started_rules))
- `acknowledged_rules` (Attributes List) The conditions `acknowledged_rule_type` checks, with the same keys as `started_rules`. At least one is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--acknowledged_rules))
- `resolved_rules` (Attributes List) The conditions `resolved_rule_type` checks, with the same keys as `started_rules`. At least one is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--resolved_rules))
- `created_at` (String (RFC 3339)) When the integration was created.
- `updated_at` (String (RFC 3339)) When the integration was last updated.
- `team_name` (String) Name of the team that owns this integration.
- `metadata` (Attributes Map) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. User references are returned as `u123` and team references as `t123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). A key accepts a single value or a list. Values under one key must all be the same type and can't repeat. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. (see [below for nested schema](#nestedatt--metadata))

<a id="nestedatt--cause_field"></a>
### Nested Schema for `cause_field`

Read-Only:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--title_field"></a>
### Nested Schema for `title_field`

Read-Only:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--started_alert_id_field"></a>
### Nested Schema for `started_alert_id_field`

Read-Only:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--acknowledged_alert_id_field"></a>
### Nested Schema for `acknowledged_alert_id_field`

Read-Only:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--resolved_alert_id_field"></a>
### Nested Schema for `resolved_alert_id_field`

Read-Only:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_started_fields"></a>
### Nested Schema for `other_started_fields`

Read-Only:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_acknowledged_fields"></a>
### Nested Schema for `other_acknowledged_fields`

Read-Only:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_resolved_fields"></a>
### Nested Schema for `other_resolved_fields`

Read-Only:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later e-mails to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `field_target` (String) Part of the e-mail this field reads.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--started_rules"></a>
### Nested Schema for `started_rules`

Read-Only:

- `rule_target` (String) Part of the e-mail this rule matches.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.
- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--acknowledged_rules"></a>
### Nested Schema for `acknowledged_rules`

Read-Only:

- `rule_target` (String) Part of the e-mail this rule matches.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.
- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--resolved_rules"></a>
### Nested Schema for `resolved_rules`

Read-Only:

- `rule_target` (String) Part of the e-mail this rule matches.
  - `from_email`: The sender's e-mail address.
  - `subject`: The e-mail's subject.
  - `body`: The e-mail's body.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.
- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--metadata"></a>
### Nested Schema for `metadata`

Read-Only:

- `values` (Attributes List) (see [below for nested schema](#nestedatt--metadata--values))

<a id="nestedatt--metadata--values"></a>
### Nested Schema for `metadata.values`

Read-Only:

- `type` (String) `String`, or the type of the referenced item, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`.
- `value` (String) The text of a `String` value.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.
