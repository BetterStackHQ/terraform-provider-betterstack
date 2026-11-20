---
page_title: "betterstack_errors_microsoft_teams_integration Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An Error tracking Microsoft Teams integration: which errors reach which channels of the team the Better Stack app was added to.
---

# betterstack_errors_microsoft_teams_integration (Resource)

An Error tracking Microsoft Teams integration: which errors reach which channels of the team the Better Stack app was added to.

This resource adopts an existing object using its required `id`, reads it, then applies configured
attributes through the update endpoint. It never creates a remote object.

Removing this resource only forgets Terraform state. No remote delete request is made.

## Example Usage

```terraform
resource "betterstack_errors_microsoft_teams_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) Existing object identity. Changing it replaces this Terraform resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `notifications` (Attributes, see [nested schema](#nestedatt--notifications)) "Notify me when": which events post to the channels. An update changes only the toggles sent.
- `occurrence_threshold_count` (Number) Number of occurrences for `notifications.occurrence_threshold`.
- `occurrence_threshold_interval` (String) Time window for `notifications.occurrence_threshold`. Valid values: `minute`, `hour`, `day`, `week`.
- `affected_users_threshold_count` (Number) Total number of affected users for `notifications.affected_users_threshold`.
- `tag_filter` (Attributes List, see [nested schema](#nestedatt--tag_filter)) Acts on an error only when one of its exceptions has every tag listed, each with one of its values. Empty means every error. At most 10 tags with 50 values each, 200 characters per tag or value, each tag once. An update replaces the whole list.
- `routes` (Attributes List, see [nested schema](#nestedatt--routes)) Which applications' errors go to which channels. An application in no route gets no messages unless a route has no applications. Empty means everything goes to the channel the app was added to. An update replaces the whole list.

### Read-Only

- `team_id` (String) ID of the team that owns the integration, as the web app shows it (t123).
- `team_name` (String) Name of the team that owns the integration.
- `microsoft_teams_team_name` (String) Microsoft Teams team the Better Stack app was added to.
- `microsoft_teams_channel_name` (String) Channel the app was added to, which gets every error while `routes` is empty.

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

<a id="nestedatt--routes"></a>
### Nested Schema for `routes`

- `applications` (Attributes List, see [nested schema](#nestedatt--routes--applications)) Applications whose errors this route sends, each with its environments. Empty means any application.
- `channels` (Attributes List, see [nested schema](#nestedatt--routes--channels)) Channels the route posts to. At least one.

<a id="nestedatt--routes--applications"></a>
### Nested Schema for `routes.applications`

- `application_id` (Number)
- `environments` (List of String)

<a id="nestedatt--routes--channels"></a>
### Nested Schema for `routes.channels`

- `id` (String) Microsoft Teams channel ID: a standard channel of the team the app was added to.
- `name` (String) Channel name, as Teams lists it.

## Endpoints

- `GET /api/v1/microsoft-teams-integrations`
- `GET /api/v1/microsoft-teams-integrations/:id`
- `PATCH /api/v1/microsoft-teams-integrations/:id`

-> Connecting the integration and removing it happen in Error tracking › Integrations in the web app. This API has no create or delete. Every request needs Manage integrations on errors.

-> A channel new to the routes is checked against the Microsoft Teams team's channels. While Teams can't list them, the update answers 503.

## Import

```shell
terraform import betterstack_errors_microsoft_teams_integration.example 12345
```
