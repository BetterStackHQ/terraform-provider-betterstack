---
page_title: "betterstack_errors_slack_integration Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Slack integration. It sets the Slack channel that gets error notifications, which errors reach it and how updates are posted.
---

# betterstack_errors_slack_integration (Resource)

An Error tracking Slack integration. It sets the Slack channel that gets error notifications, which errors reach it and how updates are posted.

This resource takes over an object that already exists in Better Stack, named by its `id`, and
applies the arguments you set to it. It never creates one.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

-> Replaces `logtail_errors_slack_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_errors_slack_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) ID of the object in Better Stack this resource takes over. Changing it makes Terraform take over that object instead. The previous one keeps its settings in Better Stack.

### Optional

- `applications` (Attributes List, Computed) Applications whose errors this integration posts, each with its environments. Empty means any application. An update replaces the whole list. (see [below for nested schema](#nestedatt--applications))
- `notifications` (Attributes, Computed) Which events post to the channel, as under **Notify me when** in the web app. An update changes only the toggles sent. (see [below for nested schema](#nestedatt--notifications))
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. A positive whole number. Defaults to `10`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`. Defaults to `minute`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. A positive whole number. Defaults to `10`.
- `tag_filter` (Attributes List, Computed) Posts an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. The `new_releases` and `spike_protection` notifications and state changes are posted whatever the tags. (see [below for nested schema](#nestedatt--tag_filter))
- `post_updates` (String) How updates are posted to Slack. Defaults to `single_channel`.
  - `single_channel`: Posts every update to the channel.
  - `thread_based`: Posts each new error to the channel and its updates in a thread.
- `broadcast_updates` (Boolean) With `thread_based`, also send important updates to the channel. Those are resolved and re-opened errors, spikes, and errors reaching an occurrence milestone. Defaults to `true`.
- `ai_sre_on_mention` (Boolean) Start an AI SRE conversation when Better Stack is tagged in the channel. Defaults to `true`.
- `team_id` (String) ID of the team that owns the integration, such as `t123`. Defaults to the provider's `team_id`. Set it when the object is in another team. Changing it replaces the resource, which takes over the object in that team instead.

### Read-Only

- `slack_team_name` (String) Slack workspace of the channel.
- `slack_channel_name` (String) Slack channel errors are posted to.
- `team_name` (String) Name of the team that owns the integration.

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

Optional:

- `application_id` (String) ID of an Error tracking application in the token's team.
- `environments` (List of String) Environments whose errors it posts. Empty means every environment.

<a id="nestedatt--notifications"></a>
### Nested Schema for `notifications`

Optional:

- `spikes` (Boolean) Application has a spike in errors.
- `spike_protection` (Boolean) Spike protection starts sampling exceptions.
- `new_releases` (Boolean) New release is observed.
- `new_errors` (Boolean) New error occurs.
- `occurrence_milestones` (Boolean) Error reaches an occurrence milestone of 10, 100, 200, 500 and so on.
- `occurrence_threshold` (Boolean) Error occurs `occurrence_threshold_count` times within `occurrence_threshold_interval`.
- `automatically_reopened` (Boolean) Error is automatically re-opened.
- `affected_users_threshold` (Boolean) Error reaches `affected_users_threshold_count` affected users.

<a id="nestedatt--tag_filter"></a>
### Nested Schema for `tag_filter`

Optional:

- `tag` (String) Exception tag, such as environment or level.
- `values` (List of String) Matches any of these values.

-> Remove the integration in [Error tracking → Integrations](https://errors.betterstack.com/team/0/integrations) in the web app. Every request needs Manage integrations on errors.

## Import

Import an existing Errors Slack integration by its ID. Get the ID via [List Slack integrations](https://betterstack.com/docs/errors/api/integrations/slack/list/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_errors_slack_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_errors_slack_integration.example 12345
```
