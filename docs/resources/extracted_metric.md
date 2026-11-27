---
page_title: "betterstack_extracted_metric Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A rule that extracts a metric, or a label, from a source's logs using its `sql_expression`.
---

# betterstack_extracted_metric (Resource)

A rule that extracts a metric, or a label, from a source's logs using its `sql_expression`.

-> Replaces `logtail_metric`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a metric

```terraform
resource "betterstack_extracted_metric" "example" {
  source_id      = "123"
  name           = "duration_ms"
  sql_expression = "JSONExtract(raw, 'message_json', 'duration_ms', 'Nullable(Float64)')"
  aggregations   = ["avg", "histogram", "min", "max"]
}
```

### Create a label

```terraform
resource "betterstack_extracted_metric" "example" {
  source_id      = "123"
  name           = "log_type"
  sql_expression = "JSONExtractString(raw, 'message_json', 'log_type')"
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `source_id` (String, Required) The ID of the parent this source belongs to. Changing it replaces the resource.

### Required

- `name` (String) Metric or label name.
- `sql_expression` (String) SQL expression evaluated for each log. Log fields are in the `raw` JSON field, so read them with `JSONExtract(raw, ..., 'Nullable(Type)')`.

### Optional

- `team_id` (String) Team that owns the metric's source, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `type` (String, Computed) Stored ClickHouse type. Valid values: `string`, `string_low_cardinality`, `int64_delta`, `int64`, `uint64_delta`, `uint64`, `float64_delta`, `datetime64_delta`, `boolean`.
- `aggregations` (List of String, Computed) Aggregations to precompute, `avg`, `count`, `uniq`, `max`, `min`, `sum` or `histogram`. An empty array makes this a label rather than a metric.
- `build_type` (String) Whether this change also applies to logs already stored. Defaults to `new_data`. Better Stack never returns it, so Terraform keeps the value you set.
  - `new_data`: Applies the rule to new logs only.
  - `historical_logs`: Also rebuilds the stored metrics from the retained logs, which is expensive.

### Read-Only

- `id` (String) The ID of this extracted metric.

-> Metrics and labels share this resource. A rule with aggregations is a metric, and one without is a label.

## Import

Import an existing extracted metric by its ID as `source_id/id`. Get the ID from the [`betterstack_extracted_metric` data source](../data-sources/extracted_metric.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_extracted_metric.example
  id = "<source_id>/12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_extracted_metric.example <source_id>/12345
```
