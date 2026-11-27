---
page_title: "betterstack_exploration Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A saved exploration, which is a chart with its queries and variables over logs, metrics, or traces.
---

# betterstack_exploration (Data Source)

A saved exploration, which is a chart with its queries and variables over logs, metrics, or traces.

## Example Usage

```terraform
data "betterstack_exploration" "example" {
  exploration_group_id = "456"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Exploration name. While it's still the default, a name generated from the query is returned. The chart is named after it, so the chart's name is never sent.
- `exploration_group_id` (String) Group this exploration belongs to.
- `product` (String) The Explore the exploration lives in. Leave it out on create to put the exploration in Telemetry's Explore. A funnel over web events goes in RUM's Explore instead. Changing it moves the exploration to the other Explore.
  - `telemetry`: Telemetry's Explore.
  - `rum`: RUM's Explore, which reads an application's web events only.

### Optional

- `team_id` (String) Team that owns the exploration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `date_range_from` (String) Start of the default date range, such as `now-3h`. A funnel's defaults to `now-7d`. It's applied only when sent, so it can't be cleared.
- `date_range_to` (String) Default range end, for example now. Applied only when sent, so it can't be cleared.
- `chart` (Attributes) Chart settings. `chart_type` is required on create. A funnel's `settings` accept `funnel_unit_name`, the word the chart counts in. The chart is always placed at `x: 0, y: 0, w: 12, h: 10`, and its name comes from the exploration's name. Returned without an `id`. (see [below for nested schema](#nestedatt--chart))
- `queries` (Attributes List) The chart's queries, in ID order. Updating an exploration replaces all of its queries. Each query accepts only the keys its `query_type` uses. A `sql_expression` query uses `sql_query` and `source_variable`. A `tail_query` uses `where_condition` and `source_variable`. A `static_text` query uses `static_text`, and a `funnel_query` uses `funnel_query`. A key another type uses, or more than one query on a chart other than `line_chart`, is refused. (see [below for nested schema](#nestedatt--queries))
- `variables` (Attributes List) Template variables. An update replaces all variables. Any the list leaves out is removed, except the built-in `time`, `start_time`, `end_time` and `source` variables, which always exist. A listed variable without `default_values`, `sql_definition` or `structured_value` keeps the stored ones while its `variable_type` stays the same, and a new `variable_type` clears them. One sent without `values` gets empty `values`, except a `source` variable, which keeps its sources while it leaves out both `values` and `structured_value`. One sent without `allow_multiple_values` keeps the stored flag. `sql_definition` applies only to `select_with_sql` and `multi_select_with_sql` variables. `allow_multiple_values` is saved only for `select_value` and `select_with_sql` variables. When a variable is saved with empty `values`, a `select_value` gets the first of its `default_values`, a `string`, `number`, `boolean` or `sql_expression` gets all of them, and a `select_with_sql` whose `sql_definition` is new or changed gets the first option it returns. An ISO 8601 date in a `date` or `datetime` variable is saved in `values` as a Unix timestamp in microseconds. A new or changed `sql_definition` or filter `sql` runs against the selected source, and the change is refused when the query fails. A `source` variable's values are IDs of sources the team can access. In a `rum` exploration, pick each source as its web events, `<table_name>:web_events`. A funnel can't have variables. Its source is set in `source_id`. See [Query variables](https://betterstack.com/docs/logs/dashboards/variables/). (see [below for nested schema](#nestedatt--variables))
- `source_id` (String) ID of the source a funnel reads. Required when creating a funnel, and accepted only then. Other explorations pick their sources in the `source` variable.
- `source_type` (String) The events a funnel reads from its source. When it's left out or the source doesn't have them, the funnel reads web events, logs or spans, the first of them the source has. Valid values: `web_events`, `logs`, `spans`.
- `created_at` (String (RFC 3339)) When the exploration was created.
- `updated_at` (String (RFC 3339)) When the exploration was last updated.
- `team_name` (String) Name of the team that owns the exploration.

<a id="nestedatt--chart"></a>
### Nested Schema for `chart`

Read-Only:

- `chart_type` (String) Chart type. A funnel chart can switch only to the funnel trend chart, its trend over time, and back.
  - `line_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `table_chart`: Accepts `sql_expression` queries.
  - `bar_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `number_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `heatmap_chart`: Accepts `sql_expression` queries.
  - `anomalies_chart`: Accepts `tail_query` queries.
  - `pie_chart`: Accepts `sql_expression` queries.
  - `scatter_chart`: Accepts `sql_expression` queries.
  - `gauge_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `tail_chart`: Accepts `tail_query` queries.
  - `static_text_chart`: Accepts a `static_text` query.
  - `map_chart`: Accepts `sql_expression` queries.
  - `text_chart`: Accepts `sql_expression` and `tail_query` queries.
  - `funnel_chart`: Accepts a `funnel_query` query.
  - `funnel_trend_chart`: Accepts a `funnel_query` query.
  - `hierarchy_chart`: Accepts `sql_expression` queries.
- `name` (String) The chart's name as stored. `name` returns a generated name in its place while it's the default.
- `description` (String) The chart's description, or `null` when it has none.
- `settings` (Attributes) Display settings for the chart type, such as unit, axes, legend, stacking and decimal places. An update replaces all of them. (see [below for nested schema](#nestedatt--chart--settings))

<a id="nestedatt--chart--settings"></a>
### Nested Schema for `chart.settings`

Read-Only:

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
- `column_units` (Attributes Map) Unit of each table column, by column name. (see [below for nested schema](#nestedatt--chart--settings--column_units))
- `percentage_highlight_columns` (List of String) Table columns holding ratios from 0 to 1, shown with a bar.
- `series_unit` (String) Unit of a heatmap's Y axis labels.
- `series_custom_unit` (Boolean) Whether `series_unit` is your own suffix.
- `heatmap_color_mode` (String) How a heatmap colors its cells. Usually `gradient` or `categorized`.
- `value_categories` (Attributes Map) Label and color of each cell value of a categorized heatmap. (see [below for nested schema](#nestedatt--chart--settings--value_categories))
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
- `y_axis_annotations` (Attributes List) Horizontal lines drawn at values on the Y axis, each with a label and color. (see [below for nested schema](#nestedatt--chart--settings--y_axis_annotations))
- `x_axis_annotations` (Attributes List) Vertical lines drawn at values on the X axis, each with a label and color. A value is a time or a number, as the X axis reads it. (see [below for nested schema](#nestedatt--chart--settings--x_axis_annotations))

<a id="nestedatt--chart--settings--column_units"></a>
### Nested Schema for `chart.settings.column_units`

Read-Only:

- `unit` (String) Unit the column's values are shown in, such as `auto`, `percent` or `bytes`.
- `custom_unit` (Boolean) Whether `unit` is your own suffix, shown after the value as written and never scaled.

<a id="nestedatt--chart--settings--value_categories"></a>
### Nested Schema for `chart.settings.value_categories`

Read-Only:

- `label` (String) Text shown for the value.
- `color` (String) Color of the cells with the value, as a hex color such as `#22c55e`.

