---
page_title: "betterstack_source_gcp_project Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  The GCP project connected to a `gcp` source, through Workload Identity Federation from the Better Stack GCP module.
---

# betterstack_source_gcp_project (Resource)

The GCP project connected to a `gcp` source, through Workload Identity Federation from the Better Stack GCP module.

This resource takes over an object that already exists in Better Stack, named by its `source_id`, and
applies the arguments you set to it. It never creates one.

Destroying this resource only removes it from Terraform state. Nothing changes in Better Stack.

-> Replaces `logtail_source_gcp_project`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Connect a GCP project through the Better Stack GCP module

```terraform
resource "betterstack_source" "gcp" {
  name     = "GCP production"
  platform = "gcp"
}

module "better_stack" {
  source         = "github.com/BetterStackHQ/gcp//terraform"
  project_id     = "my-gcp-project"
  source_token   = betterstack_source.gcp.token
  ingesting_host = betterstack_source.gcp.ingesting_host
}

resource "betterstack_source_gcp_project" "gcp" {
  source_id          = betterstack_source.gcp.id
  gcp_project_id     = module.better_stack.project_id
  gcp_project_number = module.better_stack.project_number
}
```

### Reuse a GCP project another source connected

```terraform
resource "betterstack_source" "gcp_staging" {
  name     = "GCP staging"
  platform = "gcp"
}

resource "betterstack_source_gcp_project" "gcp_staging" {
  source_id      = betterstack_source.gcp_staging.id
  gcp_account_id = "1234"
}
```

## Schema

Configure exactly one of these: `gcp_project_id` with `gcp_project_number`, or `gcp_account_id`. The provider checks this when it plans the resource, before it creates or replaces anything.

### Required

- `source_id` (String) ID of the object in Better Stack this resource takes over. Changing it makes Terraform take over that object instead. The previous one keeps its settings in Better Stack.

### Optional

- `gcp_account_id` (String) Better Stack's own ID for a Google Cloud project that another `gcp` source in your organization connected, to reuse that connection. It differs from the project's ID and number. At least one of `gcp_project_id` or `gcp_account_id` must be configured. Better Stack never returns it, so Terraform keeps the value you set.
- `gcp_project_id` (String) ID of the GCP project to connect. At least one of `gcp_project_id` or `gcp_account_id` must be configured. Better Stack never returns it, so Terraform keeps the value you set.
- `gcp_project_number` (String, Computed) Number of the GCP project to connect. Better Stack never returns it, so Terraform keeps the value you set.
- `gcp_log_filter` (String) Google Cloud Logging query that selects which logs the source receives. `null` forwards every log, which is the default. Google Cloud checks the query when it's saved and refuses an invalid one. It's sent once the project is connected, and needs the log sink Better Stack's GCP module creates for the source.
- `team_id` (String) Team that owns the source, such as `t123`. Defaults to the provider's `team_id`. Set it when the object is in another team. Changing it replaces the resource, which takes over the object in that team instead.

## Import

Import an existing source GCP project by its ID. Get the ID via [List sources](https://betterstack.com/docs/logs/api/list-all-existing-sources/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_source_gcp_project.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_source_gcp_project.example 12345
```
