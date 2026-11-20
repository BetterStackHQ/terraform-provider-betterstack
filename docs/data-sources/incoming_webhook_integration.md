---
page_title: "betterstack_incoming_webhook_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An inbound webhook URL that turns received payloads into incidents, using field and rule extraction.
---

# betterstack_incoming_webhook_integration (Data Source)

An inbound webhook URL that turns received payloads into incidents, using field and rule extraction.

## Example Usage

```terraform
data "betterstack_incoming_webhook_integration" "example" {
  name = "Backend service"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Integration name.

### Read-Only

- `team_name` (String) Name of the team that owns this integration.
- `metadata` (Attributes Map, see [nested schema](#nestedatt--metadata)) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. Team references are returned as team IDs such as `t123`. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`.
- `policy_id` (String) ID of the escalation policy used for alerts. Returned as a string. You can send it as a number or a string.
- `url` (String, Sensitive) URL that third-party payloads are posted to.
- `paused` (Boolean) Whether the integration is paused. A paused integration opens no incidents.
- `call` (Boolean) Alert by phone call.
- `sms` (Boolean) Alert by SMS.
- `email` (Boolean) Alert by e-mail.
- `push` (Boolean) Alert by push notification.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode.
- `team_wait` (Number) Seconds before escalating to the whole team.
- `recovery_period` (Number) Seconds a matching resolved payload must hold before the incident resolves. One of 0, 60, 180, 300, 900, 1800, 3600, 7200.
- `started_rule_type` (String) How the started rules open an incident: `any` when one rule matches, `all` when every rule matches. `unused` turns off opening incidents by rules. Valid values: `unused`, `any`, `all`.
- `acknowledged_rule_type` (String) How the acknowledged rules acknowledge the incident: `any` when one rule matches, `all` when every rule matches. `unused` turns off acknowledging by rules. Valid values: `unused`, `any`, `all`.
- `resolved_rule_type` (String) How the resolved rules resolve the incident: `any` when one rule matches, `all` when every rule matches. `unused` turns off resolving by rules. Valid values: `unused`, `any`, `all`.
- `cause_field` (Attributes, see [nested schema](#nestedatt--cause_field)) Field that extracts the incident cause.
- `title_field` (Attributes, see [nested schema](#nestedatt--title_field)) Field that extracts the incident title.
- `started_alert_id_field` (Attributes, see [nested schema](#nestedatt--started_alert_id_field)) Field that extracts the alert ID, which links the started, acknowledged and resolved payloads of one alert.
- `acknowledged_alert_id_field` (Attributes, see [nested schema](#nestedatt--acknowledged_alert_id_field)) Field that extracts the alert ID from an acknowledging payload.
- `resolved_alert_id_field` (Attributes, see [nested schema](#nestedatt--resolved_alert_id_field)) Field that extracts the alert ID from a resolving payload.
- `other_started_fields` (Attributes List, see [nested schema](#nestedatt--other_started_fields)) Custom fields extracted from a started payload, stored as incident metadata.
- `other_acknowledged_fields` (Attributes List, see [nested schema](#nestedatt--other_acknowledged_fields)) Custom fields extracted from an acknowledging payload.
- `other_resolved_fields` (Attributes List, see [nested schema](#nestedatt--other_resolved_fields)) Custom fields extracted from a resolving payload.
- `started_rules` (Attributes List, see [nested schema](#nestedatt--started_rules)) Rules that classify a payload as opening an incident.
- `acknowledged_rules` (Attributes List, see [nested schema](#nestedatt--acknowledged_rules)) Rules that classify a payload as acknowledging the incident.
- `resolved_rules` (Attributes List, see [nested schema](#nestedatt--resolved_rules)) Rules that classify a payload as resolving the incident.
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

- `GET /api/v2/incoming-webhooks`
- `GET /api/v2/incoming-webhooks/:id`
- `POST /api/v2/incoming-webhooks`
- `PATCH /api/v2/incoming-webhooks/:id`
- `DELETE /api/v2/incoming-webhooks/:id`