<a id="nestedatt--chart--settings--y_axis_annotations"></a>
### Nested Schema for `chart.settings.y_axis_annotations`

Read-Only:

- `value` (Number) The Y axis value the line is drawn at.
- `label` (String) Text shown next to the line.
- `color` (String) The line's color, as a hex color such as `#22c55e`. Left out, the line is gray.

<a id="nestedatt--chart--settings--x_axis_annotations"></a>
### Nested Schema for `chart.settings.x_axis_annotations`

Read-Only:

- `value` (String) The X axis value the line is drawn at, a time or a number.
- `label` (String) Text shown next to the line.
- `color` (String) The line's color, as a hex color such as `#22c55e`. Left out, the line is gray.

<a id="nestedatt--queries"></a>
### Nested Schema for `queries`

Read-Only:

- `id` (String) The query's ID.
- `query_type` (String) Query type. You can replace a read-only query with a writable one, never edit it. Reads can also return `promql_query`, `promql_v2_query` and `promql_via_llm` for queries made in the app.
  - `sql_expression`: A query written in SQL. You can write it through the API.
  - `tail_query`: The filter of a tail chart. You can write it through the API.
  - `static_text`: The text of a `static_text_chart`. You can write it through the API.
  - `pql_expression`: A query written in PQL. It's read-only.
  - `tail_query_builder`: A tail chart's query made in the query builder. It's read-only.
  - `query_builder`: A query made in the query builder. It's read-only.
  - `funnel_query`: A funnel's one query. You can write it through the API.
