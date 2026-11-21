---
page_title: "betterstack_amazon_sns_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Receives Amazon SNS messages and turns them into incidents using extraction fields and rules. Also shows whether AWS has confirmed the subscription.
---

# betterstack_amazon_sns_integration (Resource)

Receives Amazon SNS messages and turns them into incidents using extraction fields and rules. Also shows whether AWS has confirmed the subscription.

-> Replaces `betteruptime_amazon_sns_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_amazon_sns_integration" "example" {
}
```

## Schema

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Integration name. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `metadata` (Attributes Map, see [nested schema](#nestedatt--metadata)) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. Team references are returned as team IDs such as `t123`. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `escalation_policy_id` (String) ID of the escalation policy used for alerts. Returned as a string. You can send it as a number or a string. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents. Defaults to `false`.
- `call` (Boolean) Alert by phone call. Defaults to `false`.
- `sms` (Boolean) Alert by SMS. Defaults to `false`.
- `email` (Boolean) Alert by e-mail. Defaults to `true`.
- `push` (Boolean) Alert by push notification. Defaults to `false`.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode. Defaults to `false`.
- `team_wait` (Number) Seconds before escalating to the whole team. Defaults to `180`.
- `recovery_period` (Number) Seconds a matching resolved payload must hold before the incident resolves. One of 0, 60, 180, 300, 900, 1800, 3600, 7200. Has no effect while `resolved_rule_type` is `unused`. Defaults to `0`.
- `started_rule_type` (String) How the started rules open an incident: `any` when one rule matches, `all` when every rule matches. `unused` turns off opening incidents by rules. Valid values: `unused`, `any`, `all`. Defaults to `unused`.
- `acknowledged_rule_type` (String) How the acknowledged rules acknowledge the incident: `any` when one rule matches, `all` when every rule matches. `unused` turns off acknowledging by rules. Valid values: `unused`, `any`, `all`. Defaults to `unused`.
- `resolved_rule_type` (String) How the resolved rules resolve the incident: `any` when one rule matches, `all` when every rule matches. `unused` turns off resolving by rules. Valid values: `unused`, `any`, `all`. Defaults to `unused`.
- `cause_field` (Attributes, see [nested schema](#nestedatt--cause_field)) Field that extracts the incident cause. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `title_field` (Attributes, see [nested schema](#nestedatt--title_field)) Field that extracts the incident title. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `started_alert_id_field` (Attributes, see [nested schema](#nestedatt--started_alert_id_field)) Field that extracts the alert ID, which links the started, acknowledged and resolved payloads of one alert. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `acknowledged_alert_id_field` (Attributes, see [nested schema](#nestedatt--acknowledged_alert_id_field)) Field that extracts the alert ID from an acknowledging payload. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `resolved_alert_id_field` (Attributes, see [nested schema](#nestedatt--resolved_alert_id_field)) Field that extracts the alert ID from a resolving payload. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `other_started_fields` (Attributes List, see [nested schema](#nestedatt--other_started_fields)) Custom fields extracted from a started payload, stored as incident metadata. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `other_acknowledged_fields` (Attributes List, see [nested schema](#nestedatt--other_acknowledged_fields)) Custom fields extracted from an acknowledging payload. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `other_resolved_fields` (Attributes List, see [nested schema](#nestedatt--other_resolved_fields)) Custom fields extracted from a resolving payload. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `started_rules` (Attributes List, see [nested schema](#nestedatt--started_rules)) Rules that classify a payload as opening an incident. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `acknowledged_rules` (Attributes List, see [nested schema](#nestedatt--acknowledged_rules)) Rules that classify a payload as acknowledging the incident. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `resolved_rules` (Attributes List, see [nested schema](#nestedatt--resolved_rules)) Rules that classify a payload as resolving the incident. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.

### Read-Only

- `id` (String) The id of this amazon_sns_integration.
- `topic_arn` (String) ARN of the SNS topic, recorded when AWS confirms the subscription.
- `subscription_state` (String) Whether AWS has confirmed the subscription. Valid values: `awaiting`, `active`.
- `team_name` (String) Name of the team that owns this integration.
- `url` (String, Sensitive) The URL Amazon SNS posts messages to. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `sample_query_string` (String) Query string from the most recent received payload, for troubleshooting extraction.
- `sample_headers` (String) Headers from the most recent received payload, for troubleshooting extraction.
- `sample_body` (String) Body from the most recent received payload, for troubleshooting extraction.
- `created_at` (String (RFC 3339)) When the integration was created.
- `updated_at` (String (RFC 3339)) When the integration was last updated.

<a id="nestedatt--metadata"></a>
### Nested Schema for `metadata`

- `values` (Attributes List, see [nested schema](#nestedatt--metadata--values))

<a id="nestedatt--metadata--values"></a>
### Nested Schema for `metadata.values`

- `type` (String) `String`, or the type of the referenced item.
- `value` (String) The text of a String value.
- `item_id` (String) ID of the referenced item: a number, or for a team its ID starting with `t`, such as `t123`.
- `name` (String) Referenced item's name.
- `email` (String) Referenced user's e-mail.

<a id="nestedatt--cause_field"></a>
### Nested Schema for `cause_field`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--title_field"></a>
### Nested Schema for `title_field`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--started_alert_id_field"></a>
### Nested Schema for `started_alert_id_field`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--acknowledged_alert_id_field"></a>
### Nested Schema for `acknowledged_alert_id_field`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--resolved_alert_id_field"></a>
### Nested Schema for `resolved_alert_id_field`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--other_started_fields"></a>
### Nested Schema for `other_started_fields`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--other_acknowledged_fields"></a>
### Nested Schema for `other_acknowledged_fields`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--other_resolved_fields"></a>
### Nested Schema for `other_resolved_fields`

- `name` (String)
- `special_type` (String) The built-in role of this field. `null` for a custom field. Valid values: `alert_id`, `cause`, `title`.
- `field_target` (String) Part of the payload this field reads. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How the value is extracted. Valid values: `match_before`, `match_after`, `match_between`, `match_regex`, `match_everything`.
- `content` (String) The delimiter for `match_before` and `match_after`, or the regex for `match_regex`.
- `content_before` (String) The start delimiter for `match_between`.
- `content_after` (String) The end delimiter for `match_between`.

<a id="nestedatt--started_rules"></a>
### Nested Schema for `started_rules`

- `rule_target` (String) Part of the payload this rule matches. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How to compare the content. Valid values: `contains`, `contains_not`, `matches_regex`, `matches_regex_not`, `equals`, `equals_not`.
- `content` (String) Value or pattern to match.

<a id="nestedatt--acknowledged_rules"></a>
### Nested Schema for `acknowledged_rules`

- `rule_target` (String) Part of the payload this rule matches. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How to compare the content. Valid values: `contains`, `contains_not`, `matches_regex`, `matches_regex_not`, `equals`, `equals_not`.
- `content` (String) Value or pattern to match.

<a id="nestedatt--resolved_rules"></a>
### Nested Schema for `resolved_rules`

- `rule_target` (String) Part of the payload this rule matches. Valid values: `query_string`, `header`, `body`, `json`, `xml`, `sns_envelope`.
- `target_field` (String) The key to read in the target: a JSON path, an XML path, a header name, or an SNS envelope field (`TopicArn`, `MessageId` or `Subject`).
- `match_type` (String) How to compare the content. Valid values: `contains`, `contains_not`, `matches_regex`, `matches_regex_not`, `equals`, `equals_not`.
- `content` (String) Value or pattern to match.

## Endpoints

- `GET /api/v3/amazon-sns`
- `GET /api/v3/amazon-sns/:id`
- `POST /api/v3/amazon-sns`
- `PATCH /api/v3/amazon-sns/:id`
- `DELETE /api/v3/amazon-sns/:id`

-> cause_field, title_field, started_alert_id_field, acknowledged_alert_id_field and resolved_alert_id_field each hold one extraction field, as an object or `null`. other_started_fields, other_acknowledged_fields, other_resolved_fields, started_rules, acknowledged_rules and resolved_rules are lists. A list you send is the complete new set for that event and replaces the current one.

-> The receiving URL has the form `/api/v1/incoming-webhook/:token`. A new integration extracts the cause and alert ID from the message body, and the title from the SNS Subject. topic_arn and subscription_state are read-only.

-> A metadata object you send replaces all keys, including keys written through the metadata endpoints (`/api/v4/metadata`). Leave metadata out to keep it unchanged, or send an empty object (`{}`) to clear it. Manage an object's metadata either here or through those endpoints, not both.

## Import

```shell
terraform import betterstack_amazon_sns_integration.example 12345
```
