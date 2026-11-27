---
page_title: "betterstack_dashboard Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A dashboard of charts visualizing logs, metrics, and traces.
---

# betterstack_dashboard (Data Source)

A dashboard of charts visualizing logs, metrics, and traces.

## Example Usage

```terraform
data "betterstack_dashboard" "example" {
  dashboard_group_id = "789"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Dashboard name.
- `dashboard_group_id` (String) Group this dashboard belongs to.
- `product` (String) The product the dashboard belongs to. Defaults to `telemetry`, and it can't be changed after create.
  - `telemetry`: A dashboard over logs, metrics and traces.
  - `rum`: A Real User Monitoring dashboard.

### Optional

- `team_id` (String) Team that owns the dashboard, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `refresh_interval` (Number) How often the dashboard refreshes its charts, in seconds. `0` turns automatic refresh off.
- `date_range_from` (String) Default range start, for example now-3h.
- `date_range_to` (String) Default range end, for example now.
- `source_eligibility_sql` (String) SQL that decides which sources this dashboard's charts can read. Returned when you get one dashboard.
- `variables` (Attributes List) The dashboard's variables, which chart queries use by name, such as `{{source}}`. Returned when you get one dashboard. Creating a dashboard adds any of the built-in `time`, `start_time`, `end_time` and `source` variables the list leaves out. An update replaces all variables. Any the list leaves out is removed, except the built-in `time`, `start_time`, `end_time` and `source` variables, which always exist. The built-in variables keep their own type. A listed variable that already exists keeps the `variable_type`, `default_values`, `sql_definition`, `structured_value` and `allow_multiple_values` it leaves out, and `null` clears any of them except `variable_type`. Its `values` aren't kept, so list them again, except on a `source` variable, which keeps its sources while it leaves out both `values` and `structured_value`. Changing `variable_type` clears the old `default_values`, `sql_definition` and `structured_value`. See [Query variables](https://betterstack.com/docs/logs/dashboards/variables/). `sql_definition` is accepted only on `select_with_sql` and `multi_select_with_sql` variables. When a variable is saved with empty `values`, a `select_value` gets the first of its `default_values`, a `string`, `number`, `boolean` or `sql_expression` gets all of them, and a `select_with_sql` whose `sql_definition` is new or changed gets the first option it returns. An ISO 8601 date in a `date` or `datetime` variable is saved in `values` as a Unix timestamp in microseconds. A new or changed `sql_definition` or filter `sql` runs against the selected source, and the change is refused when the query fails. (see [below for nested schema](#nestedatt--variables))
- `charts` (Attributes List) A summary of each chart on the dashboard, ordered top to bottom, then left to right. Returned when you get one dashboard. The dashboard chart endpoints return the full chart. (see [below for nested schema](#nestedatt--charts))
- `sections` (Attributes List) Sections on the dashboard. Returned when you get one dashboard. (see [below for nested schema](#nestedatt--sections))
- `has_overlaps` (Boolean) Whether any charts overlap on the grid. Returned when you get one dashboard.
- `template_id` (String) ID of the dashboard template to create the dashboard from, as [List dashboard templates](https://betterstack.com/docs/logs/api/dashboards/templates/) lists it. The dashboard starts with the template's charts, sections, alerts, variables and settings. With it, creating a dashboard accepts `name`, `product`, `source_id`, either `dashboard_group_id` or `resource_group_name`, and `refresh_interval`, `date_range_from`, `date_range_to` and `source_eligibility_sql`, which replace the template's. It doesn't accept `variables`.
- `source_id` (String) ID of the source or application to set the new dashboard up for, as [List sources](https://betterstack.com/docs/logs/api/list-all-existing-sources/) or [List applications](https://betterstack.com/docs/errors/api/applications/list/) lists it. For a `rum` dashboard, it's an application. With it, creating a dashboard accepts `name`, `product`, `template_id`, either `dashboard_group_id` or `resource_group_name`, and `refresh_interval`, `date_range_from`, `date_range_to` and `source_eligibility_sql`. It doesn't accept `variables`.
- `resource_group_name` (String) Name of the dashboard group to put the new dashboard in. The team's group with that name is used, or created when there is none. Use it instead of `dashboard_group_id`, not with it.
- `created_at` (String (RFC 3339)) When the dashboard was created.
- `updated_at` (String (RFC 3339)) When the dashboard was last updated.
- `url` (String) Link to this dashboard's page in the Better Stack web app.
- `team_name` (String) Name of the team that owns the dashboard.

<a id="nestedatt--variables"></a>
### Nested Schema for `variables`

Read-Only:

- `name` (String) The name of the variable, used as `{{name}}` in queries.
- `variable_type` (String) The type of the variable, `source`, `date`, `datetime`, `string`, `number`, `sql_expression`, `select_value`, `boolean`, `select_with_sql` or `select_predefined_sql`.
- `values` (List of String) The variable's current values as strings, or an empty list when it has none. For a `source` variable, the IDs of the selected sources.
- `default_values` (List of String) The variable's default values as strings, or an empty list when it has none.
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
- `canonical_source_ids` (List of String) Selected sources of a source variable. Source IDs in `values` set them.
- `service_ids` (List of String) Selected services of a source variable.
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

<a id="nestedatt--charts"></a>
### Nested Schema for `charts`

Read-Only:

- `id` (String) The chart's ID.
- `chart_type` (String) The chart's type, `line_chart`, `table_chart`, `bar_chart`, `number_chart`, `heatmap_chart`, `anomalies_chart`, `pie_chart`, `scatter_chart`, `gauge_chart`, `tail_chart`, `static_text_chart`, `map_chart`, `text_chart`, `funnel_chart`, `funnel_trend_chart` or `hierarchy_chart`.
- `name` (String) The chart's name.
- `x` (Number) The chart's horizontal position on the 12-column grid, from `0` at the left.
- `y` (Number) The chart's vertical position on the grid, from `0` at the top.
- `w` (Number) The chart's width in grid columns, from `1` to `12`.
- `h` (Number) The chart's height in grid rows.

<a id="nestedatt--sections"></a>
### Nested Schema for `sections`

Read-Only:

- `id` (String) The section's ID.
- `name` (String) The section's name.
- `y` (Number) The section's vertical position on the grid, from `0` at the top. A section spans the full width.
- `collapsed` (Boolean) Whether the section is collapsed.