- `name` (String) The query's name.
- `sql_query` (String) The SQL the query runs. Present for `sql_expression`, `pql_expression`, `query_builder`, `tail_query_builder` and `funnel_query` queries, and omitted when not set. Read-only except for `sql_expression`.
- `pql_query` (String) The PQL the query is written in. Present only for `pql_expression` queries.
- `where_condition` (String) The query's filter. Present only for `tail_query` and `tail_query_builder` queries, and omitted when not set.
- `static_text` (String) The query's Markdown content. Present only for `static_text` queries, and omitted when not set.
- `source_variable` (String) The sources the query reads, either the name of a `source` variable, such as `source`, or `source:` followed by specific sources. Defaults to `source` on create. Omitted when not set.
- `x_axis` (Attributes List) X-axis blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--x_axis))
- `y_axis` (Attributes List) Y-axis blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--y_axis))
- `point_size` (Attributes List) Point size blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--point_size))
- `coordinates` (Attributes List) Coordinate blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--coordinates))
- `filters` (Attributes List) Filter blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--filters))
- `where` (Attributes List) Where blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--where))
- `group_by` (Attributes List) Grouping blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--group_by))
- `having` (Attributes List) Having blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--having))
- `order_by` (Attributes List) Ordering blocks of a `query_builder` or `tail_query_builder` query. Read-only. (see [below for nested schema](#nestedatt--queries--order_by))
- `funnel_query` (Attributes) A funnel's steps, session field and breakdown. Present only for `funnel_query` queries. An update keeps the steps and fields it leaves out as they are. (see [below for nested schema](#nestedatt--queries--funnel_query))

<a id="nestedatt--queries--x_axis"></a>
### Nested Schema for `queries.x_axis`

Read-Only:

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

Read-Only:

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

Read-Only:

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

Read-Only:

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

Read-Only:

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

Read-Only:

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

Read-Only:

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

Read-Only:

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

Read-Only:

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

<a id="nestedatt--queries--funnel_query"></a>
### Nested Schema for `queries.funnel_query`

Read-Only:

- `steps` (Attributes List) The funnel's steps in journey order, 2 to 32. An update with `steps` replaces every step. (see [below for nested schema](#nestedatt--queries--funnel_query--steps))
- `session_field` (String) Field that ties one journey's events together. Defaults to `session_id`. Use `user.id` to follow an identified user. An empty value resets it to `session_id`.
- `breakdown_field` (String) Field each step is split by, such as `contexts.browser.name`. An empty value removes the breakdown.
- `breakdown` (Attributes) The breakdown the funnel is split by. Its field is in `value`. (see [below for nested schema](#nestedatt--queries--funnel_query--breakdown))

<a id="nestedatt--queries--funnel_query--steps"></a>
### Nested Schema for `queries.funnel_query.steps`

Read-Only:

- `name` (String) Step name shown on the chart. Defaults to `Step <n>`.
- `conditions` (String) Live tail query that the step's events match, such as `payload.meta.path="/cart"`.

<a id="nestedatt--queries--funnel_query--breakdown"></a>
### Nested Schema for `queries.funnel_query.breakdown`

Read-Only:

- `value` (String) The field each step is split by, such as `contexts.browser.name`.
- `type` (String) The type of the breakdown field, always `string`.

<a id="nestedatt--variables"></a>
### Nested Schema for `variables`

Read-Only:

- `name` (String) The variable's name, used as `{{name}}` in queries.
- `variable_type` (String) Variable kind. Valid values: `source`, `date`, `datetime`, `string`, `number`, `sql_expression`, `select_value`, `boolean`, `select_with_sql`, `multi_select_with_sql`, `select_predefined_sql`.
- `values` (List of String) The variable's current values as strings, or an empty array when it has none. For a `source` variable, these are the IDs of the selected sources.
- `default_values` (List of String) The variable's default values as strings, or an empty array when it has none.
- `sql_definition` (String) The SQL expression whose distinct values are the options of a `select_with_sql` variable, such as `JSONExtractString(raw, 'service')`. It runs against the variable's source over the selected range.
- `structured_value` (Attributes) The variable's other settings, such as the selected sources of a `source` variable in `canonicalSourceIds` and its services in `serviceIds`, or the saved filters of a `select_predefined_sql` variable. Left out when not set. (see [below for nested schema](#nestedatt--variables--structured_value))
- `allow_multiple_values` (Boolean) Whether a `select_value` or `select_with_sql` variable accepts more than one value at a time.
- `source_ids` (List of String) The IDs of the sources a `source` variable selects. They set the selection. Each must be a source the team can access.
- `filters` (Attributes List) The saved filters of a `select_predefined_sql` variable. Each has a `label` and an `sql` query, a complete query that returns one column with the filter's options. (see [below for nested schema](#nestedatt--variables--filters))
- `selected_label` (String) The label of the selected filter of a `select_predefined_sql` variable. When not sent, the first filter is selected. A label that matches none of `filters` is refused.

<a id="nestedatt--variables--structured_value"></a>
### Nested Schema for `variables.structured_value`

Read-Only:

- `selected_label` (String) Label of the selected value.
- `canonical_source_ids` (List of String) Selected sources of a `source` variable. Source IDs in `values` set them.
- `service_ids` (List of String) Selected services of a `source` variable.
- `filters` (Attributes List) Saved filters, each a label and the SQL that lists its values. (see [below for nested schema](#nestedatt--variables--structured_value--filters))

<a id="nestedatt--variables--structured_value--filters"></a>
### Nested Schema for `variables.structured_value.filters`

Read-Only:

- `label` (String) The filter's label, shown in the variable's menu.
- `sql` (String) A complete query that returns one column with the filter's options.

<a id="nestedatt--variables--filters"></a>
### Nested Schema for `variables.filters`

Read-Only:

- `label` (String) The filter's label, shown in the variable's menu.
- `sql` (String) A complete query that returns one column with the filter's options.
