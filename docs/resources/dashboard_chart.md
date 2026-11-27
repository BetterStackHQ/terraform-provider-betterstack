---
page_title: "betterstack_dashboard_chart Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A chart on a dashboard, backed by one or more queries.
---

# betterstack_dashboard_chart (Resource)

A chart on a dashboard, backed by one or more queries.

-> Replaces `logtail_dashboard_chart`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a dashboard chart

```terraform
resource "betterstack_dashboard_chart" "example" {
  dashboard_id = "1234"
  chart_type   = "line_chart"
  name         = "Log volume"
  x            = 0
  y            = 4
  queries      = [{ query_type = "sql_expression", sql_query = "SELECT {{time}} as time, count(*) as value FROM {{source}} WHERE dt BETWEEN {{start_time}} AND {{end_time}} GROUP BY time" }]
}
```

### Error count, red above 100

```terraform
resource "betterstack_dashboard_chart" "error_count_red_above_100" {
  dashboard_id = "1234"
  chart_type   = "number_chart"
  name         = "Errors"
  w            = 3
  settings     = { color_mode = "high_red", color_upper_threshold = 100, color_lower_threshold = 10 }
  queries      = [{ query_type = "sql_expression", sql_query = "SELECT count(*) AS value FROM {{source}} WHERE dt BETWEEN {{start_time}} AND {{end_time}} AND JSONExtract(raw, 'level', 'Nullable(String)') = 'error'" }]
}
```

### Runbook link as text

```terraform
resource "betterstack_dashboard_chart" "runbook_link_as_text" {
  dashboard_id = "1234"
  chart_type   = "static_text_chart"
  name         = "Runbook"
  x            = 0
  y            = 0
  w            = 12
  h            = 2
  queries      = [{ query_type = "static_text", static_text = "On call? Start with the [checkout runbook](https://wiki.example.com/runbooks/checkout)." }]
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `dashboard_id` (String, Required) Dashboard this chart is on. Can't be changed after the chart is created. Changing it replaces the resource.

### Required

- `chart_type` (String) Chart type. Only a `line_chart` can have more than one query. An update can't change it to a type the chart's existing queries or alerts don't work with. Changing it in an update without `queries` turns Drag & drop queries into `sql_expression` queries with the same SQL.
  - `line_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `bar_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `pie_chart`: Accepts `sql_expression` queries.
  - `scatter_chart`: Accepts `sql_expression` queries.
  - `gauge_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `number_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `heatmap_chart`: Accepts `sql_expression` queries.
  - `table_chart`: Accepts `sql_expression` queries.
  - `map_chart`: Accepts `sql_expression` queries.
  - `text_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `tail_chart`: Accepts `tail_query` queries.
  - `static_text_chart`: Accepts a `static_text` query.
  - `funnel_chart`: Holds a `funnel_query` query made in the app.
  - `funnel_trend_chart`: Holds a `funnel_query` query made in the app.
  - `anomalies_chart`: Accepts `tail_query` queries.
  - `hierarchy_chart`: Accepts `sql_expression` queries.
