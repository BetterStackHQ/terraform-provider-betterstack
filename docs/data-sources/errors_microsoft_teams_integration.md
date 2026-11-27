---
page_title: "betterstack_errors_microsoft_teams_integration Data Source - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Microsoft Teams integration. It sets which errors reach which channels of the team the Better Stack app was added to.
---

# betterstack_errors_microsoft_teams_integration (Data Source)

An Error tracking Microsoft Teams integration. It sets which errors reach which channels of the team the Better Stack app was added to.

## Example Usage

```terraform
data "betterstack_errors_microsoft_teams_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the Errors Microsoft Teams integration to look up.

### Optional

- `team_id` (String) ID of the team that owns the integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `notifications` (Attributes) Which events post to the channels, as under **Notify me when** in the web app. An update changes only the toggles sent. (see [below for nested schema](#nestedatt--notifications))
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. A positive whole number.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. A positive whole number.
- `tag_filter` (Attributes List) Posts an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. The `new_releases` and `spike_protection` notifications and state changes are posted whatever the tags. (see [below for nested schema](#nestedatt--tag_filter))
- `microsoft_teams_team_name` (String) Microsoft Teams team the Better Stack app was added to.
- `microsoft_teams_channel_name` (String) Channel the app was added to, which gets every error while `routes` is empty.
- `routes` (Attributes List) Which applications' errors go to which channels. An application in no route gets no messages unless a route has no applications. Empty means everything goes to the channel the app was added to. An update replaces the whole list. (see [below for nested schema](#nestedatt--routes))
- `team_name` (String) Name of the team that owns the integration.

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

<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

Read-Only:

- `applications` (Attributes List) Applications whose errors this route sends, each with its environments. Empty means any application. (see [below for nested schema](#nestedatt--routes--applications))
- `channels` (Attributes List) Channels the route posts to. At least one. (see [below for nested schema](#nestedatt--routes--channels))

<a id="nestedatt--routes--applications"></a>
### Nested Schema for `routes.applications`

Read-Only:

- `application_id` (String) The ID of the application whose errors the route sends.
- `environments` (List of String) The environments whose errors the route sends, such as `production`. Empty means every environment.

<a id="nestedatt--routes--channels"></a>
### Nested Schema for `routes.channels`

Read-Only:

- `id` (String) Microsoft Teams channel ID of a standard channel in the team the app was added to.
- `name` (String) Channel name, as Teams lists it.
