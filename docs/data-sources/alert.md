---
page_title: "betterstack_alert Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  An alert on a dashboard chart or a saved exploration that opens an incident when its condition is met.
---

# betterstack_alert (Data Source)

An alert on a dashboard chart or a saved exploration that opens an incident when its condition is met.

## Example Usage

```terraform
data "betterstack_alert" "example" {
  dashboard_id = 1
  chart_id     = 1
  name         = "High Error Rate"
}
```

## Schema

### Address

- `dashboard_id` (Number, Required) Dashboard of the chart this alert watches. `null` for an alert on an exploration. Changing it replaces the resource.
- `chart_id` (Number, Required) Chart this alert watches. It can't be changed after the alert is created. The alert belongs to the chart's dashboard or exploration. Changing it replaces the resource.

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Alert name. Renaming marks the alert as manually named, so it stops tracking the chart's name.

### Read-Only

- `exploration_id` (Number) Exploration this alert watches. `null` for an alert on a dashboard chart.
- `alert_type` (String) Alert type: `threshold` fires when the value crosses a threshold, `relative` compares it with the previous period, and `anomaly_rrcf` uses anomaly detection. Valid values: `threshold`, `relative`, `anomaly_rrcf`.
- `operator` (String) How the value is compared. Required for threshold and relative alerts. Anomaly alerts don't use it and return `null`. Valid values: `equal`, `not_equal`, `higher_than`, `higher_than_or_equal`, `lower_than`, `lower_than_or_equal`, `increases_by`, `decreases_by`, `changes_by`.
- `value` (Number) Threshold value. Creating a threshold or relative alert needs `value` or `string_value`. `null` on anomaly alerts.
- `string_value` (String) Threshold value for a string series, for threshold alerts only.
- `paused` (Boolean) Whether the alert is paused. Send it as a JSON boolean: the string `"true"` is ignored.
- `paused_reason` (String) Why the alert is paused: `manual`, `too_complex`, `query_errors`, `sources_not_eligible`, `anomaly_training_no_data` or `chart_not_alertable`. `null` when the alert isn't paused.
- `query_period` (Number) Query window in seconds. Defaults to `aggregation_interval`, then `check_period`, then 60.
- `confirmation_period` (Number) Seconds the condition must hold before the incident opens.
- `recovery_period` (Number) Seconds the value must stay recovered before the incident resolves on its own. Defaults to 60. Send `null` to never resolve it automatically.
- `aggregation_interval` (Number) Aggregation interval in seconds. Defaults to `query_period`, then `check_period`, then 60.
- `check_period` (Number) How often the alert is evaluated, in seconds. Required when creating a threshold or relative alert. Not returned for anomaly alerts. An update that sends `check_period` without `query_period` or `aggregation_interval` sets both to it.
- `series_names` (List of String) Evaluate only these series. Can't be combined with `series_names_except`. Sending either one replaces both.
- `series_names_except` (List of String) Evaluate every series except these. Can't be combined with `series_names`. Sending either one replaces both.
- `source_variable` (String) Source variable to evaluate, for example `source:table_name`. Defaults to the chart's own source.
- `source_mode` (String) How the sources under evaluation are selected. Valid values: `source_variable`, `platforms_single_source`, `platforms_all_sources`.
- `source_platforms` (List of String) Platforms to evaluate, for the `platforms_single_source` and `platforms_all_sources` source modes.
- `escalation_target` (Dynamic) Where the incident is routed: the string `current_team`, which is the default, an object with a numeric `policy_id` and `policy_name`, or an object with `team_id` and `team_name`. Reads return the string `unknown` when the target can't be looked up or no longer exists. When writing, name the policy or team by ID, by name or both. Naming both a policy and a team, an unknown one, or a name several match without an ID returns 422.
- `incident_cause` (String) Incident cause text.
- `incident_name` (String) Incident name text.
- `incident_per_series` (Boolean) Open one incident per series rather than one for the whole alert.
- `metadata` (Dynamic) Custom incident metadata: arbitrary keys whose values are a string or an array of strings. Sending an empty object clears it.
- `additional_conditions` (Attributes List, see [nested schema](#nestedatt--additional_conditions)) Extra conditions evaluated alongside the alert's own.
- `variable_values` (Attributes List, see [nested schema](#nestedatt--variable_values)) Values this alert sets for the template variables of its dashboard or exploration. Each `name` must be a variable defined there. Source variables aren't accepted: use `source_mode` and `source_variable` instead.
- `on_missing_data` (String) What to do when the query returns no data. Anomaly alerts take it from the chart, so it can't be sent for them and is `null`. Valid values: `treat_as_zero`, `dont_fire`, `treat_as_previous`, `start_incident`.
- `anomaly_sensitivity` (Number) Anomaly sensitivity, for `anomaly_rrcf` alerts only. `null` on threshold and relative alerts.
- `anomaly_trigger` (String) Which direction of anomaly fires the alert, for `anomaly_rrcf` alerts only. `null` on threshold and relative alerts. Valid values: `any`, `higher`, `lower`.
- `anomaly_training_range_days` (Number) Days of history the anomaly model is first trained on, from 1 to 30. For `anomaly_rrcf` alerts only. `null` on threshold and relative alerts.
- `call` (Boolean) Trigger phone-call alerts.
- `sms` (Boolean) Trigger SMS alerts.
- `email` (Boolean) Trigger e-mail alerts.
- `push` (Boolean) Trigger mobile push alerts.
- `critical_alert` (Boolean) Trigger critical push alerts that bypass silent mode.
- `created_at` (String (RFC 3339)) When the alert was created.
- `updated_at` (String (RFC 3339)) When the alert was last updated.

<a id="nestedatt--additional_conditions"></a>
### Nested Schema for `additional_conditions`

- `alert_type` (String)
- `operator` (String)
- `value` (Number)
- `string_value` (String)
- `series_names` (List of String)
- `series_names_except` (List of String)

<a id="nestedatt--variable_values"></a>
### Nested Schema for `variable_values`

- `name` (String)
- `selected_label` (String)
- `values` (List of String)

## Endpoints

- `GET /api/v2/dashboards/:dashboard_id/charts/:chart_id/alerts`
- `GET /api/v2/dashboards/:dashboard_id/charts/:chart_id/alerts/:id`
- `POST /api/v2/dashboards/:dashboard_id/charts/:chart_id/alerts`
- `PATCH /api/v2/dashboards/:dashboard_id/charts/:chart_id/alerts/:id`
- `DELETE /api/v2/dashboards/:dashboard_id/charts/:chart_id/alerts/:id`
