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

```terraform
resource "betterstack_collector_target" "example" {
  collector_id = "12345"
  kind         = "postgres"
  host         = "db.example.com"
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `collector_id` (String, Required) Id of the parent this collector belongs to. Changing it replaces the resource.

### Required

- `kind` (String) Target type. Can't be changed after the target is created. Valid values: `postgres`, `pgbouncer`, `mysql`, `redis`, `mongodb`, `memcached`, `elasticsearch`, `nginx`, `apache`, `kafka`, `prometheus`, `traefik`. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `host` (String) Target host. For the process kinds this is the host the collector runs on.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `enabled` (Boolean) Whether the collector scrapes this target. Disabling it sets `status` to `disabled`. Enabling it again checks the target again. Defaults to `true`.
- `port` (Number) Target port (databases, nginx, apache, kafka, traefik).
- `service` (String) Service name (nginx, apache, kafka, traefik, prometheus).
- `listen_ip` (String) Listen IP (nginx, apache, kafka, traefik).
- `endpoint` (String) Scrape endpoint (prometheus).
- `scheme` (String) Scheme (elasticsearch). Valid values: `http`, `https`.
- `username` (String) Auth username (databases, elasticsearch).
- `password` (String, Sensitive) Auth password (databases, elasticsearch). Never returned by any read. Write-only: the API never returns it; Terraform keeps the configured value.
- `api_key` (String, Sensitive) API key (elasticsearch). Never returned by any read. Write-only: the API never returns it; Terraform keeps the configured value.
- `ssl_mode` (String) SSL mode (postgres, pgbouncer). Valid values: `disable`, `require`, `verify-ca`.
- `tls` (String) TLS mode (mysql). Valid values: `false`, `true`, `skip-verify`, `preferred`.

### Read-Only

- `id` (String) The id of this collector_target.
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

-> Routed on telemetry.betterstack.com only, like its parent collector. A collector outside the token's teams returns 404.

-> With an OAuth token, the user needs the sources permission.

## Import

```shell
terraform import betterstack_collector_target.example <collector_id>/12345
```
