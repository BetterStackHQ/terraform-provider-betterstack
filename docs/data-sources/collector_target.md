---
page_title: "betterstack_collector_target Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A database or process a collector scrapes for metrics, for example Postgres, nginx, or a Prometheus endpoint.
---

# betterstack_collector_target (Data Source)

A database or process a collector scrapes for metrics, for example Postgres, nginx, or a Prometheus endpoint.

## Example Usage

```terraform
data "betterstack_collector_target" "example" {
  collector_id = "12345"
  id           = "12345"
}
```

## Schema

### Address

- `collector_id` (String, Required) The ID of the parent this collector belongs to. Changing it replaces the resource.

### Required

- `id` (String) The ID of the collector target to look up.

### Optional

- `team_id` (String) Team that owns the collector's source, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `kind` (String) Target type. Reads can also return `pgcat`, `node_exporter` and `kube_state_metrics`, which the collector finds by itself and the API doesn't accept on create. Valid values: `postgres`, `pgbouncer`, `mysql`, `redis`, `mongodb`, `memcached`, `elasticsearch`, `nginx`, `apache`, `kafka`, `prometheus`, `traefik`.
- `host` (String) Target host. For the process kinds this is the host the collector runs on.
- `enabled` (Boolean) Whether the collector scrapes this target. A new target is enabled unless it's set to `false`. Disabling it sets `status` to `disabled`, and enabling it again checks the target again.
- `port` (Number) The target's port. `null` when a target is found by its `listen_ip` or `service` instead. Applies only when `kind` is `postgres` or `pgbouncer` or `mysql` or `redis` or `mongodb` or `memcached` or `elasticsearch` or `nginx` or `apache` or `kafka` or `traefik`.
- `service` (String) The service name. Required for `nginx`, `apache`, `kafka`, `prometheus` and `traefik` targets, and accepted and returned only for those. Applies only when `kind` is `nginx` or `apache` or `kafka` or `prometheus` or `traefik`.
- `listen_ip` (String) The IP address the process listens on. Accepted and returned only for `nginx`, `apache`, `kafka` and `traefik`, and `null` when not set. Applies only when `kind` is `nginx` or `apache` or `kafka` or `traefik`.
- `endpoint` (String) The full scrape URL, like `http://10.0.0.5:9090/metrics`. Required for `prometheus` targets and returned only for them. Applies only when `kind` is `prometheus`.
- `scheme` (String) The connection scheme. Accepted and returned only for `elasticsearch`. Valid values: `http`, `https`. Applies only when `kind` is `elasticsearch`.
- `username` (String) The username used to connect, or `null` when not set. Accepted and returned only for `postgres`, `pgbouncer`, `mysql`, `redis`, `mongodb`, `memcached` and `elasticsearch`. Applies only when `kind` is `postgres` or `pgbouncer` or `mysql` or `redis` or `mongodb` or `memcached` or `elasticsearch`.
- `password` (String, Sensitive) Password for database and `elasticsearch` targets.
- `api_key` (String, Sensitive) API key for `elasticsearch` targets.
- `ssl_mode` (String) The SSL mode for the connection. Required for `postgres` and `pgbouncer`, and accepted and returned only for those. Applies only when `kind` is `postgres` or `pgbouncer`.
  - `disable`: Connects without SSL.
  - `require`: Connects over SSL without verifying the server's certificate.
  - `verify-ca`: Connects over SSL and checks that a trusted certificate authority signed the server's certificate.
- `tls` (String) The TLS mode for the connection. Required for `mysql`, and accepted and returned only for it. Applies only when `kind` is `mysql`.
  - `false`: Connects without TLS.
  - `true`: Connects over TLS and verifies the server's certificate.
  - `skip-verify`: Connects over TLS without verifying the server's certificate.
  - `preferred`: Uses TLS when the server supports it, without verifying the server's certificate.
- `status` (String) The connection status. An update that includes `host`, `port`, `service`, `listen_ip`, `endpoint`, `scheme`, `username`, `password`, `api_key`, `ssl_mode` or `tls` sets it back to `pending`.
  - `pending`: The collector hasn't reported on the target yet.
  - `up`: The collector collects metrics from the target.
  - `unreachable`: The collector can't connect to the target or hasn't seen it lately.
  - `down`: The collector reaches the target, but its check fails.
  - `invalid_credentials`: The target rejects the credentials.
  - `configuration_required`: Metrics are collected, but the database needs more setup, such as the `pg_stat_statements` extension.
  - `disabled`: The target is paused and the collector doesn't check it.
  - `removed`: The collector no longer finds the target.
- `container` (String) Detected container.
- `created_at` (String (RFC 3339)) When the target was created.
- `updated_at` (String (RFC 3339)) When the target was last updated.
