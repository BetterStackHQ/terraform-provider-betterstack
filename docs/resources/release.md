---
page_title: "betterstack_release Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  A release of an error-tracking application, detected automatically or registered through the API so it appears the moment it deploys.
---

# betterstack_release (Resource)

A release of an error-tracking application, detected automatically or registered through the API so it appears the moment it deploys.

Creating this resource registers the configured inputs. Changes to registration inputs require
replacement; existing registrations remain remote history.

Removing this resource only forgets Terraform state. No remote delete request is made.

~> The API has no update endpoint for this resource: changing any attribute forces a replacement.

-> Replaces `logtail_errors_release`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_release" "example" {
  version        = "v1.2.3"
  application_id = 1
}
```

## Schema

### Required

- `version` (String) Release version or commit ref. Changing it replaces the resource.
- `application_id` (Number) IDs of the applications to register this release on. The ID of a connected application registers the release on the application it's connected to. Without it, the release is registered on every application already reporting this version. If none do, nothing is registered and the response is still 201. Can only be set when creating the resource; changing it afterwards replaces the resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `environment` (List of String) Environments this release was deployed to. A single string is accepted too. Added to any environments the release already has. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `team_id` (String) Team the release is reported for: t123, or the numeric Errors team ID the Sentry CLI sends as `SENTRY_ORG`. It limits which applications a registration with only a version reaches. An unknown team answers 422. Write-only: the API never returns it; Terraform keeps the configured value. Can only be set when creating the resource; changing it afterwards replaces the resource.

### Read-Only

- `id` (String) The id of this release.
- `environments` (List of String) Environments this release was deployed to. On create, the given environments are added to any the release already has. A registration with only a version and no applications adds its environment to every application's existing release of that version.
- `origin` (String) `detected` means it was found by the periodic scan of exceptions. `reported` means it was registered through this API, MCP or a CI/CD step. Whichever comes first sets it, and later registrations never change it. Valid values: `detected`, `reported`.
- `first_seen_at` (String (RFC 3339)) When this release was first seen.
- `last_seen_at` (String (RFC 3339)) When this release was last seen.
- `created_at` (String (RFC 3339)) When the release was created.
- `updated_at` (String (RFC 3339)) When the release was last updated.

## Endpoints

- `GET /api/v1/releases`
- `GET /api/v1/releases/:id`
- `POST /api/v1/releases`
- `DELETE /api/v1/releases/:id`

## Import

```shell
terraform import betterstack_release.example 12345
```
