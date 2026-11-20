---
page_title: "betterstack_errors_linear_integration Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Linear integration: which errors it files as Linear issues, with which issue properties, and how it keeps them in sync.
---

# betterstack_errors_linear_integration (Resource)

An Error tracking Linear integration: which errors it files as Linear issues, with which issue properties, and how it keeps them in sync.

This resource adopts an existing object using its required `id`, reads it, then applies configured
attributes through the update endpoint. It never creates a remote object.

Removing this resource only forgets Terraform state. No remote delete request is made.

## Example Usage

```terraform
resource "betterstack_errors_linear_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) Existing object identity. Changing it replaces this Terraform resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `applications` (Attributes List, see [nested schema](#nestedatt--applications)) Applications this integration acts on, each with its environments. Empty means any application. An update replaces the whole list.
- `notifications` (Attributes, see [nested schema](#nestedatt--notifications)) "Create issues when": which events file a Linear issue. An update changes only the toggles sent.
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`.
- `tag_filter` (Attributes List, see [nested schema](#nestedatt--tag_filter)) Acts on an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list.
- `name` (String) Integration name.
- `linear_team_id` (String) Linear team issues are created in. A new team clears `linear_status_id`, `linear_resolved_status_id`, `linear_project_id` and `linear_label_ids` unless the same request sets them. `null` clears it.
- `linear_assignee_id` (String) Linear user new issues are assigned to. `null` clears it.
- `linear_label_ids` (List of String) Linear labels put on new issues. An update replaces the whole list.
- `linear_priority` (String) Priority of new issues. Valid values: `no_priority`, `urgent`, `high`, `medium`, `low`.
- `linear_status_id` (String) Status of new issues. `null` clears it.
- `linear_project_id` (String) Linear project new issues are added to. `null` clears it.
- `auto_resolve` (Boolean) Resolve the error when its Linear issue is done, and move the issue to `linear_resolved_status_id` when the error is resolved.
- `linear_resolved_status_id` (String) Status a resolved error moves its issue to. `null` clears it.
- `auto_reopen` (Boolean) Reopen the Linear issue when its error is automatically re-opened.
- `sync_comments` (Boolean) Share comments between errors and their linked Linear issues.

### Read-Only

- `team_id` (String) ID of the team that owns the integration, as the web app shows it (t123).
- `team_name` (String) Name of the team that owns the integration.

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

- `GET /api/v1/linear-integrations`
- `GET /api/v1/linear-integrations/:id`
- `PATCH /api/v1/linear-integrations/:id`

-> Connecting the integration and removing it happen in Error tracking › Integrations in the web app. This API has no create or delete. Every request needs Manage integrations on errors.

## Import

```shell
terraform import betterstack_errors_linear_integration.example 12345
```
