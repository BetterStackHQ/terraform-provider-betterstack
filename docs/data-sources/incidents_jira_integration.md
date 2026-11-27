---
page_title: "betterstack_incidents_jira_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A Jira integration that creates and syncs issues for incidents.
---

# betterstack_incidents_jira_integration (Data Source)

A Jira integration that creates and syncs issues for incidents.

## Example Usage

```terraform
data "betterstack_incidents_jira_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the Jira integration to look up.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) Integration name. `null` while the integration is still being connected.
- `automatic_issue_creation` (Boolean) Whether to create a Jira issue for every incident automatically.
- `jira_project_key` (String) Key of the Jira project that issues are created in. Get the key from your Jira issue keys, such as `ABC` in `ABC-123`. A key that isn't found is refused, and the error lists the available keys.
- `jira_issue_type_id` (String) ID of the Jira issue type for created issues. Get the ID via the Get issue types for project endpoint of Jira's REST API.
- `jira_fields` (Dynamic) Extra Jira fields to set on created issues, keyed by Jira field ID. Set the integration up in the app, then use the `jira_fields` from [Get a Jira integration](https://betterstack.com/docs/uptime/api/single-jira-integration/) as a template. Each value has the shape of its Jira field type. A list field such as labels accepts a list of `{label, value}` objects, for example `{"labels": [{"label": "outage", "value": "outage"}]}`. A text field accepts a string, for example `{"customfield_10042": "Checkout"}`. A user, team or parent field accepts one `{label, value}` object, and a single-select or priority field the option's ID as a string. A number field accepts a number, and a date picker field an ISO 8601 time such as `2026-12-01T09:00:00Z`.
- `team_name` (String) Name of the team that owns this integration.
