---
page_title: "betterstack_release Data Source - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  A release of an error-tracking application, detected automatically or registered through the API so it appears the moment it deploys.
---

# betterstack_release (Data Source)

A release of an error-tracking application, detected automatically or registered through the API so it appears the moment it deploys.

## Example Usage

```terraform
data "betterstack_release" "example" {
  application_id = "12345"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `version` (String) Release version or commit ref.
- `application_id` (String) IDs of the applications to register this release on. The ID of a connected application registers the release on the application it's connected to. Leave it out to register the release on every application already reporting this version. If none do, nothing is registered and the response is still 201.

### Read-Only

- `environment` (List of String) Environments this release was deployed to. A single string is accepted too. Added to any environments the release already has.
- `environments` (List of String) Environments this release was deployed to. A registration with only a version and no applications adds its environment to every application's existing release of that version.
- `origin` (String) How the release became known. The first one wins.
  - `detected`: The release was found in the first event that carries its version.
  - `reported`: The release was registered through this API, MCP or the Sentry CLI.
- `first_seen_at` (String (RFC 3339)) When the release was registered or first seen in an event, whichever came first. Later ones don't change it.
- `last_seen_at` (String (RFC 3339)) When an event last carried the release. A registered release keeps its registration time until its first event, and it's `null` for releases with no event since the field was added.
- `created_at` (String (RFC 3339)) When the release was created.
- `updated_at` (String (RFC 3339)) When the release was last updated.
- `team_id` (String) Team the release is reported for, as t123 or as the numeric Errors team ID the Sentry CLI sends as `SENTRY_ORG`. It limits which applications a registration with only a version reaches. An unknown team is refused.
