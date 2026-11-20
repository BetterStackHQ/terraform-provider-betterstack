---
page_title: "betterstack_monitor Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  An uptime monitor that periodically checks a URL, host, or service and alerts on failures.
---

# betterstack_monitor (Data Source)

An uptime monitor that periodically checks a URL, host, or service and alerts on failures.

## Example Usage

```terraform
data "betterstack_monitor" "example" {
  url = "https://uptime.betterstack.com"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `url` (String) URL or host the monitor checks. Required for every type except Playwright, which uses scenario_name.
- `pronounceable_name` (String) Human-friendly name used in alerts.

### Read-Only

- `team_name` (String) Name of the team that owns this monitor.
- `metadata` (Attributes Map, see [nested schema](#nestedatt--metadata)) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. Team references are returned as team IDs such as `t123`. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`.
- `monitor_type` (String) What kind of check to run. Valid values: `status`, `expected_status_code`, `keyword`, `keyword_absence`, `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap`, `dns`, `playwright`.
- `monitor_group_id` (Number) ID of the monitor group this monitor belongs to.
- `paused` (Boolean) Whether checks are paused.
- `check_frequency` (Number) How often to check, in seconds. The monitor can check less often than this. effective_check_frequency shows how often it actually checks.
- `effective_check_frequency` (Number) How often the monitor actually checks, in seconds. This is check_frequency, unless checks are slowed down, or raised to the request timeout or to the plan's minimum after a downgrade.
- `effective_check_frequency_reason` (String) Why effective_check_frequency differs from check_frequency. `null` when they match.
- `request_timeout` (Number) Check timeout. In seconds for HTTP and Playwright checks, and in milliseconds for `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap` and `dns` monitors. A value of 0.5 is saved as 500.
- `recovery_period` (Number) How long checks must keep passing before the incident resolves, in seconds.
- `confirmation_period` (Number) How long to keep confirming a failure before alerting, in seconds.
- `required_keyword` (String) Keyword that must be present, or absent for keyword_absence monitors.
- `http_method` (String) HTTP method for HTTP checks.
- `request_body` (String) Request body sent with HTTP checks.
- `request_headers` (Attributes List, see [nested schema](#nestedatt--request_headers)) Custom request headers, as a list of `{name, value}` objects. A list you send replaces all headers.
- `expected_status_codes` (List of Number) Status codes that count as up, for expected_status_code monitors.
- `follow_redirects` (Boolean) Follow HTTP redirects.
- `remember_cookies` (Boolean) Persist cookies across redirects.
- `verify_ssl` (Boolean) Fail the check on an invalid TLS certificate.
- `auth_username` (String, Sensitive) HTTP basic-auth username.
- `auth_password` (String, Sensitive) HTTP basic-auth password.
- `regions` (List of String) Regions to check from: us, eu, as or au. If none are set, the monitor type's default regions are returned.
- `ip_version` (String) IP version to check over. Valid values: `ipv4`, `ipv6`.
- `port` (String) Port to check, for `tcp`, `udp`, `smtp`, `pop` and `imap` monitors.
- `proxy_host` (String) Proxy host to route the check through.
- `proxy_port` (Number) Proxy port to route the check through.
- `email` (Boolean) Alert by e-mail.
- `sms` (Boolean) Alert by SMS.
- `call` (Boolean) Alert by phone call.
- `push` (Boolean) Alert by push notification.
- `critical_alert` (Boolean) Alert by a critical push notification that bypasses silent mode.
- `team_wait` (Number) Seconds before escalating to the whole team.
- `policy_id` (String) ID of the escalation policy used for alerts. Returned as a string. You can send it as a number or a string.
- `expiration_policy_id` (Number) ID of the escalation policy for SSL certificate and domain expiration alerts.
- `ssl_expiration` (Number) Days before the SSL certificate expires to start alerting. `-1` turns it off.
- `domain_expiration` (Number) Days before the domain expires to start alerting. `-1` turns it off.
- `maintenance_from` (String) Start time of the maintenance window, as `HH:MM:SS`.
- `maintenance_to` (String) End time of the maintenance window, as `HH:MM:SS`.
- `maintenance_timezone` (String) Time zone of the maintenance window.
- `maintenance_days` (Set of String) Days of the week the maintenance window applies to.
- `playwright_script` (String) Playwright script to run, for Playwright monitors.
- `scenario_name` (String) Playwright scenario name. Playwright monitors send it instead of url.
- `environment_variables` (Dynamic, Sensitive) Environment variables for the Playwright run. An object you send replaces all of them.
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

- `GET /api/v2/monitors`
- `GET /api/v2/monitors/:id`
- `POST /api/v2/monitors`
- `PATCH /api/v2/monitors/:id`
- `DELETE /api/v2/monitors/:id`
