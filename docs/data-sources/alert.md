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
  dashboard_id = "123"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `dashboard_id` (String) Dashboard of the chart this alert watches. `null` for an alert on an exploration.
- `chart_id` (String) Chart this alert watches, on the dashboard in `dashboard_id`. `null` for an alert on an exploration.
- `exploration_id` (String) Exploration this alert watches. `null` for an alert on a dashboard chart.
- `name` (String) Alert name. Renaming marks the alert as manually named, so it stops tracking the chart's name.

### Optional

- `team_id` (String) Team that owns the alert, the team of its chart's dashboard or exploration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `alert_type` (String) Alert type. An update can change it.
  - `threshold`: Fires when the value crosses a threshold.
  - `relative`: Compares the value with the previous period as a percentage change.
  - `anomaly_rrcf`: Uses machine learning anomaly detection. Table charts don't accept it.
- `operator` (String) How the value is compared. Required for threshold and relative alerts. `anomaly_rrcf` alerts don't use it, and it's left out of their responses.
  - `equal`: A `threshold` alert fires when the value equals `value` or `string_value`.
  - `not_equal`: A `threshold` alert fires when the value differs from `value` or `string_value`.
  - `higher_than`: A `threshold` alert fires when the value is above `value`.
  - `higher_than_or_equal`: A `threshold` alert fires when the value is at or above `value`.
  - `lower_than`: A `threshold` alert fires when the value is below `value`.
  - `lower_than_or_equal`: A `threshold` alert fires when the value is at or below `value`.
  - `increases_by`: A `relative` alert fires when the value rises by `value` percent or more.
  - `decreases_by`: A `relative` alert fires when the value falls by `value` percent or more.
  - `changes_by`: A `relative` alert fires when the value rises or falls by `value` percent or more.
- `value` (Number) The threshold for `threshold` alerts, or the percentage change for `relative` alerts. `0` is a valid value. `null` when the alert compares `string_value`. Creating a threshold or relative alert needs `value` or `string_value`. Left out for `anomaly_rrcf` alerts.
- `string_value` (String) The exact text a `threshold` alert compares, instead of `value`, with `equal` or `not_equal` only. When it's set, `value` isn't needed and reads back as `null`. Returned for threshold alerts only.
- `paused` (Boolean) Whether the alert is paused.
- `paused_reason` (String) Why the alert is paused. It's `Manually paused` for an alert paused by hand, or a sentence that says Better Stack paused it and why. `null` when the alert isn't paused.
- `query_period` (Number) The time window of data each check looks at, in seconds.
- `confirmation_period` (Number) How long the condition must hold before an incident starts, in seconds, from `0` to `86400`. `0` starts it at once.
- `recovery_period` (Number) Seconds the value must stay recovered before the incident resolves on its own, or `null` to never resolve it automatically. Set it to `-1` to clear it.
- `aggregation_interval` (Number) The size of the buckets the data is aggregated into before each bucket is evaluated, in seconds.
- `check_period` (Number) How often the alert is checked, in seconds. Not returned for `anomaly_rrcf` alerts. An update that includes `check_period` without `query_period` or `aggregation_interval` sets both to it.
- `series_names` (List of String) The series the alert is limited to. An empty list means every series. Can't be combined with `series_names_except`.
- `series_names_except` (List of String) The series the alert ignores while it watches every other series. Can't be combined with `series_names`, so it's empty when `series_names` is used.
- `source_variable` (String) The sources the alert runs on, as `source:` followed by comma-separated source table names, or `null` when none is set. With the `source_variable` and `platforms_single_source` modes, the value must start with `source:`.
- `source_mode` (String) How the alert picks its sources.
  - `source_variable`: Runs on the sources in `source_variable`.
  - `platforms_single_source`: Runs on the one source in `source_variable`, limited to `source_platforms`.
  - `platforms_all_sources`: Runs on every source that matches `source_platforms`.
