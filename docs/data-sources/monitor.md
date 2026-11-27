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
- `url` (String) The URL of your website, or the host to check for `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap` and `dns` monitors, without a protocol. It's `null` on `playwright` monitors.
- `pronounceable_name` (String) The name of the monitor, which we use when we call you.
- `paused` (Boolean) Whether the monitor is paused. A paused monitor opens no incidents and sends no alerts.
- `status` (String) The monitor's current status.
  - `up`: Checks pass.
  - `down`: Checks fail.
  - `validating`: The service seems back up, but the `recovery_period` hasn't passed yet.
  - `paused`: The monitor is paused.
  - `pending`: The monitor waits for its first check.
  - `maintenance`: The monitor is in a maintenance window.

### Optional

- `team_id` (String) Team that owns this monitor, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `monitor_type` (String) What the monitor checks.
  - `status`: Checks that the URL returns a 2XX status code.
  - `expected_status_code`: Checks that the URL returns one of `expected_status_codes`.
  - `keyword`: Checks that the page contains `required_keyword`.
  - `keyword_absence`: Checks that the page doesn't contain `required_keyword`.
  - `ping`: Pings the host in `url`.
  - `tcp`: Checks that a TCP `port` on the host accepts connections.
  - `udp`: Sends `request_body` to a UDP `port` on the host and checks that the answer contains `required_keyword`.
  - `smtp`: Checks for an SMTP server on the host at `port`.
  - `pop`: Checks for a POP3 server on the host at `port`.
  - `imap`: Checks for an IMAP server on the host at `port`.
  - `dns`: Checks that the DNS server on the host answers a query for the domain in `request_body`.
  - `playwright`: Runs the scenario in `playwright_script`.
- `monitor_group_id` (String) ID of the monitor group this monitor belongs to.
- `check_frequency` (Number) How often to check, in seconds. It must be at least the `request_timeout`. A new or changed value below the plan's shortest interval is refused with a 422. A value already saved keeps saving. Its accepted values depend on `monitor_type`.
  - `playwright`: It accepts `60`, `180`, `300`, `600`, `900`, `1800`, `3600`, `21600`, `43200` or `86400`.
  - Every other type: It accepts `30`, `45`, `60`, `120`, `180`, `300`, `600`, `900` or `1800`.
- `effective_check_frequency` (Number) How often the monitor actually checks, in seconds. `effective_check_frequency_reason` says why it differs from `check_frequency`.
- `effective_check_frequency_reason` (String) Why `effective_check_frequency` differs from `check_frequency`. `null` when they're the same.
  - `slowed_down`: We check less often after repeated `429` responses or timeouts.
  - `plan_minimum`: Your plan's shortest interval between checks is longer than `check_frequency`, which was saved before the plan changed.
  - `request_timeout`: The monitor's `request_timeout` is longer than `check_frequency`.
- `request_timeout` (Number) How long a check waits before it times out. Its unit and accepted values depend on `monitor_type`.
  - `status`, `expected_status_code`, `keyword` and `keyword_absence`: In seconds. It accepts `2`, `3`, `5`, `10`, `15`, `30`, `45`, `60`, `90`, `120` or `180` and defaults to `30`.
  - `ping`, `tcp`, `udp`, `smtp`, `pop`, `imap` and `dns`: In milliseconds. It accepts `500`, `1000`, `2000`, `3000` or `5000` and defaults to `5000`.
  - `playwright`: In seconds. It accepts `15`, `30`, `45`, `60`, `120`, `180`, `240`, `300`, `360`, `480`, `600` or `900` and defaults to `30`.
