---
page_title: "betterstack_errors_jira_integration Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Jira integration. It sets which errors it files as Jira work items, in which space, and with which work type and fields.
---

# betterstack_errors_jira_integration (Resource)

An Error tracking Jira integration. It sets which errors it files as Jira work items, in which space, and with which work type and fields.

This resource takes over an object that already exists in Better Stack, named by its `id`, and
applies the arguments you set to it. It never creates one.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

-> Replaces `logtail_errors_jira_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_errors_jira_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) ID of the object in Better Stack this resource takes over. Changing it makes Terraform take over that object instead. The previous one keeps its settings in Better Stack.

### Optional

- `applications` (Attributes List, Computed) Applications this integration files work items for, each with its environments. Empty means any application. An update replaces the whole list. (see [below for nested schema](#nestedatt--applications))
- `notifications` (Attributes, Computed) Which events file a Jira work item, as under **Automatic work item creation** in the web app. An update changes only the toggles sent. (see [below for nested schema](#nestedatt--notifications))
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. A positive whole number. Defaults to `10`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`. Defaults to `minute`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. A positive whole number. Defaults to `10`.
- `tag_filter` (Attributes List, Computed) Files a work item for an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. (see [below for nested schema](#nestedatt--tag_filter))
- `name` (String, Computed) Integration name.
- `jira_project_key` (String, Computed) Jira space work items are created in. A new space needs `jira_issue_type_id` in the same request and clears `jira_fields` unless they're sent too.
- `jira_issue_type_id` (String, Computed) Jira work type of new work items.
- `jira_fields` (Dynamic, Computed) Further Jira fields set on new work items, keyed by Jira field ID. Each value has the shape of its Jira field type. A list field such as labels accepts a list of `{label, value}` objects, for example `{"labels": [{"label": "outage", "value": "outage"}]}`. A text field accepts a string, for example `{"customfield_10042": "Checkout"}`. A user, team or parent field accepts one `{label, value}` object, and a single-select or priority field the option's ID as a string. A number field accepts a number, and a date picker field an ISO 8601 time such as `2026-12-01T09:00:00Z`.
- `team_id` (String) ID of the team that owns the integration, such as `t123`. Defaults to the provider's `team_id`. Set it when the object is in another team. Changing it replaces the resource, which takes over the object in that team instead.

### Read-Only

- `jira_site_url` (String) Jira site the integration is connected to.
- `team_name` (String) Name of the team that owns the integration.

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

Optional:

- `application_id` (String) ID of an Error tracking application in the token's team.
- `environments` (List of String) Environments whose errors it files work items for. Empty means every environment.

<a id="nestedatt--notifications"></a>
### Nested Schema for `notifications`

Optional:

- `spikes` (Boolean) Application has a spike in errors.
- `new_errors` (Boolean) New error occurs.
- `automatically_reopened` (Boolean) Error is automatically re-opened.
- `occurrence_milestones` (Boolean) Error reaches an occurrence milestone of 10, 100, 200, 500 and so on. When this integration already filed a Jira work item for the error, we comment on it instead.
- `occurrence_threshold` (Boolean) Error occurs `occurrence_threshold_count` times within `occurrence_threshold_interval`. When this integration already filed a Jira work item for the error, we comment on it instead.
- `affected_users_threshold` (Boolean) Error reaches `affected_users_threshold_count` affected users. When this integration already filed a Jira work item for the error, we comment on it instead.

<a id="nestedatt--tag_filter"></a>
### Nested Schema for `tag_filter`

Optional:

- `tag` (String) Exception tag, such as environment or level.
- `values` (List of String) Matches any of these values.

-> Remove the integration in [Error tracking → Integrations](https://errors.betterstack.com/team/0/integrations) in the web app. Every request needs Manage integrations on errors.

-> The `jira_project_key`, `jira_issue_type_id` and `jira_fields` attributes can't change before a Jira site is chosen or while the authorization has expired, and such a change is refused. Re-authorize Jira in the web app first.

## Import

Import an existing Errors Jira integration by its ID. Get the ID via [List Jira integrations](https://betterstack.com/docs/errors/api/integrations/jira/list/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_errors_jira_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_errors_jira_integration.example 12345
```
