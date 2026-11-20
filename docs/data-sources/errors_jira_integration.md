---
page_title: "betterstack_errors_jira_integration Data Source - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Jira integration: which errors it files as Jira work items, in which space, with which work type and fields.
---

# betterstack_errors_jira_integration (Data Source)

An Error tracking Jira integration: which errors it files as Jira work items, in which space, with which work type and fields.

## Example Usage

```terraform
data "betterstack_errors_jira_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the errors_jira_integration to look up.

### Read-Only

- `team_id` (String) ID of the team that owns the integration, as the web app shows it (t123).
- `team_name` (String) Name of the team that owns the integration.
- `applications` (Attributes List, see [nested schema](#nestedatt--applications)) Applications this integration acts on, each with its environments. Empty means any application. An update replaces the whole list.
- `notifications` (Attributes, see [nested schema](#nestedatt--notifications)) "Create work items when": which events file a Jira work item. An update changes only the toggles sent.
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`.
- `tag_filter` (Attributes List, see [nested schema](#nestedatt--tag_filter)) Acts on an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list.
- `name` (String) Integration name.
- `jira_site_url` (String) Jira site the integration is connected to.
- `jira_project_key` (String) Jira space work items are created in. A new space needs `jira_issue_type_id` in the same request and clears `jira_fields` unless they're sent too.
- `jira_issue_type_id` (String) Jira work type of new work items.
- `jira_fields` (Dynamic) Further Jira fields set on new work items, keyed by Jira field ID.

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

- `application_id` (Number) ID of an Error tracking application in the token's team.
- `environments` (List of String) Environments to act on errors from. Empty means every environment.

<a id="nestedatt--notifications"></a>
### Nested Schema for `notifications`

- `spikes` (Boolean) Application has a spike in errors.
- `new_errors` (Boolean) New error occurs.
- `automatically_reopened` (Boolean) Error is automatically re-opened.
- `occurrence_milestones` (Boolean) Error reaches an occurrence milestone (10, 100, 200, 500...).
- `occurrence_threshold` (Boolean) Error occurs `occurrence_threshold_count` times within `occurrence_threshold_interval`.
- `affected_users_threshold` (Boolean) Error reaches `affected_users_threshold_count` affected users.

<a id="nestedatt--tag_filter"></a>
### Nested Schema for `tag_filter`

- `tag` (String) Exception tag, such as environment or level.
- `values` (List of String) Matches any of these values.

## Endpoints

- `GET /api/v1/jira-integrations`
- `GET /api/v1/jira-integrations/:id`
- `PATCH /api/v1/jira-integrations/:id`
