---
page_title: "betterstack_errors_slack_integration Data Source - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Slack integration. It sets the Slack channel that gets error notifications, which errors reach it and how updates are posted.
---

# betterstack_errors_slack_integration (Data Source)

An Error tracking Slack integration. It sets the Slack channel that gets error notifications, which errors reach it and how updates are posted.

## Example Usage

```terraform
data "betterstack_errors_slack_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the Errors Slack integration to look up.

### Optional

- `team_id` (String) ID of the team that owns the integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `applications` (Attributes List) Applications whose errors this integration posts, each with its environments. Empty means any application. An update replaces the whole list. (see [below for nested schema](#nestedatt--applications))
- `notifications` (Attributes) Which events post to the channel, as under **Notify me when** in the web app. An update changes only the toggles sent. (see [below for nested schema](#nestedatt--notifications))
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. A positive whole number.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. A positive whole number.
- `tag_filter` (Attributes List) Posts an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. The `new_releases` and `spike_protection` notifications and state changes are posted whatever the tags. (see [below for nested schema](#nestedatt--tag_filter))
- `slack_team_name` (String) Slack workspace of the channel.
- `slack_channel_name` (String) Slack channel errors are posted to.
- `post_updates` (String) How updates are posted to Slack.
  - `single_channel`: Posts every update to the channel.
  - `thread_based`: Posts each new error to the channel and its updates in a thread.
- `broadcast_updates` (Boolean) With `thread_based`, also send important updates to the channel. Those are resolved and re-opened errors, spikes, and errors reaching an occurrence milestone.
- `ai_sre_on_mention` (Boolean) Start an AI SRE conversation when Better Stack is tagged in the channel.
- `team_name` (String) Name of the team that owns the integration.

<a id="nestedatt--applications"></a>
### Nested Schema for `applications`

Read-Only:

- `application_id` (String) ID of an Error tracking application in the token's team.
- `environments` (List of String) Environments whose errors it posts. Empty means every environment.

<a id="nestedatt--notifications"></a>
### Nested Schema for `notifications`

Read-Only:

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

Read-Only:

- `tag` (String) Exception tag, such as environment or level.
- `values` (List of String) Matches any of these values.
