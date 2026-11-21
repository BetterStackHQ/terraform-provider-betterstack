---
page_title: "betterstack_dashboard_chart Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A chart on a dashboard, backed by one or more queries.
---

# betterstack_dashboard_chart (Data Source)

A chart on a dashboard, backed by one or more queries.

## Example Usage

```terraform
data "betterstack_dashboard_chart" "example" {
  dashboard_id = "12345"
  name         = "CPU Usage"
}
```

## Schema

### Address

- `dashboard_id` (String, Required) Dashboard this chart is on. Can't be changed after the chart is created. Changing it replaces the resource.

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Chart name.

### Read-Only

- `chart_type` (String) Chart type. Each type takes only certain query types, and only `line_chart` takes more than one query. An update can't change it to a type the chart's existing queries or alerts don't work with. Valid values: `line_chart`, `bar_chart`, `pie_chart`, `scatter_chart`, `gauge_chart`, `number_chart`, `heatmap_chart`, `table_chart`, `map_chart`, `text_chart`, `tail_chart`, `static_text_chart`, `funnel_chart`, `funnel_trend_chart`, `anomalies_chart`, `hierarchy_chart`.
- `description` (String) Chart description.
- `x` (Number) Grid x position. The grid is 12 columns wide. If left out, the chart goes in the first free space that fits.
- `y` (Number) Grid y position. Set automatically when left out.
- `w` (Number) Grid width.
- `h` (Number) Grid height.
- `settings` (Dynamic) Display settings for the chart type, such as unit, axes, legend, stacking and decimal places. Defaults to an empty object.
- `queries` (Dynamic, see [nested schema](#nestedatt--queries)) The chart's queries. You can write `sql_expression`, `tail_query` and `static_text` queries. Other types, such as `query_builder`, `pql_expression` and PromQL, are returned but can't be written. Every query has `id`, `query_type`, `name` and `source_variable`, and its other fields depend on its type. Updating a chart replaces all of its queries.
- `queries_result` (Dynamic, see [nested schema](#nestedatt--queries_result)) Complete server result for `queries`, including computed fields.

<a id="nestedatt--queries"></a>
### Nested Schema for `queries`

- `id` (String)
- `query_type` (String)
- `name` (String)
- `source_variable` (String)
- `sql_query` (String)
- `pql_query` (String) PQL source of a `pql_expression` query. Read-only: PQL queries can't be written through the API.
- `where_condition` (String)
- `static_text` (String)
- `funnel_query` (Dynamic) Funnel definition of a `funnel_query` query. Read-only.
- `x_axis` (Attributes List, see [nested schema](#nestedatt--queries--x_axis)) X-axis blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `y_axis` (Attributes List, see [nested schema](#nestedatt--queries--y_axis)) Y-axis blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `point_size` (Attributes List, see [nested schema](#nestedatt--queries--point_size)) Point size blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `coordinates` (Attributes List, see [nested schema](#nestedatt--queries--coordinates)) Coordinate blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `filters` (Attributes List, see [nested schema](#nestedatt--queries--filters)) Filter blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `where` (Attributes List, see [nested schema](#nestedatt--queries--where)) Where blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `group_by` (Attributes List, see [nested schema](#nestedatt--queries--group_by)) Grouping blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `having` (Attributes List, see [nested schema](#nestedatt--queries--having)) Having blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `order_by` (Attributes List, see [nested schema](#nestedatt--queries--order_by)) Ordering blocks of a `query_builder` or `tail_query_builder` query. Read-only.

<a id="nestedatt--queries--x_axis"></a>
### Nested Schema for `queries.x_axis`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--y_axis"></a>
### Nested Schema for `queries.y_axis`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--point_size"></a>
### Nested Schema for `queries.point_size`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--coordinates"></a>
### Nested Schema for `queries.coordinates`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--filters"></a>
### Nested Schema for `queries.filters`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--where"></a>
### Nested Schema for `queries.where`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--group_by"></a>
### Nested Schema for `queries.group_by`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--having"></a>
### Nested Schema for `queries.having`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries--order_by"></a>
### Nested Schema for `queries.order_by`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result"></a>
### Nested Schema for `queries_result`

- `id` (String)
- `query_type` (String)
- `name` (String)
- `source_variable` (String)
- `sql_query` (String)
- `pql_query` (String) PQL source of a `pql_expression` query. Read-only: PQL queries can't be written through the API.
- `where_condition` (String)
- `static_text` (String)
- `funnel_query` (Dynamic) Funnel definition of a `funnel_query` query. Read-only.
- `x_axis` (Attributes List, see [nested schema](#nestedatt--queries_result--x_axis)) X-axis blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `y_axis` (Attributes List, see [nested schema](#nestedatt--queries_result--y_axis)) Y-axis blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `point_size` (Attributes List, see [nested schema](#nestedatt--queries_result--point_size)) Point size blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `coordinates` (Attributes List, see [nested schema](#nestedatt--queries_result--coordinates)) Coordinate blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `filters` (Attributes List, see [nested schema](#nestedatt--queries_result--filters)) Filter blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `where` (Attributes List, see [nested schema](#nestedatt--queries_result--where)) Where blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `group_by` (Attributes List, see [nested schema](#nestedatt--queries_result--group_by)) Grouping blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `having` (Attributes List, see [nested schema](#nestedatt--queries_result--having)) Having blocks of a `query_builder` or `tail_query_builder` query. Read-only.
- `order_by` (Attributes List, see [nested schema](#nestedatt--queries_result--order_by)) Ordering blocks of a `query_builder` or `tail_query_builder` query. Read-only.

<a id="nestedatt--queries_result--x_axis"></a>
### Nested Schema for `queries_result.x_axis`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--y_axis"></a>
### Nested Schema for `queries_result.y_axis`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--point_size"></a>
### Nested Schema for `queries_result.point_size`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--coordinates"></a>
### Nested Schema for `queries_result.coordinates`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--filters"></a>
### Nested Schema for `queries_result.filters`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--where"></a>
### Nested Schema for `queries_result.where`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--group_by"></a>
### Nested Schema for `queries_result.group_by`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--having"></a>
### Nested Schema for `queries_result.having`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

<a id="nestedatt--queries_result--order_by"></a>
### Nested Schema for `queries_result.order_by`

- `explore_block_type` (String)
- `value` (String)
- `type` (String)
- `name` (String)
- `measure` (String)
- `value_type` (String)
- `operator` (String)
- `operator_symbol` (String)
- `condition_value` (String)
- `series_name` (String)
- `logical_operator` (String)
- `direction` (String)
- `interval_seconds` (Number)
- `kind` (String)
- `schema` (String)

## Endpoints

- `GET /api/v2/dashboards/:dashboard_id/charts`
- `GET /api/v2/dashboards/:dashboard_id/charts/:id`
- `POST /api/v2/dashboards/:dashboard_id/charts`
- `PATCH /api/v2/dashboards/:dashboard_id/charts/:id`
- `DELETE /api/v2/dashboards/:dashboard_id/charts/:id`
