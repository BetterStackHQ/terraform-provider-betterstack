---
page_title: "betterstack_connection Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A ClickHouse connection for querying ingested data directly over SQL.
---

# betterstack_connection (Resource)

A ClickHouse connection for querying ingested data directly over SQL.

~> Better Stack can't update a connection in place, so changing any argument replaces it.

-> Replaces `logtail_connection`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a connection

```terraform
resource "betterstack_connection" "example" {
  client_type  = "clickhouse"
  team_ids     = ["t123"]
  ip_allowlist = ["192.168.1.0/24", "10.0.0.1"]
  valid_until  = "2030-12-31T00:00:00Z"
  note         = "Production connection for analytics"
}
```

### Connect Grafana to one cluster

```terraform
resource "betterstack_connection" "connect_grafana_to_one_cluster" {
  client_type = "grafana"
  cluster     = "eu-nbg-2"
  team_ids    = ["t123"]
  valid_until = "2030-12-31T00:00:00Z"
  note        = "Grafana Cloud data source"
}
```

## Schema

### Required

- `client_type` (String, Forces replacement) Client protocol. Valid values: `clickhouse`, `grafana`.
- `team_ids` (List of String, Forces replacement) Teams whose data this connection reaches, as `t`-prefixed team IDs such as `t123`. Get the IDs via [List teams](https://betterstack.com/docs/uptime/api/teams/list/).

### Optional

- `cluster` (String, Forces replacement) Cluster a Grafana connection points at, by its name. [Get a source](https://betterstack.com/docs/logs/api/get-a-single-source/) returns the `cluster` its data is stored on. Required for a Grafana connection. Leave it out for a ClickHouse connection, which reaches every cluster where its teams store data and returns `null` here.
- `ip_allowlist` (List of String, Forces replacement) Allowed source IPs or CIDRs. Defaults to `[]`.
- `note` (String, Forces replacement) Free-form note.
- `valid_until` (String (RFC 3339), Forces replacement) When the credential expires, or `null` if it never does. Leaving it out clears it, on create too.

### Read-Only

- `id` (String) The ID of this connection.
- `clusters` (Attributes List) Every cluster the credential reaches, each with its `name` and the `host` to connect to. (see [below for nested schema](#nestedatt--clusters))
- `team_names` (List of String) Names of the teams whose data this connection reaches.
- `host` (String) Connection host.
- `port` (Number) Connection port.
- `username` (String) The username to connect with.
- `password` (String, Sensitive) Connection password. Returned only when the connection is created and never again, so save it then. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `created_by` (Attributes) User who created the connection in the web UI. Always `null` for a connection created through the API or MCP. (see [below for nested schema](#nestedatt--created_by))
- `sample_query` (String, Sensitive) An example `curl` query to get started. The response that creates the connection includes its password here, and later reads show `INSERT_PASSWORD` instead. Not returned when listing connections. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `data_sources` (Attributes List) Sources this connection can query. Not returned when listing connections. (see [below for nested schema](#nestedatt--data_sources))
- `created_at` (String (RFC 3339)) When the connection was created.

<a id="nestedatt--clusters"></a>
### Nested Schema for `clusters`

Optional:

- `name` (String) The cluster's name.
- `host` (String) Host to connect to the cluster on.

<a id="nestedatt--created_by"></a>
### Nested Schema for `created_by`

Optional:

- `id` (String) The creator's user ID, such as `u123`.
- `email` (String) The creator's e-mail.

<a id="nestedatt--data_sources"></a>
### Nested Schema for `data_sources`

Optional:

- `source_id` (String) The source's ID.
- `source_name` (String) The source's name.
- `team_name` (String) The name of the team the source belongs to.
- `connection_url` (String) The URL to connect to for this source's data.
- `data_sources` (List of String) ClickHouse collection names you can query for this source.

-> Error tracking has its own connections, separate from these.

-> With an OAuth token, the user sees and can grant only the teams where their role has the `sql_api` permission. Deleting a connection needs that permission in every team it reaches.

## Import

Import an existing connection by its ID.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_connection.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_connection.example 12345
```