- `recovery_period` (Number) How long checks must keep passing before the incident resolves, in seconds. Valid values: `0`, `60`, `180`, `300`, `900`, `1800`, `3600`, `7200`.
- `confirmation_period` (Number) How long to wait after a failure before starting an incident, in seconds, from `0` to `86400`.
- `required_keyword` (String) The keyword to look for.
- `http_method` (String) The HTTP method of the request. It accepts `get`, `head`, `post`, `put`, `patch` or `options`. Responses return it in lowercase. `head` can't be used with `keyword` or `keyword_absence` monitors.
- `request_body` (String) The body of the request. A `dns` monitor requires it. Set it to the domain to query. A `udp` monitor sends it as its content.
- `request_headers` (Attributes List) Custom request headers, as a list of `{name, value}` objects. A list replaces all headers, and an empty list removes them all. (see [below for nested schema](#nestedatt--request_headers))
- `expected_status_codes` (List of Number) HTTP status codes from `100` to `599` that count as up. A `3xx` code needs `follow_redirects` and `remember_cookies` set to `false`.
- `follow_redirects` (Boolean) Whether we follow HTTP redirects.
- `remember_cookies` (Boolean) Whether we keep cookies when redirecting.
- `verify_ssl` (Boolean) Whether a check fails on an invalid SSL certificate.
- `auth_username` (String, Sensitive) HTTP basic-auth username. `null` when none is set.
- `auth_password` (String, Sensitive) HTTP basic-auth password. It isn't returned after it's saved.
- `regions` (List of String) Regions to check from, as any combination of `us`, `eu`, `as` and `au`. When none are set, we check from all regions. Responses then list all of them.
- `ip_version` (String) The IP version to check over. `null` checks over both IPv4 and IPv6.
  - `ipv4`: Checks over IPv4 only.
  - `ipv6`: Checks over IPv6 only.
- `port` (String) The port to check. `tcp` and `udp` accept one port. `smtp` accepts `25`, `465` or `587`, `pop` accepts `110` or `995`, and `imap` accepts `143` or `993`, with several separated by commas like `25,465,587`.
- `proxy_host` (String) The host of the proxy server we send the check through, without a protocol. It can include credentials, like `user:pass@proxy.example.com`. Only for `status`, `expected_status_code`, `keyword` and `keyword_absence` monitors.
- `proxy_port` (Number) The port of the proxy server, from `1` to `65535`. Requires `proxy_host`.
- `email` (Boolean) Whether to send an e-mail to the on-call person. Ignored while an escalation policy is set.
- `sms` (Boolean) Whether to send an SMS to the on-call person. Ignored while an escalation policy is set.
- `call` (Boolean) Whether to call the on-call person. Ignored while an escalation policy is set.
- `push` (Boolean) Whether to send a push notification to the on-call person. Ignored while an escalation policy is set.
- `critical_alert` (Boolean) Whether to send a critical push notification that ignores the mute switch and Do Not Disturb mode. Ignored while an escalation policy is set.
- `team_wait` (Number) Seconds to wait before escalating to the whole team. `0` alerts the whole team right away, and `null` never escalates to the team. Ignored while an escalation policy is set.
- `escalation_policy_id` (String) ID of the escalation policy for alerts, or `null` to alert the current on-call person directly.
- `expiration_escalation_policy_id` (String) ID of the escalation policy for alerts about expiring SSL certificates and domains. Monitors of the same domain share this setting, subdomains included. For example, monitors of `example.com` and `www.example.com` share it. A value set on one monitor is copied to all of them, so set the same value on each.
- `ssl_expiration` (Number) How many days before the SSL certificate expires we alert you. `null` turns the check off and is the default. Monitors of exactly the same domain share this setting, not its subdomains. For example, monitors of `example.com/a` and `example.com/b` share it, while `www.example.com` has its own. A value set on one monitor is copied to all of them, so set the same value on each. Valid values: `1`, `2`, `3`, `7`, `14`, `30`, `60`.
- `domain_expiration` (Number) How many days before the domain expires we alert you. `null` turns the check off and is the default. Monitors of the same domain share this setting, subdomains included. For example, monitors of `example.com` and `www.example.com` share it. A value set on one monitor is copied to all of them, so set the same value on each. Valid values: `1`, `2`, `3`, `7`, `14`, `30`, `60`.
- `maintenance_from` (String) Start of the daily maintenance window, when we don't check, as `HH:MM:SS`, like `01:00:00`. `null` when no window is set.
- `maintenance_to` (String) End of the daily maintenance window, as `HH:MM:SS`, like `03:00:00`. Required with `maintenance_from`. `null` when no window is set.
- `maintenance_timezone` (String) The time zone of the maintenance window, as a Rails time zone name like `Amsterdam` or its IANA ID like `Europe/Amsterdam`. `null` when it isn't set, which works as UTC.
- `maintenance_days` (Set of String) Days of the week the maintenance window applies to, any of `mon`, `tue`, `wed`, `thu`, `fri`, `sat` and `sun`. An overnight window belongs to the day it starts.
- `playwright_script` (String) The JavaScript source of the scenario, for `playwright` monitors.
- `scenario_name` (String) The name of a `playwright` monitor's scenario. Required for `playwright` monitors, and returned only for them.
- `environment_variables` (Map of String, Sensitive) Environment variables the `playwright` scenario can use, as name and value pairs like `{"PASSWORD": "passw0rd"}`.
- `last_checked_at` (String (RFC 3339)) When the monitor was last checked.
- `paused_at` (String (RFC 3339)) When the monitor was paused.
- `checks_version` (String) Version of the checks engine. Always `v2`.
- `created_at` (String (RFC 3339)) When the monitor was created.
- `updated_at` (String (RFC 3339)) When the monitor was last updated.
- `team_name` (String) Name of the team that owns this monitor.
- `metadata` (Attributes Map) Custom metadata. Each key holds a list of values. A value is a string, written as `{type: "String", value}`, or a catalog reference, written as `{type, item_id, name, email}`. User references are returned as `u123` and team references as `t123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). Values under one key must all be the same type and can't repeat. Each key holds its list under `values`, for example `{ owner = { values = [ ... ] } }`. (see [below for nested schema](#nestedatt--metadata))
- `deletion_protection` (Boolean) Whether destroying or replacing this monitor is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.

<a id="nestedatt--request_headers"></a>
### Nested Schema for `request_headers`

Read-Only:

- `id` (String) The header's ID.
- `name` (String) The header name, like `Content-Type`.
- `value` (String) The header value, like `application/json`.

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
