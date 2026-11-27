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

- `id` (String) The ID of the connection to look up.

### Read-Only

- `client_type` (String) Client protocol. Valid values: `clickhouse`, `grafana`.
- `cluster` (String) Cluster a Grafana connection points at, by its name. [Get a source](https://betterstack.com/docs/logs/api/get-a-single-source/) returns the `cluster` its data is stored on.
- `clusters` (Attributes List) Every cluster the credential reaches, each with its `name` and the `host` to connect to. (see [below for nested schema](#nestedatt--clusters))
- `team_ids` (List of String) Teams whose data this connection reaches, as `t`-prefixed team IDs such as `t123`. Get the IDs via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).
- `team_names` (List of String) Names of the teams whose data this connection reaches.
- `ip_allowlist` (List of String) Allowed source IPs or CIDRs.
- `valid_until` (String (RFC 3339)) When the credential expires, or `null` if it never does. Leaving it out clears it, on create too.
- `note` (String) Free-form note.
- `host` (String) Connection host.
- `port` (Number) Connection port.
- `username` (String) The username to connect with.
- `password` (String, Sensitive) Connection password. Returned only when the connection is created and never again, so save it then.
- `created_by` (Attributes) User who created the connection in the web UI. Always `null` for a connection created through the API or MCP. (see [below for nested schema](#nestedatt--created_by))
- `sample_query` (String, Sensitive) An example `curl` query to get started. The response that creates the connection includes its password here, and later reads show `INSERT_PASSWORD` instead. Not returned when listing connections.
- `data_sources` (Attributes List) Sources this connection can query. Not returned when listing connections. (see [below for nested schema](#nestedatt--data_sources))
- `created_at` (String (RFC 3339)) When the connection was created.

<a id="nestedatt--clusters"></a>
### Nested Schema for `clusters`

Read-Only:

- `name` (String) The cluster's name.
- `host` (String) Host to connect to the cluster on.

<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

Read-Only:

- `id` (String) The creator's user ID, such as `u123`.
- `email` (String) The creator's e-mail.

<a id="nestedatt--data_sources"></a>
### Nested Schema for `data_sources`

Read-Only:

- `source_id` (String) The source's ID.
- `source_name` (String) The source's name.
- `team_name` (String) The name of the team the source belongs to.
- `connection_url` (String) The URL to connect to for this source's data.
- `data_sources` (List of String) ClickHouse collection names you can query for this source.
