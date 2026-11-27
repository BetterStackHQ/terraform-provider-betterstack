---
page_title: "Better Stack Provider"
description: |-
  Manage Better Stack resources for uptime monitoring, on-call, telemetry and error tracking with one provider and one API token.
---

# Better Stack Provider

The `betterstack` provider manages [Better Stack](https://betterstack.com) resources. It covers uptime monitors,
heartbeats, on-call and escalation, status pages, log sources, collectors, dashboards, error tracking
and everything else in the Better Stack API, with 56 resources, 59 data sources
and one API token.

Moving from the `better-uptime` or `logtail` provider? See the
[migration guide](guides/migration.md).

## Example Usage

```terraform
terraform {
  required_providers {
    betterstack = {
      source = "BetterStackHQ/betterstack"
    }
  }
}

provider "betterstack" {
  # Falls back to the BETTERSTACK_API_TOKEN environment variable.
  api_token = var.betterstack_api_token
}

resource "betterstack_monitor" "example" {
  url = "https://uptime.betterstack.com"
}
```

## Schema

### Optional

- `api_token` (String, Sensitive) Better Stack API token. Falls back to the `BETTERSTACK_API_TOKEN` environment variable.
- `endpoint` (String) The API base URL to use instead of Better Stack's, such as a test server.
- `api_retry_max` (Number) How many times a failed request is retried. Defaults to `4`.
- `api_retry_wait_min` (Number) The shortest wait before a retry, in seconds. Defaults to `10`.
- `api_retry_wait_max` (Number) The longest wait before a retry, in seconds, also when Better Stack asks to wait longer. Defaults to `300`.
- `api_timeout` (Number) How long one request may take, in seconds. Defaults to `60`, and `0` turns the timeout off.
- `api_rate_limit` (Number) The most requests per second. Defaults to `8`, and `0` turns the limit off.
- `api_rate_burst` (Number) How many requests may go out at once above the rate. Defaults to twice the rate, at least `10`, when set to `0`.
- `team_id` (String) Selects the team when the API token can reach several.

Retry waits can't be negative, and `api_retry_wait_min` can't be more than `api_retry_wait_max`.
A create, or another change that isn't safe to repeat, is retried only after Better Stack answers
that too many requests were sent, never after a network error.
