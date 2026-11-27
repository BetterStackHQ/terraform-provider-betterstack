---
page_title: "betterstack_incidents_incoming_webhook_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An inbound webhook URL that turns received payloads into incidents, using field and rule extraction.
---

# betterstack_incidents_incoming_webhook_integration (Resource)

An inbound webhook URL that turns received payloads into incidents, using field and rule extraction.

-> Replaces `betteruptime_incoming_webhook`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create an incoming webhook integration

```terraform
resource "betterstack_incidents_incoming_webhook_integration" "example" {
  name = "Backend service"
}
```

### Alert through an escalation policy

```terraform
data "betterstack_severity" "high" {
  name = "High severity"
}

resource "betterstack_escalation_policy" "on_call" {
  name = "On-call, then the whole team"
  steps = [
    { type = "escalation", wait_before = 0, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "current_on_call" }] },
    { type = "escalation", wait_before = 600, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "entire_team" }] },
  ]
}

resource "betterstack_incidents_incoming_webhook_integration" "alert_through_an_escalation_policy" {
  name                 = "Backend service"
  escalation_policy_id = betterstack_escalation_policy.on_call.id
}
```

## Schema

### Optional

- `name` (String) The name of the incoming webhook shown in the dashboard.
- `escalation_policy_id` (String) ID of the escalation policy for alerts, or `null` to alert the current on-call person directly. Returned as a string.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents. Defaults to `false`.
- `call` (Boolean) Whether to call the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `sms` (Boolean) Whether to send an SMS to the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `email` (Boolean) Whether to send an e-mail to the on-call person. Ignored while an escalation policy is set. Defaults to `true`.
- `push` (Boolean) Whether to send a push notification to the on-call person. Ignored while an escalation policy is set. Defaults to `false`.
- `critical_alert` (Boolean) Whether to send a critical push notification that ignores the mute switch and Do Not Disturb mode. Ignored while an escalation policy is set. Defaults to `false`.
- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `team_wait` (Number) Seconds to wait before escalating to the whole team. `0` alerts the whole team right away, and `null` never escalates to the team. Ignored while an escalation policy is set.
- `recovery_period` (Number) Seconds a matching resolved payload must hold before the incident resolves. It applies only while `resolved_rule_type` isn't `unused`. Valid values: `0`, `60`, `180`, `300`, `900`, `1800`, `3600`, `7200`. Defaults to `0`.
- `started_rule_type` (String) When requests open an incident. Every value other than `unused` needs at least one rule. When `started_alert_id_field` finds an alert ID that matches an open incident, the request goes to that incident instead of opening a new one. Defaults to `unused`.
  - `unused`: Every request that doesn't acknowledge or resolve an incident opens one.
  - `any`: A request opens an incident when at least one of `started_rules` matches.
  - `all`: A request opens an incident when all of `started_rules` match.
- `acknowledged_rule_type` (String) When requests acknowledge an incident. Any value other than `unused` needs `started_alert_id_field`, `acknowledged_alert_id_field` and at least one rule. Defaults to `unused`.
  - `unused`: Requests never acknowledge an incident.
  - `any`: A request acknowledges an incident when at least one of `acknowledged_rules` matches.
  - `all`: A request acknowledges an incident when all of `acknowledged_rules` match.
- `resolved_rule_type` (String) When requests resolve an incident. Any value other than `unused` needs `started_alert_id_field`, `resolved_alert_id_field` and at least one rule. Resolving is checked first, then acknowledging, then opening, and a request that matches no open incident can still open one. Defaults to `unused`.
  - `unused`: Requests never resolve an incident.
  - `any`: A request resolves an incident when at least one of `resolved_rules` matches.
  - `all`: A request resolves an incident when all of `resolved_rules` match.
