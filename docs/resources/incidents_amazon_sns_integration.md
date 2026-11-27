---
page_title: "betterstack_incidents_amazon_sns_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives Amazon SNS messages and turns them into incidents using extraction fields and rules. Also shows whether AWS has confirmed the subscription.
---

# betterstack_incidents_amazon_sns_integration (Resource)

Receives Amazon SNS messages and turns them into incidents using extraction fields and rules. Also shows whether AWS has confirmed the subscription.

-> Replaces `betteruptime_amazon_sns_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create an Amazon SNS integration

```terraform
resource "betterstack_incidents_amazon_sns_integration" "example" {
  name = "Backend alerts"
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

resource "betterstack_incidents_amazon_sns_integration" "alert_through_an_escalation_policy" {
  name                 = "Backend alerts"
  escalation_policy_id = betterstack_escalation_policy.on_call.id
}
```

## Schema

### Optional

- `name` (String) The name of the Amazon SNS integration shown in the dashboard.
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
- `started_rule_type` (String) When notifications open an incident. Every value other than `unused` needs at least one rule. When `started_alert_id_field` finds an alert ID that matches an open incident, the notification goes to that incident instead of opening a new one. Defaults to `unused`.
  - `unused`: Every notification that doesn't acknowledge or resolve an incident opens one.
  - `any`: A notification opens an incident when at least one of `started_rules` matches.
  - `all`: A notification opens an incident when all of `started_rules` match.
- `acknowledged_rule_type` (String) When notifications acknowledge an incident. Any value other than `unused` needs `started_alert_id_field`, `acknowledged_alert_id_field` and at least one rule. Defaults to `unused`.
  - `unused`: Notifications never acknowledge an incident.
  - `any`: A notification acknowledges an incident when at least one of `acknowledged_rules` matches.
  - `all`: A notification acknowledges an incident when all of `acknowledged_rules` match.
- `resolved_rule_type` (String) When notifications resolve an incident. Any value other than `unused` needs `started_alert_id_field`, `resolved_alert_id_field` and at least one rule. Resolving is checked first, then acknowledging, then opening, and a notification that matches no open incident can still open one. Defaults to `unused`.
  - `unused`: Notifications never resolve an incident.
  - `any`: A notification resolves an incident when at least one of `resolved_rules` matches.
  - `all`: A notification resolves an incident when all of `resolved_rules` match.
- `cause_field` (Attributes, Computed) How to extract the incident cause from the notification that starts an incident. Defaults to the whole message. (see [below for nested schema](#nestedatt--cause_field))
- `title_field` (Attributes) How to extract a custom incident title from the notification that starts an incident, with the same keys as `cause_field`. The dashboard's default extracts the whole SNS `Subject`, with `field_target` set to `sns_envelope`, `target_field` to `Subject` and `match_type` to `match_everything`. `null` extracts no title. Leaving it out clears it, on create too. (see [below for nested schema](#nestedatt--title_field))
- `started_alert_id_field` (Attributes, Computed) How to extract the alert ID from the notification that starts an incident, with the same keys as `cause_field`. Later notifications acknowledge or resolve the open incident with the same alert ID. Required when `acknowledged_rule_type` or `resolved_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--started_alert_id_field))
- `acknowledged_alert_id_field` (Attributes) How to extract the alert ID from a notification that acknowledges an incident, with the same keys as `cause_field`. Required when `acknowledged_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--acknowledged_alert_id_field))
- `resolved_alert_id_field` (Attributes) How to extract the alert ID from a notification that resolves an incident, with the same keys as `cause_field`. Required when `resolved_rule_type` isn't `unused`. (see [below for nested schema](#nestedatt--resolved_alert_id_field))
- `other_started_fields` (Attributes List) Custom fields extracted from a started payload, stored as incident metadata. (see [below for nested schema](#nestedatt--other_started_fields))
- `other_acknowledged_fields` (Attributes List) Custom fields extracted from an acknowledging payload. (see [below for nested schema](#nestedatt--other_acknowledged_fields))
- `other_resolved_fields` (Attributes List) Custom fields extracted from a resolving payload. (see [below for nested schema](#nestedatt--other_resolved_fields))
- `started_rules` (Attributes List, Computed) The conditions `started_rule_type` checks. At least one rule is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--started_rules))
- `acknowledged_rules` (Attributes List, Computed) The conditions `acknowledged_rule_type` checks, with the same keys as `started_rules`. At least one is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--acknowledged_rules))
- `resolved_rules` (Attributes List, Computed) The conditions `resolved_rule_type` checks, with the same keys as `started_rules`. At least one is required when the type is `any` or `all`. (see [below for nested schema](#nestedatt--resolved_rules))
- `metadata` (Attributes Map, Computed) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. User references are returned as `u123` and team references as `t123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). A key accepts a single value or a list. Values under one key must all be the same type and can't repeat. An update replaces all of the integration's metadata with the object, `{}` removes it all, and leaving `metadata` out keeps it. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. (see [below for nested schema](#nestedatt--metadata))

### Read-Only

- `id` (String) The ID of this Amazon SNS integration.
- `topic_arn` (String) The ARN of the subscribed Amazon SNS topic, or `null` until AWS confirms the subscription.
- `subscription_state` (String) Whether AWS has confirmed the SNS subscription.
  - `awaiting`: AWS hasn't confirmed the subscription yet.
  - `active`: AWS has confirmed the subscription.
- `url` (String, Sensitive) The endpoint to subscribe your Amazon SNS topic to. Leave **Enable raw message delivery** off on the subscription, or every notification is rejected with `400`. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `sample_query_string` (String) The query string of the sample notification saved in the dashboard to preview extraction.
- `sample_headers` (String) The headers of the sample notification saved in the dashboard to preview extraction.
- `sample_body` (String) The body of the sample SNS message saved in the dashboard to preview extraction.
- `created_at` (String (RFC 3339)) When the integration was created.
- `updated_at` (String (RFC 3339)) When the integration was last updated.
- `team_name` (String) Name of the team that owns this integration.

<a id="nestedatt--cause_field"></a>
### Nested Schema for `cause_field`

Required:

- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Custom fields require it, and it must be unique within its list. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--title_field"></a>
### Nested Schema for `title_field`

Required:

- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Custom fields require it, and it must be unique within its list. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--started_alert_id_field"></a>
### Nested Schema for `started_alert_id_field`

Required:

- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Custom fields require it, and it must be unique within its list. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--acknowledged_alert_id_field"></a>
### Nested Schema for `acknowledged_alert_id_field`

Required:

- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Custom fields require it, and it must be unique within its list. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--resolved_alert_id_field"></a>
### Nested Schema for `resolved_alert_id_field`

Required:

- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `name` (String) The metadata key a custom field's value is saved under. Custom fields require it, and it must be unique within its list. Not used by the cause, title and alert ID fields.
- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_started_fields"></a>
### Nested Schema for `other_started_fields`

Required:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_acknowledged_fields"></a>
### Nested Schema for `other_acknowledged_fields`

Required:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--other_resolved_fields"></a>
### Nested Schema for `other_resolved_fields`

Required:

- `name` (String) The metadata key to save the extracted value under. Required, and must be unique within the array.
- `field_target` (String) Which part of the notification to read.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) What to extract.
  - `match_before`: Extracts the text before `content`.
  - `match_after`: Extracts the text after `content`.
  - `match_between`: Extracts the text between `content_before` and `content_after`.
  - `match_regex`: Extracts the match of the regular expression in `content`, or its capture group.
  - `match_everything`: Extracts the whole value.

Optional:

- `special_type` (String) The built-in role of this field. `null` for a custom field.
  - `alert_id`: Extracts the alert ID that ties later notifications to the same incident.
  - `cause`: Extracts the incident cause.
  - `title`: Extracts the incident title.
- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.
- `content` (String) The text to extract around for `match_before` and `match_after`, or the regular expression for `match_regex`. Required for those three.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) Where extraction stops for `match_between`. Leave it empty to extract to the end.