- `source_platforms` (List of String) Platforms that limit the sources and services the alert runs on, for the `platforms_single_source` and `platforms_all_sources` source modes.
- `escalation_target` (Attributes) Where the incident is routed. It's the string `current_team`, which is the default, an object with `policy_id` and `policy_name`, or an object with `team_id` and `team_name`. Reads return the string `unknown` when the target can't be looked up or no longer exists. A policy is named by ID, by name or both, and a team by `team_id`. An empty object routes to the current team too. Naming both a policy and a team, an unknown one, or a policy name several match without an ID, is refused. (see [below for nested schema](#nestedatt--escalation_target))
- `incident_cause` (String) A custom description for the incidents the alert starts, in Markdown, or `null` when none is set. It accepts the same template variables as `incident_name`, such as `{{current_value}}` and `{{threshold_value}}`.
- `incident_name` (String) A custom name for the incidents the alert starts, in plain text, or `null` when none is set. It accepts template variables such as `{{alert_name}}`, `{{series_name}}`, `{{current_value}}`, `{{threshold_value}}`, `{{team_name}}`, the chart's variables and the columns its query returns.
- `incident_per_series` (Boolean) Whether each series that breaches the condition starts its own incident, rather than one incident for the whole alert.
- `additional_conditions` (Attributes List) Up to 4 more conditions on top of the alert's own. A series fires only when it breaches the alert's own condition and every other condition that covers it, in the same time bucket. A condition limited to other series is met when one of those series breaches it. Each has an `alert_type` of `threshold` or `relative`, an `operator`, a `value` or `string_value`, and optionally `series_names` or `series_names_except`. Anomaly detection can't be an additional condition. An empty list removes them all. (see [below for nested schema](#nestedatt--additional_conditions))
- `variable_values` (Attributes List) Values this alert sets for the template variables of its dashboard or exploration. Each `name` must be a variable defined there. Source variables aren't accepted, so use `source_mode` and `source_variable` instead. (see [below for nested schema](#nestedatt--variable_values))
- `on_missing_data` (String) What a check does when the query returns no data. It can't be sent for `anomaly_rrcf` alerts and is left out of their responses. Defaults to `dont_fire` on threshold and relative alerts. Anomaly alerts don't accept it.
  - `treat_as_zero`: Checks the missing data as if the value were 0.
  - `dont_fire`: Starts no incident for missing data.
  - `treat_as_previous`: Checks the missing data as if it had the previous value.
  - `start_incident`: Starts an incident when the query returns no data.
- `anomaly_sensitivity` (Number) How far a value must deviate from its prediction to trigger the alert, for `anomaly_rrcf` alerts only. The app shows it as Significance. A higher value needs a larger deviation, so the alert fires less often. Left out for threshold and relative alerts. Defaults to `5.0` on `anomaly_rrcf` alerts. Threshold and relative alerts don't accept it.
- `anomaly_trigger` (String) Which direction of anomaly fires the alert, for `anomaly_rrcf` alerts only. Left out of the response on threshold and relative alerts. Defaults to `any` on `anomaly_rrcf` alerts. Threshold and relative alerts don't accept it.
  - `any`: Fires on values above or below the prediction.
  - `higher`: Fires on values above the prediction.
  - `lower`: Fires on values below the prediction.
- `shown_interval` (Number) Interval in seconds the anomaly chart shows, for `anomaly_rrcf` alerts only. It's one of 30, 60, 120, 300, 600, 1800, 3600, 7200, 10800, 21600, 43200 or 86400. Left out of the response on threshold and relative alerts.
- `anomaly_training_range_days` (Number) Days of history the anomaly model is first trained on, from 1 to 30. For `anomaly_rrcf` alerts only. Left out of the response on threshold and relative alerts. Defaults to `7` on `anomaly_rrcf` alerts. Threshold and relative alerts don't accept it.
- `call` (Boolean) Trigger phone-call alerts. Left out of reads when the alert routes to an escalation policy. A plan without phone call alerts refuses `true` with a 422, unless the alert already has it. Defaults to `false` on create. An alert that routes to an escalation policy doesn't use it.
- `sms` (Boolean) Trigger SMS alerts. Left out of reads when the alert routes to an escalation policy. A plan without SMS alerts refuses `true` with a 422, unless the alert already has it. Defaults to `false` on create. An alert that routes to an escalation policy doesn't use it.
- `email` (Boolean) Trigger e-mail alerts. Left out of reads when the alert routes to an escalation policy. Defaults to `false` on create. An alert that routes to an escalation policy doesn't use it.
- `push` (Boolean) Trigger mobile push alerts. Left out of reads when the alert routes to an escalation policy. A plan without push notifications refuses `true` with a 422, unless the alert already has it. Defaults to `false` on create. An alert that routes to an escalation policy doesn't use it.
- `critical_alert` (Boolean) Trigger critical push alerts that bypass silent mode. Left out of reads when the alert routes to an escalation policy. A plan without critical alerts refuses `true` with a 422, unless the alert already has it. Defaults to `false` on create. An alert that routes to an escalation policy doesn't use it.
- `created_at` (String (RFC 3339)) When the alert was created.
- `updated_at` (String (RFC 3339)) When the alert was last updated.
- `metadata` (Map of List of String) Custom incident metadata, where each key you choose holds a list of strings. A single string is accepted too and saved as a one-item list. It's added to the alert's incidents. An update replaces all of it.

<a id="nestedatt--escalation_target"></a>
### Nested Schema for `escalation_target`

Read-Only:

- `policy_id` (String) Escalation policy ID.
- `policy_name` (String) Escalation policy name.
- `team_id` (String) Team ID, such as `t123`.
- `team_name` (String) Team name.

<a id="nestedatt--additional_conditions"></a>
### Nested Schema for `additional_conditions`

Read-Only:

- `alert_type` (String) The type of the condition, `threshold` or `relative`.
- `operator` (String) The comparison the condition makes, from the same values as `operator`.
- `value` (Number) Threshold value. `null` when the condition compares `string_value`.
- `string_value` (String) Threshold value for a string series. `null` when the condition compares `value`.
- `series_names` (List of String) The series the condition is limited to. An empty list means every series.
- `series_names_except` (List of String) The series the condition ignores while it watches every other series.

<a id="nestedatt--variable_values"></a>
### Nested Schema for `variable_values`

Read-Only:

- `name` (String) The name of a variable on the alert's dashboard or exploration.
- `selected_label` (String) For a `select_predefined_sql` variable, the label of the filter the alert uses.
- `values` (List of String) The values the alert uses for the variable, as strings.
