---
page_title: "betterstack_errors_jira_integration Data Source - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Jira integration. It sets which errors it files as Jira work items, in which space, and with which work type and fields.
---

# betterstack_errors_jira_integration (Data Source)

An Error tracking Jira integration. It sets which errors it files as Jira work items, in which space, and with which work type and fields.

## Example Usage

```terraform
data "betterstack_errors_jira_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the Errors Jira integration to look up.

### Optional

- `team_id` (String) ID of the team that owns the integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `applications` (Attributes List) Applications this integration files work items for, each with its environments. Empty means any application. An update replaces the whole list. (see [below for nested schema](#nestedatt--applications))
- `notifications` (Attributes) Which events file a Jira work item, as under **Automatic work item creation** in the web app. An update changes only the toggles sent. (see [below for nested schema](#nestedatt--notifications))
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. A positive whole number.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. A positive whole number.
- `tag_filter` (Attributes List) Files a work item for an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. (see [below for nested schema](#nestedatt--tag_filter))
- `name` (String) Integration name.
- `jira_site_url` (String) Jira site the integration is connected to.
- `jira_project_key` (String) Jira space work items are created in. A new space needs `jira_issue_type_id` in the same request and clears `jira_fields` unless they're sent too.
- `jira_issue_type_id` (String) Jira work type of new work items.
- `jira_fields` (Dynamic) Further Jira fields set on new work items, keyed by Jira field ID. Each value has the shape of its Jira field type. A list field such as labels accepts a list of `{label, value}` objects, for example `{"labels": [{"label": "outage", "value": "outage"}]}`. A text field accepts a string, for example `{"customfield_10042": "Checkout"}`. A user, team or parent field accepts one `{label, value}` object, and a single-select or priority field the option's ID as a string. A number field accepts a number, and a date picker field an ISO 8601 time such as `2026-12-01T09:00:00Z`.
- `team_name` (String) Name of the team that owns the integration.

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

Read-Only:

- `application_id` (String) ID of an Error tracking application in the token's team.
- `environments` (List of String) Environments whose errors it files work items for. Empty means every environment.

<a id="nestedatt--notifications"></a>
### Nested Schema for `notifications`

Read-Only:

- `spikes` (Boolean) Application has a spike in errors.
- `new_errors` (Boolean) New error occurs.
- `automatically_reopened` (Boolean) Error is automatically re-opened.
- `occurrence_milestones` (Boolean) Error reaches an occurrence milestone of 10, 100, 200, 500 and so on. When this integration already filed a Jira work item for the error, we comment on it instead.
- `occurrence_threshold` (Boolean) Error occurs `occurrence_threshold_count` times within `occurrence_threshold_interval`. When this integration already filed a Jira work item for the error, we comment on it instead.
- `affected_users_threshold` (Boolean) Error reaches `affected_users_threshold_count` affected users. When this integration already filed a Jira work item for the error, we comment on it instead.

<a id="nestedatt--tag_filter"></a>
### Nested Schema for `tag_filter`

Read-Only:

- `tag` (String) Exception tag, such as environment or level.
- `values` (List of String) Matches any of these values.
