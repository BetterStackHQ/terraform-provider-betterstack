---
page_title: "betterstack_monitor Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An uptime monitor that periodically checks a URL, host, or service and alerts on failures.
---

# betterstack_monitor (Resource)

An uptime monitor that periodically checks a URL, host, or service and alerts on failures.

-> Replaces `betteruptime_monitor`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_monitor" "example" {
  monitor_type = "status"
  url          = "https://uptime.betterstack.com"
}
```

## Schema

At least one of `url` or `scenario_name` must be configured. The provider validates this condition before creating or updating the resource.

### Required

- `monitor_type` (String) What kind of check to run. Valid values: `status`, `expected_status_code`, `keyword`, `keyword_absence`, `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap`, `dns`, `playwright`.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `metadata` (Attributes Map, see [nested schema](#nestedatt--metadata)) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. Team references are returned as team IDs such as `t123`. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `url` (String) URL or host the monitor checks. Required for every type except Playwright, which uses scenario_name.
- `pronounceable_name` (String) Human-friendly name used in alerts. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `monitor_group_id` (String) ID of the monitor group this monitor belongs to. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `paused` (Boolean) Whether checks are paused. Defaults to `false`.
- `check_frequency` (Number) How often to check, in seconds. The monitor can check less often than this. effective_check_frequency shows how often it actually checks. Defaults to `180`.
- `request_timeout` (Number) Check timeout. In seconds for HTTP and Playwright checks, and in milliseconds for `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap` and `dns` monitors. A value of 0.5 is saved as 500. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `recovery_period` (Number) How long checks must keep passing before the incident resolves, in seconds. Defaults to `180`.
- `confirmation_period` (Number) How long to keep confirming a failure before alerting, in seconds. Defaults to `0`.
- `required_keyword` (String) Keyword that must be present, or absent for keyword_absence monitors. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `http_method` (String) HTTP method for HTTP checks. Defaults to `get`.
- `request_body` (String) Request body sent with HTTP checks. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `request_headers` (Attributes List, see [nested schema](#nestedatt--request_headers)) Custom request headers, as a list of `{name, value}` objects. A list you send replaces all headers. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `expected_status_codes` (List of Number) Status codes that count as up, for expected_status_code monitors. Defaults to `[]`.
- `follow_redirects` (Boolean) Follow HTTP redirects. Defaults to `true`.
- `remember_cookies` (Boolean) Persist cookies across redirects. Defaults to `true`.
- `verify_ssl` (Boolean) Fail the check on an invalid TLS certificate. Defaults to `true`.
- `auth_username` (String, Sensitive) HTTP basic-auth username. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `auth_password` (String, Sensitive) HTTP basic-auth password. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `regions` (List of String) Regions to check from: us, eu, as or au. If none are set, the monitor type's default regions are returned. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `ip_version` (String) IP version to check over. Valid values: `ipv4`, `ipv6`. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `port` (String) Port to check, for `tcp`, `udp`, `smtp`, `pop` and `imap` monitors. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `proxy_host` (String) Proxy host to route the check through. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `proxy_port` (Number) Proxy port to route the check through. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `email` (Boolean) Alert by e-mail. Defaults to `true`.
- `sms` (Boolean) Alert by SMS. Defaults to `false`.
- `call` (Boolean) Alert by phone call. Defaults to `false`.
- `push` (Boolean) Alert by push notification. Defaults to `false`.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode. Defaults to `false`.
- `team_wait` (Number) Seconds before escalating to the whole team. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `escalation_policy_id` (String) ID of the escalation policy used for alerts. Returned as a string. You can send it as a number or a string. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `expiration_policy_id` (String) ID of the escalation policy for SSL certificate and domain expiration alerts. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `ssl_expiration` (Number) Days before the SSL certificate expires to start alerting. `-1` turns it off. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `domain_expiration` (Number) Days before the domain expires to start alerting. `-1` turns it off. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `maintenance_from` (String) Start time of the maintenance window, as `HH:MM:SS`. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `maintenance_to` (String) End time of the maintenance window, as `HH:MM:SS`. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `maintenance_timezone` (String) Time zone of the maintenance window. Defaults to `UTC`.
- `maintenance_days` (Set of String) Days of the week the maintenance window applies to. Defaults to `["mon", "tue", "wed", "thu", "fri", "sat", "sun"]`.
- `playwright_script` (String) Playwright script to run, for Playwright monitors. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `scenario_name` (String) Playwright scenario name. Playwright monitors send it instead of url.
- `environment_variables` (Dynamic, Sensitive) Environment variables for the Playwright run. An object you send replaces all of them. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.

### Read-Only

- `id` (String) The id of this monitor.
- `team_name` (String) Name of the team that owns this monitor.
- `effective_check_frequency` (Number) How often the monitor actually checks, in seconds. This is check_frequency, unless checks are slowed down, or raised to the request timeout or to the plan's minimum after a downgrade.
- `effective_check_frequency_reason` (String) Why effective_check_frequency differs from check_frequency. `null` when they match.
- `status` (String) Current monitor status.
- `last_checked_at` (String (RFC 3339)) When the monitor was last checked.
- `paused_at` (String (RFC 3339)) When the monitor was paused.
- `checks_version` (String) Version of the checks engine. Always `v2`.
- `created_at` (String (RFC 3339)) When the monitor was created.
- `updated_at` (String (RFC 3339)) When the monitor was last updated.

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

<a id="nestedatt--request_headers"></a>
### Nested Schema for `request_headers`

- `name` (String)
- `value` (String)

## Endpoints

- `GET /api/v3/monitors`
- `GET /api/v3/monitors/:id`
- `POST /api/v3/monitors`
- `PATCH /api/v3/monitors/:id`
- `DELETE /api/v3/monitors/:id`

-> A metadata object you send replaces all keys, including keys written through the metadata endpoints (`/api/v4/metadata`). Leave metadata out to keep it unchanged, or send an empty object (`{}`) to clear it. Manage an object's metadata either here or through those endpoints, not both.

-> request_headers also accepts a list of entries with an `id`, to change one existing header, or with `_destroy`, to remove one.

## Import

```shell
terraform import betterstack_monitor.example 12345
```
