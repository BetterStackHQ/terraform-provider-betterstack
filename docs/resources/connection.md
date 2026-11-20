---
page_title: "betterstack_connection Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A ClickHouse connection for querying ingested data directly over SQL.
---

# betterstack_connection (Resource)

A ClickHouse connection for querying ingested data directly over SQL.

~> The API has no update endpoint for this resource: changing any attribute forces a replacement.

-> Replaces `logtail_connection`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_connection" "example" {
  client_type = "clickhouse"
  team_ids    = ["t1234"]
}
```

## Schema

At least one of `team_ids` or `team_names` must be configured. The provider validates this condition before creating or updating the resource.

### Required

- `client_type` (String) Client protocol. Valid values: `clickhouse`, `grafana`. Changing it replaces the resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `data_region` (String) Cluster name or region ID to limit the connection to. Leave it out to reach every cluster where the connection's teams store data, which is returned as `null`. Otherwise it's returned as the cluster's storage backend name. Changing it replaces the resource.
- `team_ids` (List of String) Teams whose data this connection reaches, as `t`-prefixed team IDs such as `t123`. Give either `team_ids` or `team_names`, not both. Changing it replaces the resource.
- `team_names` (List of String) Names of the teams whose data this connection reaches. Give either `team_ids` or `team_names`, not both. Changing it replaces the resource.
- `ip_allowlist` (List of String) Allowed source IPs or CIDRs. Defaults to `[]`. Changing it replaces the resource.
- `valid_until` (String (RFC 3339)) When the credential expires. Changing it replaces the resource.
- `note` (String) Free-form note. Changing it replaces the resource.

### Read-Only

- `id` (String) The id of this connection.
- `host` (String) Connection host.
- `port` (Number) Connection port.
- `username` (String) Connection username.
- `password` (String, Sensitive) Connection password. Returned only when the connection is created and never again, so save it then. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
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

-> A team API token gets 401. Use a global API token or an OAuth token.

-> Error tracking has its own connections, served on its own host and separate from these. Requesting one product's connection on the other product's host returns 404.

-> With an OAuth token, the user sees and can grant only the teams where their role has the `sql_api` permission. Deleting a connection needs that permission in every team it reaches.

## Import

```shell
terraform import betterstack_connection.example 12345
```
