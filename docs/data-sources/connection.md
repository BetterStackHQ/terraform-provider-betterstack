---
page_title: "betterstack_connection Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A ClickHouse connection for querying ingested data directly over SQL.
---

# betterstack_connection (Data Source)

A ClickHouse connection for querying ingested data directly over SQL.

## Example Usage

```terraform
data "betterstack_connection" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the connection to look up.

### Read-Only

- `client_type` (String) Client protocol. Valid values: `clickhouse`, `grafana`.
- `data_region` (String) Cluster name or region ID to limit the connection to. Leave it out to reach every cluster where the connection's teams store data, which is returned as `null`. Otherwise it's returned as the cluster's storage backend name.
- `team_ids` (List of String) Teams whose data this connection reaches, as `t`-prefixed team IDs such as `t123`. Give either `team_ids` or `team_names`, not both.
- `team_names` (List of String) Names of the teams whose data this connection reaches. Give either `team_ids` or `team_names`, not both.
- `ip_allowlist` (List of String) Allowed source IPs or CIDRs.
- `valid_until` (String (RFC 3339)) When the credential expires.
- `note` (String) Free-form note.
- `host` (String) Connection host.
- `port` (Number) Connection port.
- `username` (String) Connection username.
- `password` (String, Sensitive) Connection password. Returned only when the connection is created and never again, so save it then.
- `created_by` (Attributes, see [nested schema](#nestedatt--created_by)) User who created the connection in the web UI. Always `null` for a connection created through the API or MCP.
- `sample_query` (String) An example query to get started. Not returned when listing connections.
- `data_sources` (Attributes List, see [nested schema](#nestedatt--data_sources)) Sources this connection can query. Not returned when listing connections.
- `created_at` (String (RFC 3339)) When the connection was created.

<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

- `id` (Number)
- `email` (String)

<a id="nestedatt--data_sources"></a>
### Nested Schema for `data_sources`

- `source_id` (Number)
- `source_name` (String)
- `team_name` (String)
- `connection_url` (String)
- `data_sources` (List of String) ClickHouse collection names exposed for this source.

## Endpoints

- `GET /api/v2/connections`
- `GET /api/v2/connections/:id`
- `POST /api/v2/connections`
- `DELETE /api/v2/connections/:id`
