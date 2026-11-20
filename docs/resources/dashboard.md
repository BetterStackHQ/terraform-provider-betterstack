---
page_title: "betterstack_dashboard Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A dashboard of charts visualizing logs, metrics, and traces.
---

# betterstack_dashboard (Resource)

A dashboard of charts visualizing logs, metrics, and traces.

-> Replaces `logtail_dashboard`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_dashboard" "example" {
  name = "Sample Dashboard"
}
```

## Schema

### Required

- `name` (String) Dashboard name.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `dashboard_group_id` (Number) Group this dashboard belongs to. Send `null` to remove it from its group. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `refresh_interval` (Number) Auto-refresh interval in seconds. Defaults to `30`.
- `date_range_from` (String) Default range start, for example now-3h. Defaults to `now-3h`.
- `date_range_to` (String) Default range end, for example now. Defaults to `now`.
- `source_eligibility_sql` (String) SQL that decides which sources this dashboard's charts can read. Returned when you get one dashboard.
- `variables` (Dynamic, see [nested schema](#nestedatt--variables)) Template variables, returned when you get one dashboard. An update replaces all variables: any you don't send is removed, except the built-in time, start_time, end_time and source variables, which always exist. The built-in variables keep their own type. `sql_definition` is accepted only on `select_with_sql` and `multi_select_with_sql` variables.
- `template_id` (Number) ID of the dashboard template to create the dashboard from, as GET /api/v4/dashboards/templates lists it. The dashboard starts with the template's charts, sections and alerts. With it, creating a dashboard accepts only `name`, `source_id`, and either `dashboard_group_id` or `resource_group_name`. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `source_id` (Number) ID of the source to set the new dashboard up for. With it, creating a dashboard accepts only `name`, `template_id`, and either `dashboard_group_id` or `resource_group_name`. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `resource_group_name` (String) Name of the dashboard group to put the new dashboard in. The team's group with that name is used, or created when there is none. Use it instead of `dashboard_group_id`, not with it. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.

### Read-Only

- `id` (String) The id of this dashboard.
- `team_id` (String) Team that owns the dashboard, as its `t`-prefixed Better Stack team ID, such as `t123`.
- `team_name` (String) Name of the team that owns the dashboard.
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

- `GET /api/v4/dashboards`
- `GET /api/v4/dashboards/:id`
- `POST /api/v4/dashboards`
- `PATCH /api/v4/dashboards/:id`
- `DELETE /api/v4/dashboards/:id`

-> Dashboard templates are a separate resource, `dashboard_template`, at GET /api/v4/dashboards/templates. That endpoint returns templates rather than dashboards.

## Import

```shell
terraform import betterstack_dashboard.example 12345
```
