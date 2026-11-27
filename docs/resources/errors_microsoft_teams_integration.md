---
page_title: "betterstack_errors_microsoft_teams_integration Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Microsoft Teams integration. It sets which errors reach which channels of the team the Better Stack app was added to.
---

# betterstack_errors_microsoft_teams_integration (Resource)

An Error tracking Microsoft Teams integration. It sets which errors reach which channels of the team the Better Stack app was added to.

This resource takes over an object that already exists in Better Stack, named by its `id`, and
applies the arguments you set to it. It never creates one.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

-> Replaces `logtail_errors_microsoft_teams_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_errors_microsoft_teams_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) ID of the object in Better Stack this resource takes over. Changing it makes Terraform take over that object instead. The previous one keeps its settings in Better Stack.

### Optional

- `notifications` (Attributes, Computed) Which events post to the channels, as under **Notify me when** in the web app. An update changes only the toggles sent. (see [below for nested schema](#nestedatt--notifications))
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`. A positive whole number. Defaults to `10`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`. Defaults to `minute`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`. A positive whole number. Defaults to `10`.
- `tag_filter` (Attributes List, Computed) Posts an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list. The `new_releases` and `spike_protection` notifications and state changes are posted whatever the tags. (see [below for nested schema](#nestedatt--tag_filter))
- `routes` (Attributes List, Computed) Which applications' errors go to which channels. An application in no route gets no messages unless a route has no applications. Empty means everything goes to the channel the app was added to. An update replaces the whole list. (see [below for nested schema](#nestedatt--routes))
- `team_id` (String) ID of the team that owns the integration, such as `t123`. Defaults to the provider's `team_id`. Set it when the object is in another team. Changing it replaces the resource, which takes over the object in that team instead.

### Read-Only

- `microsoft_teams_team_name` (String) Microsoft Teams team the Better Stack app was added to.
- `microsoft_teams_channel_name` (String) Channel the app was added to, which gets every error while `routes` is empty.
- `team_name` (String) Name of the team that owns the integration.

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

<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

Optional:

- `applications` (Attributes List) Applications whose errors this route sends, each with its environments. Empty means any application. (see [below for nested schema](#nestedatt--routes--applications))
- `channels` (Attributes List) Channels the route posts to. At least one. (see [below for nested schema](#nestedatt--routes--channels))

<a id="nestedatt--routes--applications"></a>
### Nested Schema for `routes.applications`

Optional:

- `application_id` (String) The ID of the application whose errors the route sends.
- `environments` (List of String) The environments whose errors the route sends, such as `production`. Empty means every environment.

<a id="nestedatt--routes--channels"></a>
### Nested Schema for `routes.channels`

Optional:

- `id` (String) Microsoft Teams channel ID of a standard channel in the team the app was added to.

Read-Only:

- `name` (String) Channel name, as Teams lists it.

-> Connecting the integration and removing it happen in [Error tracking → Integrations](https://errors.betterstack.com/team/0/integrations) in the web app. This API has no create or delete. Every request needs Manage integrations on errors.

-> A channel new to the routes is checked against the Microsoft Teams team's channels. While Teams can't list them, the change fails.

## Import

Import an existing Errors Microsoft Teams integration by its ID. Get the ID via [List Microsoft Teams integrations](https://betterstack.com/docs/errors/api/integrations/microsoft-teams/list/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_errors_microsoft_teams_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_errors_microsoft_teams_integration.example 12345
```
