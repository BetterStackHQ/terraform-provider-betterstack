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

- `collector_id` (String, Required) Id of the parent this collector belongs to. Changing it replaces the resource.

### Required

- `id` (String) The id of the collector_target to look up.

### Read-Only

- `kind` (String) Target type. Can't be changed after the target is created. Valid values: `postgres`, `pgbouncer`, `mysql`, `redis`, `mongodb`, `memcached`, `elasticsearch`, `nginx`, `apache`, `kafka`, `prometheus`, `traefik`.
- `host` (String) Target host. For the process kinds this is the host the collector runs on.
- `enabled` (Boolean) Whether the collector scrapes this target. Disabling it sets `status` to `disabled`. Enabling it again checks the target again.
- `port` (Number) Target port (databases, nginx, apache, kafka, traefik).
- `service` (String) Service name (nginx, apache, kafka, traefik, prometheus).
- `listen_ip` (String) Listen IP (nginx, apache, kafka, traefik).
- `endpoint` (String) Scrape endpoint (prometheus).
- `scheme` (String) Scheme (elasticsearch). Valid values: `http`, `https`.
- `username` (String) Auth username (databases, elasticsearch).
- `password` (String, Sensitive) Auth password (databases, elasticsearch). Never returned by any read.
- `api_key` (String, Sensitive) API key (elasticsearch). Never returned by any read.
- `ssl_mode` (String) SSL mode (postgres, pgbouncer). Valid values: `disable`, `require`, `verify-ca`.
- `tls` (String) TLS mode (mysql). Valid values: `false`, `true`, `skip-verify`, `preferred`.
- `status` (String) Connection status. An update that sends `host`, `port`, `service`, `listen_ip`, `endpoint`, `scheme`, `username`, `password`, `api_key`, `ssl_mode` or `tls` sets it back to `pending`.
- `container` (String) Detected container.
- `created_at` (String (RFC 3339)) When the target was created.
- `updated_at` (String (RFC 3339)) When the target was last updated.

## Endpoints

- `GET /api/v1/collectors/:collector_id/targets`
- `GET /api/v1/collectors/:collector_id/targets/:id`
- `POST /api/v1/collectors/:collector_id/targets`
- `PATCH /api/v1/collectors/:collector_id/targets/:id`
- `DELETE /api/v1/collectors/:collector_id/targets/:id`