<a id="nestedatt--started_rules"></a>
### Nested Schema for `started_rules`

Required:

- `rule_target` (String) Which part of the notification the rule checks.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.

Optional:

- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.

Read-Only:

- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--acknowledged_rules"></a>
### Nested Schema for `acknowledged_rules`

Required:

- `rule_target` (String) Which part of the notification the rule checks.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.

Optional:

- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.

Read-Only:

- `created_at` (String (RFC 3339)) When the rule was created.
- `updated_at` (String (RFC 3339)) When the rule was last changed.

<a id="nestedatt--resolved_rules"></a>
### Nested Schema for `resolved_rules`

Required:

- `rule_target` (String) Which part of the notification the rule checks.
  - `query_string`: The query string of AWS's own HTTPS request, which never contains your message.
  - `header`: A header of AWS's own HTTPS request, which never contains your message.
  - `body`: The message your publisher sent, as plain text.
  - `json`: A value in the message your publisher sent, read as JSON.
  - `xml`: An element in the message your publisher sent, read as XML.
  - `sns_envelope`: A value of the SNS envelope, such as its `Subject`.
- `match_type` (String) How to compare the value with `content`.
  - `contains`: Matches when the value contains `content`.
  - `contains_not`: Matches when the value doesn't contain `content`.
  - `matches_regex`: Matches when the value matches the regular expression in `content`.
  - `matches_regex_not`: Matches when the value doesn't match the regular expression in `content`.
  - `equals`: Matches when the value equals `content`.
  - `equals_not`: Matches when the value doesn't equal `content`.
- `content` (String) The value to compare with. Must be a valid regular expression for `matches_regex` and `matches_regex_not`.

Optional:

- `target_field` (String) Which value to read. It's a dot-separated path such as `alert.id` for `json` and `query_string`, a header name for `header`, a CSS selector for `xml`, or `TopicArn`, `MessageId` or `Subject` for `sns_envelope`. Not used with `body`.

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

-> The receiving URL has the form `/api/v1/incoming-webhook/:token`. A new integration extracts the cause and alert ID from the message body, and the title from the SNS Subject. `topic_arn` and `subscription_state` are read-only.

-> A metadata object replaces all keys, including keys written through the metadata endpoints at `/api/v4/metadata`. Leaving metadata out keeps it unchanged, and an empty object, `{}`, clears it. Write an object's metadata either here or through those endpoints, not both.

## Import

Import an existing Amazon SNS integration by its ID. Get the ID from the [`betterstack_incidents_amazon_sns_integration` data source](../data-sources/incidents_amazon_sns_integration.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_incidents_amazon_sns_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_incidents_amazon_sns_integration.example 12345
```
