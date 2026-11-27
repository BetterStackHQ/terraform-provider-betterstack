---
page_title: "betterstack_collector_target Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A database or process a collector scrapes for metrics, for example Postgres, nginx, or a Prometheus endpoint.
---

# betterstack_collector_target (Resource)

A database or process a collector scrapes for metrics, for example Postgres, nginx, or a Prometheus endpoint.

-> Replaces `logtail_collector_target`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Postgres

```terraform
resource "betterstack_collector_target" "postgres" {
  collector_id = "1234"
  kind         = "postgres"
  host         = "db.example.com"
  port         = 5432
  username     = "monitor"
  password     = "secret"
  ssl_mode     = "require"
}
```

### Elasticsearch

```terraform
resource "betterstack_collector_target" "elasticsearch" {
  collector_id = "1234"
  kind         = "elasticsearch"
  host         = "search.example.com"
  port         = 9200
  scheme       = "https"
  api_key      = "secret"
}
```

### Nginx

```terraform
resource "betterstack_collector_target" "nginx" {
  collector_id = "1234"
  kind         = "nginx"
  host         = "web-1"
  service      = "edge-nginx"
  listen_ip    = "10.0.0.1"
  port         = 80
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `collector_id` (String, Required) The ID of the parent this collector belongs to. Changing it replaces the resource.

### Required

- `kind` (String, Forces replacement) Target type. Reads can also return `pgcat`, `node_exporter` and `kube_state_metrics`, which the collector finds by itself and the API doesn't accept on create. Can't be changed after the target is created. Valid values: `postgres`, `pgbouncer`, `mysql`, `redis`, `mongodb`, `memcached`, `elasticsearch`, `nginx`, `apache`, `kafka`, `prometheus`, `traefik`.
- `host` (String) Target host. For the process kinds this is the host the collector runs on.

### Optional

- `enabled` (Boolean) Whether the collector scrapes this target. A new target is enabled unless it's set to `false`. Disabling it sets `status` to `disabled`, and enabling it again checks the target again. Defaults to `true`.
- `port` (Number) The target's port. `null` when a target is found by its `listen_ip` or `service` instead. Required for database kinds and accepted for `nginx`, `apache`, `kafka` and `traefik`. Applies only when `kind` is `postgres` or `pgbouncer` or `mysql` or `redis` or `mongodb` or `memcached` or `elasticsearch` or `nginx` or `apache` or `kafka` or `traefik`.
- `service` (String) The service name. Required for `nginx`, `apache`, `kafka`, `prometheus` and `traefik` targets, and accepted and returned only for those. Applies only when `kind` is `nginx` or `apache` or `kafka` or `prometheus` or `traefik`.
- `listen_ip` (String) The IP address the process listens on. Accepted and returned only for `nginx`, `apache`, `kafka` and `traefik`, and `null` when not set. Applies only when `kind` is `nginx` or `apache` or `kafka` or `traefik`.
- `endpoint` (String) The full scrape URL, like `http://10.0.0.5:9090/metrics`. Required for `prometheus` targets and returned only for them. Applies only when `kind` is `prometheus`.
- `scheme` (String) The connection scheme. Accepted and returned only for `elasticsearch`. Defaults to `http`. Valid values: `http`, `https`. Applies only when `kind` is `elasticsearch`.
- `team_id` (String) Team that owns the collector's source, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `username` (String) The username used to connect, or `null` when not set. Accepted and returned only for `postgres`, `pgbouncer`, `mysql`, `redis`, `mongodb`, `memcached` and `elasticsearch`. Applies only when `kind` is `postgres` or `pgbouncer` or `mysql` or `redis` or `mongodb` or `memcached` or `elasticsearch`.
- `password` (String, Sensitive) Password for database and `elasticsearch` targets. Better Stack never returns it, so Terraform keeps the value you set.
- `api_key` (String, Sensitive) API key for `elasticsearch` targets. Better Stack never returns it, so Terraform keeps the value you set.
- `ssl_mode` (String) The SSL mode for the connection. Required for `postgres` and `pgbouncer`, and accepted and returned only for those. Applies only when `kind` is `postgres` or `pgbouncer`.
  - `disable`: Connects without SSL.
  - `require`: Connects over SSL without verifying the server's certificate.
  - `verify-ca`: Connects over SSL and checks that a trusted certificate authority signed the server's certificate.
- `tls` (String) The TLS mode for the connection. Required for `mysql`, and accepted and returned only for it. Applies only when `kind` is `mysql`.
  - `false`: Connects without TLS.
  - `true`: Connects over TLS and verifies the server's certificate.
  - `skip-verify`: Connects over TLS without verifying the server's certificate.
  - `preferred`: Uses TLS when the server supports it, without verifying the server's certificate.

### Read-Only

- `id` (String) The ID of this collector target.
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

-> With an OAuth token, the user needs the sources permission.

## Import

Import an existing collector target by its ID as `collector_id/id`. Get the ID via [List collector targets](https://betterstack.com/docs/logs/api/list-collector-metrics-targets/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_collector_target.example
  id = "<collector_id>/12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_collector_target.example <collector_id>/12345
```
