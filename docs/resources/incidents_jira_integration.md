---
page_title: "betterstack_incidents_jira_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A Jira integration that creates and syncs issues for incidents.
---

# betterstack_incidents_jira_integration (Resource)

A Jira integration that creates and syncs issues for incidents.

This resource takes over an object that already exists in Better Stack, named by its `id`, and
applies the arguments you set to it. It never creates one.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

-> Replaces `betteruptime_jira_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_incidents_jira_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) ID of the object in Better Stack this resource takes over. Changing it makes Terraform take over that object instead. The previous one keeps its settings in Better Stack.

### Optional

- `name` (String, Computed) Integration name. `null` while the integration is still being connected.
- `automatic_issue_creation` (Boolean) Whether to create a Jira issue for every incident automatically. Defaults to `false`.
- `jira_project_key` (String, Computed) Key of the Jira project that issues are created in. Get the key from your Jira issue keys, such as `ABC` in `ABC-123`. A key that isn't found is refused, and the error lists the available keys.
- `jira_issue_type_id` (String, Computed) ID of the Jira issue type for created issues. Get the ID via the Get issue types for project endpoint of Jira's REST API.
- `jira_fields` (Dynamic, Computed) Extra Jira fields to set on created issues, keyed by Jira field ID. Set the integration up in the app, then use the `jira_fields` from [Get a Jira integration](https://betterstack.com/docs/uptime/api/single-jira-integration/) as a template. Each value has the shape of its Jira field type. A list field such as labels accepts a list of `{label, value}` objects, for example `{"labels": [{"label": "outage", "value": "outage"}]}`. A text field accepts a string, for example `{"customfield_10042": "Checkout"}`. A user, team or parent field accepts one `{label, value}` object, and a single-select or priority field the option's ID as a string. A number field accepts a number, and a date picker field an ISO 8601 time such as `2026-12-01T09:00:00Z`. Defaults to `{}`.
- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it when the object is in another team. Changing it replaces the resource, which takes over the object in that team instead.

### Read-Only

- `team_name` (String) Name of the team that owns this integration.

## Import

Import an existing Jira integration by its ID.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_incidents_jira_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_incidents_jira_integration.example 12345
```
