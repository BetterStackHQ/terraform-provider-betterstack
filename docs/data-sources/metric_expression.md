---
page_title: "betterstack_metric_expression Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A rule that extracts a metric, or a label, from a source's logs using a SQL expression.
---

# betterstack_metric_expression (Data Source)

A rule that extracts a metric, or a label, from a source's logs using a SQL expression.

## Example Usage

```terraform
data "betterstack_metric_expression" "example" {
  source_id = "12345"
  name      = "level"
}
```

## Schema

### Address

- `source_id` (String, Required) Id of the parent this source belongs to. Changing it replaces the resource.

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Metric or label name.

### Read-Only

- `sql_expression` (String) SQL expression evaluated for each log. Log fields are in the `raw` JSON field, so read them with `JSONExtract(raw, ..., 'Nullable(Type)')`.
- `type` (String) Stored ClickHouse type. Valid values: `string`, `string_low_cardinality`, `int64_delta`, `int64`, `uint64_delta`, `uint64`, `float64_delta`, `datetime64_delta`, `boolean`.
- `aggregations` (List of String) Aggregations to precompute: `avg`, `count`, `uniq`, `max`, `min`, `sum` or `histogram`. An empty array makes this a label rather than a metric.
- `build_type` (String) Whether this change also applies to logs already stored. `new_data` applies the rule to new logs only. `historical_logs` also rebuilds the stored metrics, which is expensive. Valid values: `new_data`, `historical_logs`.

## Endpoints

- `GET /api/v2/sources/:source_id/metrics`
- `POST /api/v2/sources/:source_id/metrics`
- `PATCH /api/v2/sources/:source_id/metrics/:id`
- `DELETE /api/v2/sources/:source_id/metrics/:id`
