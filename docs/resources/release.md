---
page_title: "betterstack_release Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  A release of an error-tracking application, detected automatically or registered through the API so it appears the moment it deploys.
---

# betterstack_release (Resource)

A release of an error-tracking application, detected automatically or registered through the API so it appears the moment it deploys.

Creating this resource records its arguments in Better Stack. Changing them records a new one,
and earlier ones stay in Better Stack's history.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

~> Better Stack can't update a release in place, so changing any argument replaces it.

-> Replaces `logtail_errors_release`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a release

```terraform
resource "betterstack_release" "example" {
  version        = "v1.2.3"
  application_id = "123"
  environment    = ["production"]
}
```

### Register a commit deployed to staging

```terraform
resource "betterstack_release" "register_a_commit_deployed_to_staging" {
  version        = "9f3c2a1"
  application_id = "123"
  environment    = ["staging"]
}
```

## Schema

### Required

- `version` (String, Forces replacement) Release version or commit ref.
- `application_id` (String, Forces replacement) IDs of the applications to register this release on. The ID of a connected application registers the release on the application it's connected to. Leave it out to register the release on every application already reporting this version. If none do, nothing is registered and the response is still 201.

### Optional

- `environment` (List of String, Forces replacement) Environments this release was deployed to. A single string is accepted too. Added to any environments the release already has. Better Stack never returns it, so Terraform keeps the value you set.
- `team_id` (String, Forces replacement) Team the release is reported for, as t123 or as the numeric Errors team ID the Sentry CLI sends as `SENTRY_ORG`. It limits which applications a registration with only a version reaches. An unknown team is refused. Better Stack never returns it, so Terraform keeps the value you set.

### Read-Only

- `id` (String) The ID of this release.
- `environments` (List of String) Environments this release was deployed to. A registration with only a version and no applications adds its environment to every application's existing release of that version.
- `origin` (String) How the release became known. The first one wins.
  - `detected`: The release was found in the first event that carries its version.
  - `reported`: The release was registered through this API, MCP or the Sentry CLI.
- `first_seen_at` (String (RFC 3339)) When the release was registered or first seen in an event, whichever came first. Later ones don't change it.
- `last_seen_at` (String (RFC 3339)) When an event last carried the release. A registered release keeps its registration time until its first event, and it's `null` for releases with no event since the field was added.
- `created_at` (String (RFC 3339)) When the release was created.
- `updated_at` (String (RFC 3339)) When the release was last updated.

## Import

Import an existing release by its ID. Get the ID from the [`betterstack_release` data source](../data-sources/release.md), which finds it by `application_id` or `version`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_release.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_release.example 12345
```
