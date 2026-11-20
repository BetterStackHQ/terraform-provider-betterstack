---
page_title: "betterstack_exploration Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A saved exploration: a chart, its queries, and its variables over logs, metrics, or traces.
---

# betterstack_exploration (Data Source)

A saved exploration: a chart, its queries, and its variables over logs, metrics, or traces.

## Example Usage

```terraform
data "betterstack_exploration" "example" {
  exploration_group_id = 456
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Exploration name. The chart takes its name from it, so the chart's name is never sent.
- `exploration_group_id` (Number) Group this exploration belongs to. Send `null` or `"0"` to remove it from its group.

### Read-Only

- `team_name` (String) Name of the team that owns the exploration.
- `date_range_from` (String) Default range start, for example now-3h. Applied only when sent, so it can't be cleared.
- `date_range_to` (String) Default range end, for example now. Applied only when sent, so it can't be cleared.
- `chart` (Attributes, see [nested schema](#nestedatt--chart)) Chart settings. `chart_type` is required on create. The chart is always placed at `x: 0, y: 0, w: 12, h: 10`, and its name comes from the exploration's name. Returned without an `id`.
- `queries` (Dynamic, see [nested schema](#nestedatt--queries)) The chart's queries, in ID order. Updating an exploration replaces all of its queries. Each query takes only the keys its `query_type` uses: `sql_query` and `source_variable` for `sql_expression`, `where_condition` and `source_variable` for `tail_query`, and `static_text` for `static_text`. Sending a key another type uses, or more than one query on a chart other than `line_chart`, returns 422.
- `variables` (Dynamic, see [nested schema](#nestedatt--variables)) Template variables. An update replaces all variables: any you don't send is removed, except the built-in time, start_time, end_time and source variables, which always exist. `sql_definition` applies only to `select_with_sql` and `multi_select_with_sql` variables. `allow_multiple_values` is saved only for `select_value` and `select_with_sql` variables. A source variable's values are IDs of sources the team can access.
- `created_at` (String (RFC 3339)) When the exploration was created.
- `updated_at` (String (RFC 3339)) When the exploration was last updated.
- `queries_result` (Dynamic, see [nested schema](#nestedatt--queries_result)) Complete server result for `queries`, including computed fields.

<a id="nestedatt--chart"></a>
### Nested Schema for `chart`

- `chart_type` (String) Chart type. Each type renders only certain query types, for example `tail_chart` only a `tail_query` and `static_text_chart` only a `static_text` query. Valid values: `line_chart`, `table_chart`, `bar_chart`, `number_chart`, `heatmap_chart`, `anomalies_chart`, `pie_chart`, `scatter_chart`, `gauge_chart`, `tail_chart`, `static_text_chart`, `map_chart`, `text_chart`, `funnel_chart`, `funnel_trend_chart`, `hierarchy_chart`.
- `name` (String)
- `description` (String)
- `settings` (Dynamic)

<a id="nestedatt--queries"></a>
### Nested Schema for `queries`

- `id` (Number)
- `query_type` (String) Query type. Only `sql_expression`, `tail_query` and `static_text` can be written through the API. Other types are read-only: you can replace them with a writable type, never edit them. Valid values: `sql_expression`, `tail_query`, `static_text`, `pql_expression`, `tail_query_builder`, `query_builder`, `funnel_query`.
- `name` (String)
- `sql_query` (String)
- `pql_query` (String)
- `where_condition` (String)
- `static_text` (String)
- `source_variable` (String)
- `x_axis` (Dynamic)
- `y_axis` (Dynamic)
- `point_size` (Dynamic)
- `coordinates` (Dynamic)
- `filters` (Dynamic)
- `where` (Dynamic)
- `group_by` (Dynamic)
- `having` (Dynamic)
- `order_by` (Dynamic)
- `funnel_query` (Dynamic)

<a id="nestedatt--variables"></a>
### Nested Schema for `variables`

- `name` (String)
- `variable_type` (String) Variable kind. Valid values: `source`, `date`, `datetime`, `string`, `number`, `sql_expression`, `select_value`, `boolean`, `select_with_sql`, `multi_select_with_sql`, `select_predefined_sql`.
- `values` (List of String)
- `default_values` (List of String)
- `sql_definition` (String)
- `structured_value` (Dynamic)
- `allow_multiple_values` (Boolean)

<a id="nestedatt--queries_result"></a>
### Nested Schema for `queries_result`

- `id` (Number)
- `query_type` (String) Query type. Only `sql_expression`, `tail_query` and `static_text` can be written through the API. Other types are read-only: you can replace them with a writable type, never edit them. Valid values: `sql_expression`, `tail_query`, `static_text`, `pql_expression`, `tail_query_builder`, `query_builder`, `funnel_query`.
- `name` (String)
- `sql_query` (String)
- `pql_query` (String)
- `where_condition` (String)
- `static_text` (String)
- `source_variable` (String)
- `x_axis` (Dynamic)
- `y_axis` (Dynamic)
- `point_size` (Dynamic)
- `coordinates` (Dynamic)
- `filters` (Dynamic)
- `where` (Dynamic)
- `group_by` (Dynamic)
- `having` (Dynamic)
- `order_by` (Dynamic)
- `funnel_query` (Dynamic)

## Endpoints

- `GET /api/v2/explorations`
- `GET /api/v2/explorations/:id`
- `POST /api/v2/explorations`
- `PATCH /api/v2/explorations/:id`
- `DELETE /api/v2/explorations/:id`
