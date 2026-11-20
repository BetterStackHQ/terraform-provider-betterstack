---
page_title: "betterstack_jira_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A Jira integration that creates and syncs issues for incidents. You connect Jira through OAuth in the Better Stack UI, not through the API.
---

# betterstack_jira_integration (Resource)

A Jira integration that creates and syncs issues for incidents. You connect Jira through OAuth in the Better Stack UI, not through the API.

This resource adopts an existing object using its required `id`, reads it, then applies configured
attributes through the update endpoint. It never creates a remote object.

Removing this resource only forgets Terraform state. No remote delete request is made.

-> Replaces `betteruptime_jira_integration`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_jira_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) Existing object identity. Changing it replaces this Terraform resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Integration name.
- `automatic_issue_creation` (Boolean) Whether to create a Jira issue for every incident automatically. Defaults to `false`.
- `jira_project_key` (String) Key of the Jira project that issues are created in. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `jira_issue_type_id` (String) ID of the Jira issue type for created issues. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `jira_fields` (Dynamic) Extra Jira fields to set on created issues, keyed by Jira field id. Defaults to `{}`.

### Read-Only

- `team_name` (String) Name of the team that owns this integration.

## Endpoints

- `GET /api/v2/jira-integrations`
- `GET /api/v2/jira-integrations/:id`
- `PATCH /api/v2/jira-integrations/:id`

-> Connecting and disconnecting Jira both happen through OAuth in the Better Stack UI. This API can't create or delete a Jira integration.

## Import

```shell
terraform import betterstack_jira_integration.example 12345
```
