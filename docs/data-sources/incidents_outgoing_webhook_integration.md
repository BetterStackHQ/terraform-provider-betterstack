---
page_title: "betterstack_incidents_outgoing_webhook_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Sends a webhook to an external URL when incidents, on-call, or monitors change.
---

# betterstack_incidents_outgoing_webhook_integration (Data Source)

Sends a webhook to an external URL when incidents, on-call, or monitors change.

## Example Usage

```terraform
data "betterstack_incidents_outgoing_webhook_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the outgoing webhook integration to look up.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) The integration's name, shown in the dashboard.
- `url` (String) The URL the webhook is sent to. For an `incident_change` webhook, it can use the `$INCIDENT_ID`, `$TEAM_ID` and `$METADATA` variables.
- `trigger_type` (String) What triggers the webhook. Fixed once the integration is created.
  - `incident_change`: Sends the webhook when an incident starts, is acknowledged, resolved or reopened, or gets a comment.
  - `on_call_change`: Sends the webhook when the person on call changes, or when their holiday starts or ends during their shift.
  - `monitor_change`: Sends the webhook when a monitor is created, updated, paused, unpaused or removed.
- `metadata_api_version` (String) How the `$METADATA` variables in the URL and `body_template` render.
  - `v2`: Renders `$METADATA.<key>` as the key's last plain-text value.
  - `v3`: Renders `$METADATA.<key>` as all the key's values joined by commas. `$METADATA_ARRAY` and `$METADATA.<key>.values` list each value with its type.
- `on_incident_started` (Boolean) Whether to send the webhook when an incident starts. Only for an `incident_change` webhook. Defaults to `true`.
- `on_incident_acknowledged` (Boolean) Whether to send the webhook when an incident is acknowledged. Only for an `incident_change` webhook. Defaults to `true`.
- `on_incident_resolved` (Boolean) Whether to send the webhook when an incident is resolved. Only for an `incident_change` webhook. Defaults to `true`.
- `on_incident_reopened` (Boolean) Whether to send the webhook when an incident is reopened. Only for an `incident_change` webhook. Defaults to `true`.
- `on_incident_comment` (Boolean) Whether to send the webhook when a comment is added to an incident. The default body then also has a `comment` object, whose `attributes` hold the comment's `content`, `created_at`, `updated_at`, `user_id` and `user_email`. Only for an `incident_change` webhook. Defaults to `true`.
- `notify_alongside_primary_responder` (Boolean) Whether to send the webhook alongside the primary responder when the incident has no escalation policy. Only for an `incident_change` webhook. Defaults to `true`.
- `paused` (Boolean) Whether the integration is paused. A paused integration sends no webhooks.
- `custom_webhook_template_attributes` (Attributes) Custom request method, headers, body and basic authentication. Without it, an `incident_change` or `monitor_change` webhook sends the incident or monitor under `data`, with what happened in `event`, such as `started` or `created`. An `on_call_change` webhook then sends the on-call calendar's `message`, `emails`, `on_holiday_emails`, `change`, `calendar_id`, `calendar_name` and `primary_calendar` under `data`. It comes back as `custom_webhook_template`. (see [below for nested schema](#nestedatt--custom_webhook_template_attributes))
- `custom_webhook_template` (Attributes) The saved request template, returned under `included`. `null` when the webhook sends its default body. (see [below for nested schema](#nestedatt--custom_webhook_template))
- `team_name` (String) Name of the team that owns this integration.

<a id="nestedatt--custom_webhook_template_attributes"></a>
### Nested Schema for `custom_webhook_template_attributes`

Read-Only:

- `id` (String) ID of the template saved on the integration.
- `http_method` (String) The HTTP method of the request, `post` by default. A GET request is sent without a body. Valid values: `head`, `get`, `post`, `put`, `patch`, `options`.
- `headers_template` (Attributes List) Custom request headers, as a list of objects with `name` and `value`. Header values are sent as written, without template variables. (see [below for nested schema](#nestedatt--custom_webhook_template_attributes--headers_template))
- `body_template` (String) Request body template, as text or as a JSON object. Reads return it as text. An `incident_change` template can use `$INCIDENT_ID`, `$INCIDENT_URL`, `$NAME`, `$STATUS`, `$CAUSE`, `$URL`, `$HTTP_METHOD`, `$MONITOR_ID`, `$RESOURCE_ID`, `$RESOURCE_TYPE`, `$TEAM_ID`, `$TEAM_NAME`, `$INCIDENT_GROUP_ID`, `$REQUESTER`, `$STARTED_AT`, `$ACKNOWLEDGED_AT`, `$ACKNOWLEDGED_BY`, `$RESOLVED_AT`, `$RESOLVED_BY`, `$RESPONSE_URL`, `$RESPONSE_CONTENT`, `$RESPONSE_OPTIONS`, `$SCREENSHOT_URL`, `$INTEGRATION_IMAGE_URL`, `$REGIONS`, `$ESCALATION_POLICY_ID`, `$CALL`, `$SMS`, `$EMAIL`, `$PUSH`, `$CRITICAL_ALERT`, `$INTEGRATION_ALERT_ID`, `$PLAYWRIGHT_SUMMARY`, `$COMMENT_ID`, `$COMMENT_CONTENT`, `$COMMENT_CREATED_AT`, `$COMMENT_AUTHOR_NAME` and `$COMMENT_AUTHOR_EMAIL`. A `monitor_change` template can use `$MONITOR_ID`, `$PRONOUNCEABLE_NAME`, `$URL`, `$MONITOR_TYPE`, `$STATUS`, `$MONITOR_GROUP_ID`, `$ESCALATION_POLICY_ID`, `$EXPIRATION_ESCALATION_POLICY_ID`, `$LAST_CHECKED_AT`, `$REQUIRED_KEYWORD`, `$VERIFY_SSL`, `$CHECK_FREQUENCY`, `$CALL`, `$SMS`, `$EMAIL`, `$PUSH`, `$CRITICAL_ALERT`, `$TEAM_WAIT`, `$HTTP_METHOD`, `$REQUEST_TIMEOUT`, `$RECOVERY_PERIOD`, `$REQUEST_BODY`, `$PAUSED_AT`, `$SSL_EXPIRATION`, `$DOMAIN_EXPIRATION`, `$REGIONS` and `$PORT`. An `on_call_change` template can use `$MESSAGE`, `$EMAILS`, `$EMAILS_JSON_ARRAY`, `$ON_HOLIDAY_EMAILS`, `$ON_HOLIDAY_EMAILS_JSON_ARRAY`, `$CHANGE`, `$ON_CALL_CALENDAR_ID`, `$ON_CALL_CALENDAR_NAME` and `$DEFAULT_ON_CALL_CALENDAR`. Incident and monitor templates can also use the [metadata variables](https://betterstack.com/docs/uptime/webhooks/#metadata-variables). A time such as `$STARTED_AT` renders like `2026-07-01 12:00:00 UTC`, and the same name ending in `_ISO8601` renders it in ISO 8601. A variable that holds a list, such as `$REGIONS` or `$EMAILS_JSON_ARRAY`, renders as JSON, so put it in the template without quotes.
- `auth_username` (String) The username for basic authentication. Set both username and password, or neither.
- `auth_password` (String) The password for basic authentication. Set both username and password, or neither.

<a id="nestedatt--custom_webhook_template_attributes--headers_template"></a>
### Nested Schema for `custom_webhook_template_attributes.headers_template`

Read-Only:

- `name` (String) The header's name, such as `Content-Type`.
- `value` (String) The header's value, sent as written.

<a id="nestedatt--custom_webhook_template"></a>
### Nested Schema for `custom_webhook_template`

Read-Only:

- `http_method` (String) The HTTP method of the request, `post` by default. A GET request is sent without a body. Valid values: `head`, `get`, `post`, `put`, `patch`, `options`.
- `headers_template` (Attributes List) Custom request headers, as a list of objects with `name` and `value`. Header values are sent as written, without template variables. (see [below for nested schema](#nestedatt--custom_webhook_template--headers_template))
- `body_template` (String) Request body template, as text. It can use the [metadata variables](https://betterstack.com/docs/uptime/webhooks/#metadata-variables).
- `auth_username` (String) Basic-auth username.

<a id="nestedatt--custom_webhook_template--headers_template"></a>
### Nested Schema for `custom_webhook_template.headers_template`

Read-Only:

- `name` (String) The header's name, such as `Content-Type`.
- `value` (String) The header's value, sent as written.
