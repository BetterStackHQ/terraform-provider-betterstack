---
page_title: "betterstack_extracted_metric Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A rule that extracts a metric, or a label, from a source's logs using its `sql_expression`.
---

# betterstack_extracted_metric (Data Source)

A rule that extracts a metric, or a label, from a source's logs using its `sql_expression`.

## Example Usage

```terraform
data "betterstack_extracted_metric" "example" {
  source_id = "12345"
  name      = "level"
}
```

## Schema

### Address

- `source_id` (String, Required) The ID of the parent this source belongs to. Changing it replaces the resource.

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Metric or label name.

### Optional

- `team_id` (String) Team that owns the metric's source, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `sql_expression` (String) SQL expression evaluated for each log. Log fields are in the `raw` JSON field, so read them with `JSONExtract(raw, ..., 'Nullable(Type)')`.
- `type` (String) Stored ClickHouse type. Valid values: `string`, `string_low_cardinality`, `int64_delta`, `int64`, `uint64_delta`, `uint64`, `float64_delta`, `datetime64_delta`, `boolean`.
- `aggregations` (List of String) Aggregations to precompute, `avg`, `count`, `uniq`, `max`, `min`, `sum` or `histogram`. An empty array makes this a label rather than a metric.
- `build_type` (String) Whether this change also applies to logs already stored.
  - `new_data`: Applies the rule to new logs only.
  - `historical_logs`: Also rebuilds the stored metrics from the retained logs, which is expensive.
