---
page_title: "betterstack_errors_slack_integration Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Slack integration: the Slack channel that gets error notifications, which errors reach it and how updates are posted.
---

# betterstack_errors_slack_integration (Resource)

An Error tracking Slack integration: the Slack channel that gets error notifications, which errors reach it and how updates are posted.

This resource adopts an existing object using its required `id`, reads it, then applies configured
attributes through the update endpoint. It never creates a remote object.

Removing this resource only forgets Terraform state. No remote delete request is made.

## Example Usage

```terraform
resource "betterstack_errors_slack_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) Existing object identity. Changing it replaces this Terraform resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `applications` (Attributes List, see [nested schema](#nestedatt--applications)) Applications this integration acts on, each with its environments. Empty means any application. An update replaces the whole list. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `notifications` (Attributes, see [nested schema](#nestedatt--notifications)) "Notify me when": which events post to the channel. An update changes only the toggles sent. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. Defaults to `10`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`. Defaults to `minute`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. Defaults to `10`.
- `tag_filter` (Attributes List, see [nested schema](#nestedatt--tag_filter)) Acts on an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `post_updates` (String) `single_channel` posts every update to the channel. `thread_based` posts new errors to the channel and their updates in a thread. Valid values: `single_channel`, `thread_based`. Defaults to `single_channel`.
- `broadcast_updates` (Boolean) With `thread_based`: also send important updates (resolved and re-opened errors, spikes, occurrence milestones) to the channel. Defaults to `true`.
- `ai_sre_on_mention` (Boolean) Start an AI SRE conversation when Better Stack is tagged in the channel. Defaults to `true`.

### Read-Only

- `team_id` (String) ID of the team that owns the integration, as the web app shows it (t123).
- `team_name` (String) Name of the team that owns the integration.
- `slack_team_name` (String) Slack workspace of the channel.
- `slack_channel_name` (String) Slack channel errors are posted to.

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

- `application_id` (String) ID of an Error tracking application in the token's team.
- `environments` (List of String) Environments to act on errors from. Empty means every environment.

<a id="nestedatt--notifications"></a>
### Nested Schema for `notifications`

- `spikes` (Boolean) Application has a spike in errors.
- `spike_protection` (Boolean) Spike protection starts sampling exceptions.
- `new_releases` (Boolean) New release is observed.
- `new_errors` (Boolean) New error occurs.
- `occurrence_milestones` (Boolean) Error reaches an occurrence milestone (10, 100, 200, 500...).
- `occurrence_threshold` (Boolean) Error occurs `occurrence_threshold_count` times within `occurrence_threshold_interval`.
- `automatically_reopened` (Boolean) Error is automatically re-opened.
- `affected_users_threshold` (Boolean) Error reaches `affected_users_threshold_count` affected users.

<a id="nestedatt--tag_filter"></a>
### Nested Schema for `tag_filter`

- `tag` (String) Exception tag, such as environment or level.
- `values` (List of String) Matches any of these values.

## Endpoints

- `GET /api/v1/slack-integrations`
- `GET /api/v1/slack-integrations/:id`
- `PATCH /api/v1/slack-integrations/:id`

-> Connecting the integration and removing it happen in Error tracking › Integrations in the web app. This API has no create or delete. Every request needs Manage integrations on errors.

-> `ai_sre_on_mention` is saved in a second step after the other settings. A 503 then means only that setting wasn't changed.

## Import

```shell
terraform import betterstack_errors_slack_integration.example 12345
```
