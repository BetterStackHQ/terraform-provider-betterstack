---
page_title: "betterstack_metric_expression Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A rule that extracts a metric, or a label, from a source's logs using a SQL expression.
---

# betterstack_metric_expression (Resource)

A rule that extracts a metric, or a label, from a source's logs using a SQL expression.

-> Replaces `logtail_metric`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_metric_expression" "example" {
  source_id      = "12345"
  name           = "level"
  sql_expression = "JSONExtract(raw, 'level', 'Nullable(String)')"
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `source_id` (String, Required) Id of the parent this source belongs to. Changing it replaces the resource.

### Required

- `name` (String) Metric or label name.
- `sql_expression` (String) SQL expression evaluated for each log. Log fields are in the `raw` JSON field, so read them with `JSONExtract(raw, ..., 'Nullable(Type)')`.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `type` (String) Stored ClickHouse type. Valid values: `string`, `string_low_cardinality`, `int64_delta`, `int64`, `uint64_delta`, `uint64`, `float64_delta`, `datetime64_delta`, `boolean`.
- `aggregations` (List of String) Aggregations to precompute: `avg`, `count`, `uniq`, `max`, `min`, `sum` or `histogram`. An empty array makes this a label rather than a metric.
- `build_type` (String) Whether this change also applies to logs already stored. `new_data` applies the rule to new logs only. `historical_logs` also rebuilds the stored metrics, which is expensive. Valid values: `new_data`, `historical_logs`. Defaults to `new_data`. Write-only: the API never returns it; Terraform keeps the configured value.

### Read-Only

- `id` (String) The id of this metric_expression.

## Endpoints

- `GET /api/v2/sources/:source_id/metrics`
- `POST /api/v2/sources/:source_id/metrics`
- `PATCH /api/v2/sources/:source_id/metrics/:id`
- `DELETE /api/v2/sources/:source_id/metrics/:id`

-> Metrics and labels share this resource: a rule with aggregations is a metric, and one without is a label.

## Import

```shell
terraform import betterstack_metric_expression.example <source_id>/12345
```
