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
  dashboard_group_id = 789
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Dashboard name. Required unless template_id or source_id seeds the dashboard on create.
- `dashboard_group_id` (Number) Group this dashboard belongs to. Send `null` to remove it from its group.

### Read-Only

- `team_id` (String) Team that owns the dashboard, as its `t`-prefixed Better Stack team ID, such as `t123`.
- `team_name` (String) Name of the team that owns the dashboard.
- `refresh_interval` (Number) Auto-refresh interval in seconds.
- `date_range_from` (String) Default range start, for example now-3h.
- `date_range_to` (String) Default range end, for example now.
- `source_eligibility_sql` (String) SQL that decides which sources this dashboard's charts can read. Returned when you get one dashboard.
- `variables` (Dynamic, see [nested schema](#nestedatt--variables)) Template variables, returned when you get one dashboard. An update replaces all variables: any you don't send is removed, except the built-in time, start_time, end_time and source variables, which always exist. The built-in variables keep their own type. `sql_definition` is accepted only on `select_with_sql` and `multi_select_with_sql` variables.
- `charts` (Attributes List, see [nested schema](#nestedatt--charts)) Charts on the dashboard. Returned when you get one dashboard.
- `sections` (Attributes List, see [nested schema](#nestedatt--sections)) Sections on the dashboard. Returned when you get one dashboard.
- `has_overlaps` (Boolean) Whether any charts overlap on the grid. Returned when you get one dashboard.
- `created_at` (String (RFC 3339)) When the dashboard was created.
- `updated_at` (String (RFC 3339)) When the dashboard was last updated.

<a id="nestedatt--variables"></a>
### Nested Schema for `variables`

- `name` (String)
- `variable_type` (String)
- `values` (List of String)
- `default_values` (List of String)
- `sql_definition` (String)
- `structured_value` (Dynamic) The variable's other settings as a JSON object, such as its selected sources, services and filters.
- `allow_multiple_values` (Boolean)

<a id="nestedatt--charts"></a>
### Nested Schema for `charts`

- `id` (Number)
- `chart_type` (String)
- `name` (String)
- `x` (Number)
- `y` (Number)
- `w` (Number)
- `h` (Number)

<a id="nestedatt--sections"></a>
### Nested Schema for `sections`

- `id` (Number)
- `name` (String)
- `y` (Number)
- `collapsed` (Boolean)

## Endpoints

- `GET /api/v3/dashboards`
- `GET /api/v3/dashboards/:id`
- `POST /api/v3/dashboards`
- `PATCH /api/v3/dashboards/:id`
- `DELETE /api/v3/dashboards/:id`