- `cause_field` (Attributes) How to extract the incident cause from the request that starts an incident. When it's missing or extracts nothing, the cause is `Unknown cause`. (see [below for nested schema](#nestedatt--cause_field))
- `title_field` (Attributes) Field that extracts the incident title, with the same keys as `cause_field`. (see [below for nested schema](#nestedatt--title_field))
- `started_alert_id_field` (Attributes) How to extract the alert ID from the request that starts an incident, with the same keys as `cause_field`. Later requests acknowledge or resolve the open incident with the same alert ID. Required when `acknowledged_rule_type` or `resolved_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--started_alert_id_field))
- `acknowledged_alert_id_field` (Attributes) How to extract the alert ID from a request that acknowledges an incident, with the same keys as `cause_field`. Required when `acknowledged_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--acknowledged_alert_id_field))
- `resolved_alert_id_field` (Attributes) How to extract the alert ID from a request that resolves an incident, with the same keys as `cause_field`. Required when `resolved_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--resolved_alert_id_field))
- `other_started_fields` (Attributes List) Custom fields extracted from a started payload, stored as incident metadata. (see [below for nested schema](#nestedatt--other_started_fields))
- `other_acknowledged_fields` (Attributes List) Custom fields extracted from an acknowledging payload. (see [below for nested schema](#nestedatt--other_acknowledged_fields))
- `other_resolved_fields` (Attributes List) Custom fields extracted from a resolving payload. (see [below for nested schema](#nestedatt--other_resolved_fields))
- `started_rules` (Attributes List, Computed) The conditions `started_rule_type` checks. At least one rule is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--started_rules))
- `acknowledged_rules` (Attributes List, Computed) The conditions `acknowledged_rule_type` checks, with the same keys as `started_rules`. At least one is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--acknowledged_rules))
- `resolved_rules` (Attributes List, Computed) The conditions `resolved_rule_type` checks, with the same keys as `started_rules`. At least one is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--resolved_rules))
- `metadata` (Attributes Map, Computed) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. User references are returned as `u123` and team references as `t123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). A key accepts a single value or a list. Values under one key must all be the same type and can't repeat. An update replaces all of the integration's metadata with the object, `{}` removes it all, and leaving `metadata` out keeps it. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. (see [below for nested schema](#nestedatt--metadata))

### Read-Only

- `id` (String) The ID of this incoming webhook integration.
- `url` (String, Sensitive) The URL that receives your webhooks. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `sample_query_string` (String) The query string of the sample request saved in the dashboard to preview extraction.
- `sample_headers` (String) The headers of the sample request saved in the dashboard to preview extraction.
- `sample_body` (String) The body of the sample request saved in the dashboard to preview extraction.
- `created_at` (String (RFC 3339)) When the integration was created.
- `updated_at` (String (RFC 3339)) When the integration was last updated.
- `team_name` (String) Name of the team that owns this integration.

<a id="nestedatt--cause_field"></a>
### Nested Schema for `cause_field`

Required:

- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--title_field"></a>
### Nested Schema for `title_field`

Required:

- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--started_alert_id_field"></a>
### Nested Schema for `started_alert_id_field`

Required:

- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--acknowledged_alert_id_field"></a>
### Nested Schema for `acknowledged_alert_id_field`

Required:

- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--resolved_alert_id_field"></a>
### Nested Schema for `resolved_alert_id_field`

Required:

- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Required and unique within its list for custom fields. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_started_fields"></a>
### Nested Schema for `other_started_fields`

Required:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_acknowledged_fields"></a>
### Nested Schema for `other_acknowledged_fields`

Required:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_resolved_fields"></a>
### Nested Schema for `other_resolved_fields`

Required:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `field_target` (String) Part of the payload this field reads.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later requests to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--started_rules"></a>
### Nested Schema for `started_rules`

Required:

- `rule_target` (String) Part of the payload this rule matches.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.

Optional:

- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.

Read-Only:

- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--acknowledged_rules"></a>
### Nested Schema for `acknowledged_rules`

Required:

- `rule_target` (String) Part of the payload this rule matches.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.

Optional:

- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.

Read-Only:

- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--resolved_rules"></a>
### Nested Schema for `resolved_rules`

Required:

- `rule_target` (String) Part of the payload this rule matches.
  - `query_string`: A parameter in the request's query string.
  - `header`: A request header.
  - `body`: The whole request body, as plain text.
  - `json`: A value in the request body, read as JSON.
  - `xml`: An element in the request body, read as XML.
  - `sns_envelope`: A value of the SNS envelope, which only Amazon SNS integrations receive.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.

Optional:

- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, or a CSS selector for `xml`. Not used with `body`.

Read-Only:

- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--metadata"></a>
### Nested Schema for `metadata`

Optional:

- `values` (Attributes List) (see [below for nested schema](#nestedatt--metadata--values))

<a id="nestedatt--metadata--values"></a>
### Nested Schema for `metadata.values`

Optional:

- `type` (String) `String`, or the type of the referenced item, which is one of `User`, `Team`, `EscalationPolicy`, `OnCallSchedule`, `Monitor`, `SlackIntegration`, `LinearIntegration`, `JiraIntegration`, `MicrosoftTeamsIntegration`, `ZapierZap`, `OutgoingWebhook`, `PagerDutyIntegration` or `Source`.
- `value` (String) The text of a `String` value.
- `item_id` (String) ID of the referenced item. For a user it's `u123` or their e-mail. For a team it's `t123`, and for any other item it's the item's own ID. Responses return users and teams in this form. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `name` (String) Referenced item's name, or `null` when the item has no name or no longer exists. An item can also be found by its name instead of `item_id`.
- `email` (String) Referenced user's e-mail, returned for a User. A user can also be found by e-mail.

-> The `cause_field`, `title_field`, `started_alert_id_field`, `acknowledged_alert_id_field` and `resolved_alert_id_field` attributes each hold one extraction field, as an object or `null`. `other_started_fields`, `other_acknowledged_fields`, `other_resolved_fields`, `started_rules`, `acknowledged_rules` and `resolved_rules` are lists. Each list is the complete set for its event and replaces the current one.

-> A metadata object replaces all keys, including keys written through the metadata endpoints at `/api/v4/metadata`. Leaving metadata out keeps it unchanged, and an empty object, `{}`, clears it. Write an object's metadata either here or through those endpoints, not both.

## Import

Import an existing incoming webhook integration by its ID. Get the ID from the [`betterstack_incidents_incoming_webhook_integration` data source](../data-sources/incidents_incoming_webhook_integration.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_incidents_incoming_webhook_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_incidents_incoming_webhook_integration.example 12345
```
