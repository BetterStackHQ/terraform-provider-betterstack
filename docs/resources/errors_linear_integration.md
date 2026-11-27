---
page_title: "betterstack_errors_linear_integration Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Linear integration. It sets which errors it files as Linear issues, with which issue properties, and how it keeps them in sync.
---

# betterstack_errors_linear_integration (Resource)

An Error tracking Linear integration. It sets which errors it files as Linear issues, with which issue properties, and how it keeps them in sync.

This resource takes over an object that already exists in Better Stack, named by its `id`, and
applies the arguments you set to it. It never creates one.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

-> Replaces `logtail_errors_linear_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_errors_linear_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) ID of the object in Better Stack this resource takes over. Changing it makes Terraform take over that object instead. The previous one keeps its settings in Better Stack.

### Optional

- `applications` (Attributes List, Computed) Applications this integration files issues for, each with its environments. Empty means any application. An update replaces the whole list. (see [below for nested schema](#nestedatt--applications))
- `notifications` (Attributes, Computed) Which events file a Linear issue, as under **Automatic issue creation** in the web app. An update changes only the toggles sent. (see [below for nested schema](#nestedatt--notifications))
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. A positive whole number. Defaults to `10`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`. Defaults to `minute`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. A positive whole number. Defaults to `10`.
- `tag_filter` (Attributes List, Computed) Files an issue for an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. (see [below for nested schema](#nestedatt--tag_filter))
- `name` (String, Computed) Integration name.
- `linear_team_id` (String, Computed) Linear team issues are created in. To get the Linear IDs, pick them once on the integration's Configure page and read them back via [Get a Linear integration](https://betterstack.com/docs/errors/api/integrations/linear/get/). A new team clears `linear_status_id`, `linear_resolved_status_id`, `linear_project_id` and `linear_label_ids` unless the same request sets them. `null` clears it.
- `linear_assignee_id` (String, Computed) Linear user new issues are assigned to. `null` clears it. Reads return an empty string when none is set.
- `linear_label_ids` (List of String) Linear labels put on new issues. An update replaces the whole list. Defaults to `[]`.
- `linear_priority` (String, Computed) Priority of new issues. Valid values: `no_priority`, `urgent`, `high`, `medium`, `low`.
- `linear_status_id` (String, Computed) Status of new issues. Empty uses the Linear team's default status. `null` clears it. Reads return an empty string when none is set.
- `linear_project_id` (String, Computed) Linear project new issues are added to. `null` clears it. Reads return an empty string when none is set.
- `auto_resolve` (Boolean) Resolve the error when its Linear issue is done, and move the issue to `linear_resolved_status_id` when the error is resolved. Moving the issue works only when `linear_resolved_status_id` is set. Defaults to `true`.
- `linear_resolved_status_id` (String, Computed) Status a resolved error moves its issue to. `null` clears it. Reads return an empty string when none is set.
- `auto_reopen` (Boolean) Reopen the Linear issue when its error is automatically re-opened, moving it back to `linear_status_id`. It works only when `linear_status_id` and `linear_resolved_status_id` are both set. Defaults to `true`.
- `sync_comments` (Boolean) Share comments between errors and their linked Linear issues. Defaults to `true`.
- `team_id` (String) ID of the team that owns the integration, such as `t123`. Defaults to the provider's `team_id`. Set it when the object is in another team. Changing it replaces the resource, which takes over the object in that team instead.

### Read-Only

- `team_name` (String) Name of the team that owns the integration.

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

Optional:

- `application_id` (String) ID of an Error tracking application in the token's team.
- `environments` (List of String) Environments whose errors it files issues for. Empty means every environment.

<a id="nestedatt--notifications"></a>
### Nested Schema for `notifications`

Optional:

- `spikes` (Boolean) Application has a spike in errors.
- `new_errors` (Boolean) New error occurs.
- `automatically_reopened` (Boolean) Error is automatically re-opened.
- `occurrence_milestones` (Boolean) Error reaches an occurrence milestone of 10, 100, 200, 500 and so on. When this integration already filed a Linear issue for the error, we comment on it instead.
- `occurrence_threshold` (Boolean) Error occurs `occurrence_threshold_count` times within `occurrence_threshold_interval`. When this integration already filed a Linear issue for the error, we comment on it instead.
- `affected_users_threshold` (Boolean) Error reaches `affected_users_threshold_count` affected users. When this integration already filed a Linear issue for the error, we comment on it instead.

<a id="nestedatt--tag_filter"></a>
### Nested Schema for `tag_filter`

Optional:

- `tag` (String) Exception tag, such as environment or level.
- `values` (List of String) Matches any of these values.

-> Remove the integration in [Error tracking → Integrations](https://errors.betterstack.com/team/0/integrations) in the web app. Every request needs Manage integrations on errors.

## Import

Import an existing Errors Linear integration by its ID. Get the ID via [List Linear integrations](https://betterstack.com/docs/errors/api/integrations/linear/list/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_errors_linear_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_errors_linear_integration.example 12345
```
