---
page_title: "Better Stack Provider"
description: |-
  Manage Better Stack (uptime monitoring, on-call, telemetry, error tracking) resources with one provider and one API token.
---

# Better Stack Provider

The `betterstack` provider manages [Better Stack](https://betterstack.com) resources: uptime monitors,
heartbeats, on-call and escalation, status pages, log sources, collectors, dashboards, error tracking,
and everything else in the Better Stack API (56 resources, 61 data sources,
one API token).

Every resource, attribute and endpoint on this page is generated from the machine-readable
description of the APIs themselves, so the provider cannot drift from what the APIs serve.

Migrating from the legacy `better-uptime` or `logtail` providers? See the
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
  monitor_type = "status"
  url          = "https://uptime.betterstack.com"
}
```

## Schema

### Optional

- `api_token` (String, Sensitive) Better Stack API token. Falls back to the `BETTERSTACK_API_TOKEN` environment variable.
- `endpoint` (String) Override the API base URL (used in tests).
- `api_retry_max` (Number) Maximum retries per request; default `4`.
- `api_retry_wait_min` (Number) Minimum retry delay in seconds; default `10`.
- `api_retry_wait_max` (Number) Maximum retry delay in seconds, including `Retry-After`; default `300`.
- `api_timeout` (Number) Timeout per HTTP request in seconds; default `60`, `0` disables the timeout.
- `api_rate_limit` (Number) Maximum HTTP requests per second; default `8`, `0` disables rate limiting.
- `api_rate_burst` (Number) Request burst allowance; default `0` chooses twice the rate, at least `10`.

The same six operational control names are supported by both legacy providers. Retry waits must
be nonnegative and the minimum must not exceed the maximum. Non-idempotent writes retry only
an explicit HTTP 429; ambiguous network failures are not replayed. Public data sources can be
queried without a token; authenticated endpoints require it before making a request.
- `better_stack_team_id` (String) Selects the team when the API token can reach several.
- `team_id` (String) Selects the team when the API token can reach several.
- `team_name` (String) Selects the team when the API token can reach several.