- `queries` (Attributes List) The chart's queries. You can write `sql_expression`, `tail_query` and `static_text` queries. Other types, such as `query_builder`, `pql_expression` and PromQL, are returned but can't be written. Updating a chart replaces all of its queries. A chart made or updated from `query` has the one query built from it. A chart other than `static_text_chart` needs at least one query, and only a `line_chart` can have more than one. (see [below for nested schema](#nestedatt--queries))

### Optional

- `name` (String) Chart name. Defaults to `Your chart`.
- `description` (String) Chart description.
- `team_id` (String) Team that owns the chart's dashboard, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `x` (Number, Computed) The chart's horizontal position on the 12-column grid, from `0` at the left.
- `y` (Number, Computed) The chart's vertical position on the grid, from `0` at the top.
- `w` (Number) Grid width. Defaults to `6`.
- `h` (Number) Grid height. Defaults to `4`.
- `settings` (Attributes, Computed) Display settings for the `chart_type`, such as unit, axes, legend, stacking and decimal places. An update replaces all of them. Switching `x_axis_type` to or from `category` in an update without `queries` turns the chart's Drag & drop queries into `sql_expression` queries with the same SQL. Defaults to an empty object. (see [below for nested schema](#nestedatt--settings))
- `explanation` (String) Explanation of what the chart shows, which AI SRE and the chart's AI summary read.

### Read-Only

- `id` (String) The ID of this dashboard chart.

<a id="nestedatt--queries"></a>
### Nested Schema for `queries`

Optional:

- `query_type` (String) The query's type. You can write `sql_expression`, `tail_query` and `static_text` queries. The types the editor builds are read-only. They're `pql_expression`, `query_builder`, `tail_query_builder`, `funnel_query`, `promql_query`, `promql_v2_query` and `promql_via_llm`.
- `name` (String) The query's name. Left out when the query has none.
- `source_variable` (String) The sources the query reads, as the name of a `source` variable, such as `source`, or a `source:` reference to specific sources. Defaults to `source`.
- `sql_query` (String) The SQL the query runs. Returned for `sql_expression`, `pql_expression`, `query_builder`, `tail_query_builder` and `funnel_query` queries.
- `pql_query` (String) PQL source of a `pql_expression` query. Read-only.
- `where_condition` (String) The filter of a `tail_query` or `tail_query_builder` query. Left out when not set.
- `static_text` (String) The Markdown text a `static_text` query shows. Left out when not set.
- `funnel_query` (Attributes) Funnel definition of a `funnel_query` query, with the session field, the steps, each a name and its conditions, and an optional breakdown. Read-only. (see [below for nested schema](#nestedatt--queries--funnel_query))
- `x_axis` (Attributes List) X-axis blocks of a `query_builder` or `tail_query_builder` query, as an empty list when there are none. Read-only. (see [below for nested schema](#nestedatt--queries--x_axis))
- `y_axis` (Attributes List) Y-axis blocks of a `query_builder` or `tail_query_builder` query. Left out when not set. Read-only. (see [below for nested schema](#nestedatt--queries--y_axis))
- `point_size` (Attributes List) Point size blocks of a `query_builder` or `tail_query_builder` query. Left out when not set. Read-only. (see [below for nested schema](#nestedatt--queries--point_size))
- `coordinates` (Attributes List) Coordinate blocks of a `query_builder` or `tail_query_builder` query, as an empty list when there are none. Read-only. (see [below for nested schema](#nestedatt--queries--coordinates))
- `filters` (Attributes List) Filter blocks of a `query_builder` or `tail_query_builder` query. Left out when not set. Read-only. (see [below for nested schema](#nestedatt--queries--filters))
- `where` (Attributes List) Where blocks of a `query_builder` or `tail_query_builder` query, as an empty list when there are none. Read-only. (see [below for nested schema](#nestedatt--queries--where))
- `group_by` (Attributes List) Grouping blocks of a `query_builder` or `tail_query_builder` query. Left out when not set. Read-only. (see [below for nested schema](#nestedatt--queries--group_by))
- `having` (Attributes List) Having blocks of a `query_builder` or `tail_query_builder` query, as an empty list when there are none. Read-only. (see [below for nested schema](#nestedatt--queries--having))
- `order_by` (Attributes List) Ordering blocks of a `query_builder` or `tail_query_builder` query, as an empty list when there are none. Read-only. (see [below for nested schema](#nestedatt--queries--order_by))

Read-Only:

- `id` (String) The query's ID.

<a id="nestedatt--queries--funnel_query"></a>
### Nested Schema for `queries.funnel_query`

Optional:

- `session_field` (String) Field that ties one journey's events together, such as `session_id` or `user.id`.
- `steps` (Attributes List) The funnel's steps in journey order. (see [below for nested schema](#nestedatt--queries--funnel_query--steps))
- `breakdown` (Attributes) The field each step is split by, or `null` when the funnel has no breakdown. (see [below for nested schema](#nestedatt--queries--funnel_query--breakdown))

<a id="nestedatt--queries--funnel_query--steps"></a>
### Nested Schema for `queries.funnel_query.steps`

Optional:

- `name` (String) Step name shown on the chart.
- `conditions` (String) Live tail query that the step's events match, such as `payload.meta.path="/cart"`.

<a id="nestedatt--queries--funnel_query--breakdown"></a>
### Nested Schema for `queries.funnel_query.breakdown`

Optional:

- `type` (String) The type of the breakdown field, always `string`.
- `value` (String) The field each step is split by, such as `contexts.browser.name`.

<a id="nestedatt--queries--x_axis"></a>
### Nested Schema for `queries.x_axis`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--y_axis"></a>
### Nested Schema for `queries.y_axis`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--point_size"></a>
### Nested Schema for `queries.point_size`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--coordinates"></a>
### Nested Schema for `queries.coordinates`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--filters"></a>
### Nested Schema for `queries.filters`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--where"></a>
### Nested Schema for `queries.where`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--group_by"></a>
### Nested Schema for `queries.group_by`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--having"></a>
### Nested Schema for `queries.having`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--queries--order_by"></a>
### Nested Schema for `queries.order_by`

Optional:

- `explore_block_type` (String) What the block reads, such as `count`, `field`, `json_value`, `json_value_from_column`, `special_column`, `virtual_field` or `search`.
- `value` (String) The column or field the block reads, such as `dt`.
- `type` (String) The value's data type, such as `datetime`.
- `name` (String) The name the block is shown with. The editor shows `value` when it's empty.
- `measure` (String) The aggregation applied to the value, such as `avg`, `count`, `uniq`, `max`, `min`, `sum` or `p95`.
- `value_type` (String) `rate` to chart the value's rate of change.
- `operator` (String) The comparison of a filter block.
- `operator_symbol` (String) The filter block's comparison written as a symbol.
- `condition_value` (String) The value a filter block compares against.
- `series_name` (String) The name a Y-axis or coordinates block's series is shown as.
- `logical_operator` (String) How a filter block joins the filters before it, `and` or `or`. Defaults to `and`.
- `direction` (String) The sort direction of an order block, `asc` or `desc`.
- `interval_seconds` (Number) The time bucket size in seconds, for a time block on the X axis.
- `kind` (String) `metric` for an aggregated value, or `label` for a raw dimension.
- `schema` (String) The source's storage layout, `row_based` or `column_based`.

<a id="nestedatt--settings"></a>
### Nested Schema for `settings`

Optional:

- `unit` (String) Unit the values are shown in, such as `percent`, `bytes` or `ms`. Table charts use `column_units` instead.
- `custom_unit` (Boolean) Whether `unit` is your own suffix, shown after the value as written and never scaled.
- `decimal_places` (Number) Number of decimal places shown.
- `time_column` (String) Column on the X axis, `time` by default. For a category bar chart it's the grouping column.
- `x_axis_type` (String) How the X axis is read. Usually `time`, `value` or `category`.
- `y_axis_type` (String) How the Y axis of a scatter chart is read. Usually `value` or `category`.
- `series_column` (String) Column that splits the data into series. On a heatmap it's the Y axis column.
- `value_columns` (List of String) Result columns that hold the values to chart.
- `point_size_column` (String) Column that sets each point's size on a scatter chart.
- `location_column` (String) Column holding the location on a map chart.
- `location_label` (String) What a map chart's tooltip names, either the closest city or country or the value as sent. Usually `automatic` or `original_value`.
- `label_column` (String) Column that names each hexagon on a hierarchy chart.
- `group_by_columns` (List of String) Up to two columns a hierarchy chart groups its hexagons by.
- `palette` (String) Color scale of a hierarchy chart. Usually `green_to_red` or `red_to_green`.
- `min_value` (Number) Lowest value of a hierarchy chart's color scale.
- `max_value` (Number) Highest value of a hierarchy chart's color scale.
- `link_template` (String) Link a hierarchy chart opens when you click a hexagon, with `{{column}}` placeholders.
- `y_axis_min` (Number) Lowest value on the Y axis.
- `y_axis_max` (Number) Highest value on the Y axis.
- `y_axis_auto_range` (String) How a line chart fits its Y axis when no bounds are set. Usually `include_zero` or `fit_data`.
- `y_axis_scale` (String) Scale of the Y axis. Usually `linear` or `log`.
- `min` (Number) Lower bound of a gauge.
- `max` (Number) Upper bound of a gauge, the goal value.
- `treat_missing_values` (String) How gaps in the data are drawn. Usually `connected`, `disconnected`, `zero` or `previous`.
- `stacking` (String) How series are stacked. Usually `dont_stack`, `stack` or `stack_to_100`.
- `orientation` (String) Direction of the bars on a bar chart. Usually `vertical` or `horizontal`.
- `legend` (String) Where the legend is shown. Usually `hidden`, `shown_below`, `shown_below_with_statistics` or `shown_right`.
- `series_colors` (Map of String) Color of each series, keyed by its `series_name`, as a hex color such as `#22c55e`.
- `series_order` (List of String) Order the series are shown in, by name.
- `show_previous_period` (Boolean) Whether a line chart also shows the previous period.
- `color_mode` (String) How a number chart colors its value. Usually `auto`, `high_green`, `high_red` or `specific_words`.
- `color_upper_threshold` (Number) Upper threshold of a number chart's colors.
- `color_lower_threshold` (Number) Lower threshold of a number chart's colors.
- `color_green_words` (String) Values a number chart shows in green, separated by commas.
- `color_yellow_words` (String) Values a number chart shows in yellow, separated by commas.
- `color_red_words` (String) Values a number chart shows in red, separated by commas.
- `column_units` (Attributes Map) Unit of each table column, by column name. (see [below for nested schema](#nestedatt--settings--column_units))
- `percentage_highlight_columns` (List of String) Table columns holding ratios from 0 to 1, shown with a bar.
- `series_unit` (String) Unit of a heatmap's Y axis labels.
- `series_custom_unit` (Boolean) Whether `series_unit` is your own suffix.
- `heatmap_color_mode` (String) How a heatmap colors its cells. Usually `gradient` or `categorized`.
- `value_categories` (Attributes Map) Label and color of each cell value of a categorized heatmap. (see [below for nested schema](#nestedatt--settings--value_categories))
- `column_widths` (Map of Number) Width of each table column, by column name, as the chart editor saves it.
- `table_width` (Number) Width of the table the `column_widths` were saved for.
- `guessed_series_colors` (Map of String) Colors the chart picked for its series, keyed by `series_name`.
- `label` (String) A legend placement the chart editor saves next to `legend`, such as `shown_below`. Charts ignore it and place the legend by `legend`.
- `text` (String) Markdown text of a text chart.
- `row_templates` (Map of String) Live tail row format of each source, by source ID.
- `sampling` (Number) Percentage of data the chart samples.
- `fresh` (Boolean) Whether the chart skips cached results.
- `funnel_unit_name` (String) Word a funnel chart counts in, such as `people`.
- `filtered_series` (List of String) Series the chart shows, by name.
- `filtered_releases` (List of String) Releases the chart shows.
- `force_source_series` (Boolean) Whether the chart shows one series per source.
- `treat_zero_values` (String) How zero values are drawn, `keep` or `drop_disconnected`. Defaults to `keep`.
- `projection` (String) Map projection of a map chart, such as `mercator`.
- `published_format` (String) Format a published chart is served in, `json` by default.
- `published_cache_for` (Number) Seconds a published chart's results are cached, 3 by default.
- `y_axis_annotations` (Attributes List) Horizontal lines drawn at values on the Y axis, each with a label and color. (see [below for nested schema](#nestedatt--settings--y_axis_annotations))
- `x_axis_annotations` (Attributes List) Vertical lines drawn at values on the X axis, each with a label and color. A value is a time or a number, as the X axis reads it. (see [below for nested schema](#nestedatt--settings--x_axis_annotations))

<a id="nestedatt--settings--column_units"></a>
### Nested Schema for `settings.column_units`

Optional:

- `unit` (String) Unit the column's values are shown in, such as `auto`, `percent` or `bytes`.
- `custom_unit` (Boolean) Whether `unit` is your own suffix, shown after the value as written and never scaled.

<a id="nestedatt--settings--value_categories"></a>
### Nested Schema for `settings.value_categories`

Optional:

- `label` (String) Text shown for the value.
- `color` (String) Color of the cells with the value, as a hex color such as `#22c55e`.

<a id="nestedatt--settings--y_axis_annotations"></a>
### Nested Schema for `settings.y_axis_annotations`

Optional:

- `value` (Number) The Y axis value the line is drawn at.
- `label` (String) Text shown next to the line.
- `color` (String) The line's color, as a hex color such as `#22c55e`. Left out, the line is gray.

<a id="nestedatt--settings--x_axis_annotations"></a>
### Nested Schema for `settings.x_axis_annotations`

Optional:

- `value` (String) The X axis value the line is drawn at, a time or a number.
- `label` (String) Text shown next to the line.
- `color` (String) The line's color, as a hex color such as `#22c55e`. Left out, the line is gray.

## Import

Import an existing dashboard chart by its ID as `dashboard_id/id`. Get the ID from the [`betterstack_dashboard_chart` data source](../data-sources/dashboard_chart.md), which finds it by `name`. `terraform query` with the [`betterstack_dashboard_chart` list resource](../list-resources/dashboard_chart.md) lists them all, ready to import.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_dashboard_chart.example
  id = "<dashboard_id>/12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_dashboard_chart.example <dashboard_id>/12345
```
