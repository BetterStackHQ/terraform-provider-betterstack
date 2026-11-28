---
page_title: "Migrate from better-uptime and logtail"
description: |-
  Move resources from the better-uptime and logtail providers to this provider without recreating them.
---

# Migrate from better-uptime and logtail

The `betterstack` provider replaces the `better-uptime` and `logtail` providers with one provider and one API token. The old providers keep working and get bug fixes, but no new features. There is no deadline to move.

The Better Stack CLI moves your resources over without recreating them. It reads your Terraform state and writes two new files, and the migration only changes which provider manages each resource. Nothing changes in Better Stack.

## Migrate your configuration

These steps move a typical configuration, with all resources in one team. If your setup uses several teams, modules, workspaces or provider aliases, read the section for it below first.

You need Terraform 1.8 or newer and a [global API token](https://betterstack.com/settings/global-api-tokens).

1. Install the Better Stack CLI. On Windows, run `irm https://betterstack.sh/install.ps1 | iex` in PowerShell instead.

    ```sh
    curl -fsSL https://betterstack.sh | sh
    ```

2. Sign in to the CLI. To use the API token instead, set it in `BETTERSTACK_API_TOKEN`, the variable the provider reads too.

    ```sh
    betterstack signin
    ```

3. In your configuration's directory, save a copy of the current state. Keep it private, since state holds secrets such as source tokens.

    ```sh
    terraform state pull > before-migration.tfstate
    ```

4. Refresh the state with the old providers. This brings changes made in the Better Stack web app into the state, so the new configuration matches what is in Better Stack.

    ```sh
    terraform apply -refresh-only
    ```

5. Run the migration.

    ```sh
    terraform show -json | betterstack terraform migrate
    ```

    It writes two new files next to your configuration:

    - `migration.tf` tells Terraform to hand each resource to the new provider, with `moved` blocks, and `removed` plus `import` blocks for the few types that can't be moved.
    - `betterstack.tf` holds a `betterstack_*` resource block for each migrated resource, with the values from your state.

6. Add the `betterstack` provider next to the old ones. Keep the old providers until step 10, because Terraform needs them to apply `migration.tf`. Copy `team_id` and the `version` from the commented provider block at the top of `betterstack.tf`. The version is the oldest provider that can take over every resource in your state.

    ```hcl
    terraform {
      required_providers {
        betterstack = {
          source  = "BetterStackHQ/betterstack"
          version = ">= 1.4.0"
        }
        # ...keep better-uptime and logtail here for now
      }
    }

    provider "betterstack" {
      api_token = var.betterstack_api_token
      team_id   = "t123"
    }

    variable "betterstack_api_token" {
      type      = string
      sensitive = true
      default   = null
    }
    ```

7. Replace the old resources with the new ones:

    - Delete each `betteruptime_*` and `logtail_*` resource block that `betterstack.tf` replaces. Each block in `betterstack.tf` starts with a comment naming the old address.
    - In `betterstack.tf`, put back your own variables, locals and expressions where the file has literal values.
    - Point outputs and other references at the new addresses, such as `betterstack_source.app.token` instead of `logtail_source.app.token`.

8. Install the new provider and check the plan.

    ```sh
    terraform init
    terraform plan
    ```

    The plan lists resources that have moved, resources to import and old addresses Terraform will no longer manage. It must not add, change or destroy anything:

    ```text
    Plan: 1 to import, 0 to add, 0 to change, 0 to destroy.
    ```

    If it shows a change, don't apply. See [The plan shows changes](#the-plan-shows-changes).

9. Apply the migration.

    ```sh
    terraform apply
    ```

10. Remove the old providers from `required_providers` and delete their `provider` blocks. Then run the plan again. It should say `No changes`. You can delete `migration.tf` now.

    ```sh
    terraform init
    terraform plan
    ```

Keep `before-migration.tfstate` until you're sure everything works, then delete it. To go back, see [Undo the migration](#undo-the-migration).

## The plan shows changes

Don't apply a plan that adds, changes or destroys a resource. Find the cause, fix the configuration and run `terraform plan` again:

- **A setting changed in the web app.** Run step 4 again, delete `migration.tf` and `betterstack.tf`, and run the migration again.
- **A value differs from `betterstack.tf`.** Your own variable or expression gives another value than the state held. Compare the planned change with the generated block.
- **A setting is missing.** The new provider plans its default for every setting your configuration leaves out. Write every setting you want to keep, as `betterstack.tf` does.
- **A reference still names an old address.** Terraform stops with an error on it. Point it at the new address.

The CLI never replaces a file. To run it again, delete `migration.tf` and `betterstack.tf` first.

## The state doesn't match Better Stack

The CLI migrates every resource type. It skips a resource only when the state doesn't match what's in Better Stack, for example a resource deleted in the web app. It then exits with status 3 instead of 0, and the top of `migration.tf` lists each skipped resource with the reason.

1. Migrate everything else with steps 6 to 9 of the main steps. Leave step 10 for later, since the skipped resources still need the old providers.
2. Fix the cause. Refresh the state with `terraform apply -refresh-only`, or remove a resource that no longer exists from your configuration.
3. Delete `migration.tf`, and rename `betterstack.tf`, for example to `main-betterstack.tf`. The CLI never replaces a file.
4. Run the migration again. It picks up only the resources still on the old providers. Then follow the main steps from step 6, including step 10.

A `for_each` or `count` resource is moved as a whole, so when one instance is skipped, all its instances are.

## The command stops without writing files

When the CLI can't migrate the state safely, it writes nothing and prints `Nothing was migrated and no file was written`, followed by each problem. The most common ones:

- **A team name matches no team.** The old providers never read a team's new name back into the state. Name the team with `--team-id t123`.
- **The CLI doesn't know a resource type or its fields.** Update the CLI with `betterstack update`.
- **A team isn't one your sign-in reaches.** Sign in with an account that's a member of the team, or use a global API token of its organization.
- **The state isn't `terraform show -json` output.** A raw `terraform.tfstate` file or a saved plan doesn't work.

Fix the problem and run the migration again.

## Resources in several teams

The CLI handles a state with resources in several teams of one organization, as written by the old providers' `team_name`. It checks each resource in its own team. `betterstack.tf` sets the provider's `team_id` to the first of the teams, and `team_id` on every resource block, so one provider manages all of them. Use a global API token that reaches every team.

A few resource types can't set `team_id`. For a resource of such a type in another team, the files add a provider with an alias for that team, such as `betterstack.t456`, and set `provider = betterstack.t456` on the resource block. Copy those provider blocks from the top of `betterstack.tf` too.

The CLI stops when one resource block or module call holds resources of such a type from different teams. Split it into one block or module call per team first.

## Resources inside modules

The `moved`, `removed` and `import` blocks for resources in modules go into `migration.tf` at the root, and work from there.

In `betterstack.tf`, the resource blocks of a module are commented out, once per module call, with the module's inputs written as literal values. For each module:

1. Copy the blocks into the module's source, replacing its old resource blocks.
2. Put the module's variables back in place of the literal values.
3. Add `betterstack` to the module's `required_providers`.
4. Where a `module` block passes providers, rename them, such as `providers = { betterstack = betterstack.ops }` instead of `providers = { betteruptime = betteruptime.ops }`.

A module source that other configurations use changes for them too. Migrate those configurations at the same time, or give this one its own copy of the module.

Terraform can't tell which module calls share a source. When a resource in a module is skipped, the same resource is skipped in every module call.

## Resources with count or for_each

The state keeps each instance but not the `count` or `for_each` expression that made them. `betterstack.tf` writes `count` or `for_each` from the instances in the state, with values that differ per instance indexed by `count.index` or `each.key`. Put your own expression back, such as `for_each = local.services`.

## Several workspaces

Run the migration once per workspace, with `terraform workspace select` first, and delete the two files between runs. The files differ only in `team_id`, the import IDs and values from each workspace's state, so write those per workspace:

```hcl
locals {
  team_id = { default = "t123", staging = "t456" }[terraform.workspace]
}

provider "betterstack" {
  team_id = local.team_id
}

import {
  to = betterstack_escalation_policy.main
  id = { default = "131824", staging = "131825" }[terraform.workspace]
}
```

An import ID from another workspace's team fails the plan with `Cannot import non-existent remote object`, so a mix-up stops before anything changes.

## Provider aliases

The state doesn't record which provider alias a resource used. With all resources in one team, the generated blocks have no `provider` argument. Add `provider = betterstack.<alias>` back where the old block had one, and give the `betterstack` provider the same aliases.

## Data sources and references

Data sources hold no state, so nothing moves. Replace each `betteruptime_*` and `logtail_*` data block with the `betterstack_*` type that `migration.tf` names, renaming its arguments as its comment says. Don't import a data source.

Outputs, locals and expressions that read an old address need the new one, and `terraform plan` stops on any left behind. Keep output names as they are, so configurations that read them with `terraform_remote_state` keep working.

One reference needs more than a new address:

- **Team members.** A team member's `id` is its e-mail address in both providers. The old `member_id` has no counterpart. `betterstack team-member list` shows a member's user ID if you need it.

## Secrets and values the files leave out

The CLI never copies a secret from the state into either file:

- **Secrets you set**, such as a bucket's secret key, become a sensitive `variable` in `betterstack.tf`. Set it from the same input your old configuration used, so the plan shows no change.
- **Settings the state doesn't fully record**, such as an outgoing webhook's custom template, are left out with a comment that says so. Leaving them out changes nothing. To keep managing one, copy it from your old block, with its password input.
- **Write-only values** that Better Stack never returns, such as collector database passwords, are left out with a comment. Better Stack keeps them, and leaving them out doesn't change them.
- **A source's `data_region`** is left out with a comment when Better Stack names the region differently from the old state. It can't change after the source is created, and leaving it out keeps it.

## Status page sections and resources

The new provider manages a status page's sections and their resources in `betterstack_status_page`'s `sections` list. There are no separate section or status page resource types.

The CLI moves each `betteruptime_status_page` and writes its `sections` into `betterstack.tf` as Better Stack reads them, with every resource. Each `betteruptime_status_page_section` and `betteruptime_status_page_resource` gets a `removed` block, so Terraform forgets the old address without deleting anything.

`sections` holds the page's whole structure, including sections added in the web app. A section you leave out of the list is deleted, with its resources, so keep every section the file lists.

## Deletion protection

Monitors, heartbeats, sources, applications, collectors, status pages, teams and team members get `deletion_protection = true`, also when moved or imported. The first plan doesn't show it as a change.

`terraform destroy`, or a change that replaces such a resource, then stops with `Deletion protection is on`. To destroy a configuration on purpose, such as a preview environment, set `deletion_protection = false` on its resources and apply that first.

## Undo the migration

Before step 9, nothing has changed. Delete `migration.tf` and `betterstack.tf`, and put your old resource blocks back.

After applying, restore the state you saved in step 3:

1. Put your old configuration back, without `migration.tf` and `betterstack.tf`.
2. Push the saved state. `-force` is needed, because the saved state is older than the current one. Make sure no other Terraform run changed the state since the migration.

    ```sh
    terraform init
    terraform state push -force before-migration.tfstate
    ```

3. Run `terraform plan`. It should say `No changes`.

## Reference for each resource type

What changes for each `better-uptime` and `logtail` resource and data source type. Minimum target provider version: **0.1.0**. Data sources need configuration changes, not state imports.

### Data source: `betteruptime_amazon_sns_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_amazon_sns_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_incidents_amazon_sns_integration with the same exact name and selected provider team. Update single extraction outputs from list indexing to objects. Rename references to policy_id to escalation_policy_id.

Replace the data-source type and refresh. Exact lookup must find one SNS receiver, not an IncomingWebhook; missing or duplicate names fail. Retain topic_arn, subscription_state and URL output references. Do not import data-source state.

### Data source: `betteruptime_incoming_webhook`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_incoming_webhook_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_incidents_incoming_webhook_integration and preserve the exact name lookup. Single extraction-field outputs are objects rather than one-element lists; update output indexing. Rename references to policy_id to escalation_policy_id.

Replace the data-source type, keep name and the selected provider team, and refresh. Missing or duplicate names fail. Data-source state is not imported or moved.

### Data source: `betteruptime_ip_list`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_ip_list`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename the data source to betterstack_ip_list. Keep filter_clusters and references to ips/all_clusters; no import or state move is needed. The feed is public and requires no API token.

Rename the data source and retain its cluster filters; refresh reads the live public IP feed and sorts addresses IPv4 before IPv6.

### Data source: `betteruptime_monitor`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_monitor`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_monitor and retain the exact url filter. Rewrite references to policy_id as escalation_policy_id, and references to expiration_policy_id as expiration_escalation_policy_id.

Replace the data source type and retain url and selected provider team; refresh performs an exact match. Multiple matches now fail instead of choosing the first, so add a unique pronounceable_name or use id if necessary. Do not import or move data-source state.

### Data source: `betteruptime_on_call_calendar`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_on_call_schedule`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_on_call_schedule. Exact name lookup remains available; an omitted or empty name selects the default schedule through the documented default id alias. Rename references to default_calendar to default_schedule.

Replace the data-source type and retain the provider team and optional name. Rewrite references to inline on_call_rotation through data.betterstack_on_call_rotation at the same calendar id. Refresh data sources; do not import, recreate, or regenerate schedules.

### Data source: `betteruptime_policy`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_escalation_policy`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_escalation_policy and retain exact name lookup under the same provider team. Rewrite outputs that read policy_group_id, fallback_policy_id, a step's urgency_id or policy_id, or the instructions comment/reminder fields to escalation_policy_group_id, fallback_escalation_policy_id, severity_id, escalation_policy_id and their instructions_ names.

Replace the data-source type and refresh. Zero or duplicate names fail rather than choosing an arbitrary policy; an explicit id can disambiguate. Do not import or move data-source state.

### Data source: `betteruptime_role`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_role`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Change the data-source type to betterstack_role and retain name. System-role identifiers and custom-role names remain valid lookups.

Replace the data-source declaration without importing state. Use the same organization's provider credential, retain role id references and confirm a read-only refresh resolves the same role. Ambiguous role names require an explicit id.

### Data source: `betteruptime_severity`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_severity`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_severity; preserve exact name lookup and the selected provider team.

Replace the data-source type, retain name, and refresh. Zero or duplicate matches fail; use an explicit id to disambiguate. Data-source state is not imported or moved.

### Data source: `betteruptime_slack_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_slack_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_incidents_slack_integration and retain the exact slack_channel_name filter and selected provider team. Zero or multiple matching channels must fail, not pick an arbitrary channel.

Replace the data source type, retain slack_channel_name and all computed-output references, and refresh under the same selected team. Data sources are not imported or moved.

### Data source: `betteruptime_team_member`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_team_member`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename `email` to `id` in configuration and dependent expressions.
- Change the data-source type to betterstack_team_member and look it up by id, set to the e-mail address the old email argument held. Move team_name to the provider's team_id (or a provider alias for that team).

Replace the data-source type and refresh under the same team. The e-mail address is the lookup: set it as id. Data-source state is not imported or moved.

### Resource: `betteruptime_amazon_sns_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_amazon_sns_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"cause_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"title_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"started_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"acknowledged_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"resolved_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the AmazonSnsIntegration id, receiving URL, topic ARN and subscription state; do not recreate the receiver or restart AWS confirmation. Keep every attribute. Rename policy_id to escalation_policy_id. Convert the five single extraction blocks to objects while keeping rules/custom fields as arrays. Preserve the SNS Subject/title and Message/body extraction defaults and every configured override explicitly. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. topic_arn and subscription_state remain computed, never writable.

### Resource: `betteruptime_aws_cloudwatch_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_aws_cloudwatch_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_aws_cloudwatch_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_azure_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_azure_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_azure_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_catalog_attribute`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_catalog_relation`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

A catalog attribute becomes an entry of its relation's `attributes` list, since `betterstack_catalog_relation` has no standalone attribute resource and moved can't target a list entry. Move the `betteruptime_catalog_relation` whose id equals `relation_id` first, or import the relation as `betterstack_catalog_relation` when the legacy configuration didn't manage it. Add an `attributes` entry for this column with its `name`, its `primary` flag and its `id`, which keeps the column and its values. Delete the legacy block and remove it from state with a removed block whose lifecycle sets `destroy = false`, or with `terraform state rm`, before you plan. Never destroy it, which deletes the column's values.

### Resource: `betteruptime_catalog_record`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_catalog_record`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Conversion: `{"op":"rename","path":"relation_id","to":"catalog_relation_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"group_by","path":"attribute","to":"attributes","key":"attribute","into":"values"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"map_values","path":"attributes[].values[].type","value":{"Endpoint":"Monitor","Schedule":"OnCallSchedule","Policy":"EscalationPolicy","NativeWebhook":"OutgoingWebhook","PagerDutyWebhook":"PagerDutyIntegration","MicrosoftTeamsWebhook":"MicrosoftTeamsIntegration","ZapierWebhook":"ZapierZap"}}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the relation and record ids; relation_id becomes catalog_relation_id. Group legacy attribute blocks by their attribute_id or attribute_name into attributes = [{attribute = {id/name = ...}, values = [...]}], retaining every value and its type/reference. Item types take their current names: Endpoint becomes Monitor, Schedule OnCallSchedule, Policy EscalationPolicy, NativeWebhook OutgoingWebhook, PagerDutyWebhook PagerDutyIntegration, MicrosoftTeamsWebhook MicrosoftTeamsIntegration and ZapierWebhook ZapierZap. A Team reference's item_id now reads back as the t-prefixed public id: write it that way.

Run `betterstack catalog-record get ID --catalog-relation-id RELATION_ID` with the legacy relation_id and copy every attribute value into the configuration. Back up state, stop other applies, remove only the old address from state, then import betterstack_catalog_record as RELATION_ID/ID. Check the attribute values with care: an apply clears every value the configuration leaves out. Never delete and recreate records or attributes, and never replace records in bulk. Require a plan with no changes.

### Resource: `betteruptime_catalog_relation`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_catalog_relation`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Retain the relation id, name, description and match_mode. The relation and all of its columns and records remain organization-owned; import uses the existing relation id.

### Resource: `betteruptime_datadog_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_datadog_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_datadog_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_elastic_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_elastic_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_elastic_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_email_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_incoming_email_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"cause_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"acknowledged_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"resolved_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"started_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"title_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute. Rename policy_id to escalation_policy_id. Convert the five single extraction blocks to objects; keep custom fields and rules as ordered arrays. Legacy target_field belongs to the shared webhook schema but has no meaning for an e-mail's from_email/subject/body targets; remove only those placeholders, not match content. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team.

Back up state and stop other applies. Run `betterstack incidents-incoming-email-integration get ID --team t123` and compare every extraction, rule and channel setting to the translated HCL, keeping the effective defaults. Review a state-only removal of the old address and an import of betterstack_incidents_incoming_email_integration at the same ID. The plan must neither destroy nor create the integration: a new one gets a new receiving e-mail address. Require a plan with no changes, and check that email_address is unchanged afterwards.

### Resource: `betteruptime_google_monitoring_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_google_monitoring_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_google_monitoring_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_grafana_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_grafana_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_grafana_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_heartbeat`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_heartbeat`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Keep every attribute. Rename policy_id to escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Keep the existing heartbeat id, so the ping URL stays the same.

### Resource: `betteruptime_heartbeat_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_heartbeat_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Import uses the existing group id.

### Resource: `betteruptime_incoming_webhook`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_incoming_webhook_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"cause_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"title_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"started_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"acknowledged_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"resolved_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the existing receiver id and URL; do not recreate its secret receiving token. Keep every attribute. Rename policy_id to escalation_policy_id. Convert the five single extraction blocks to nullable objects, preserving extraction target, match type and content; keep custom fields and event rules as ordered arrays. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team.

### Resource: `betteruptime_jira_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_jira_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `jira_fields_json` to `jira_fields` in configuration and dependent expressions.
- Conversion: `{"op":"json_decode","path":"jira_fields"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename better_stack_id in HCL to the required id of the existing integration. Legacy state already stores this same ID. Decode jira_fields_json into the typed jira_fields object, keeping every nested Jira value. This resource takes over a Jira integration already connected in the web app: creating it never connects a new Jira account, and destroying it only removes it from Terraform state. Keep name unchanged.

### Resource: `betteruptime_metadata`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_metadata`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `metadata_value` to `values` in configuration and dependent expressions.
- Conversion: `{"op":"map_values","path":"values[].type","value":{"Endpoint":"Monitor","Schedule":"OnCallSchedule","Policy":"EscalationPolicy","NativeWebhook":"OutgoingWebhook","PagerDutyWebhook":"PagerDutyIntegration","MicrosoftTeamsWebhook":"MicrosoftTeamsIntegration","ZapierWebhook":"ZapierZap"}}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- The identity is the owner_type, owner_id and key group the legacy provider already used. Rename metadata_value blocks to values without losing order, string contents or item references. A deprecated scalar value becomes one {type = "String", value = ...} entry. Item types take their current names: Endpoint becomes Monitor, Schedule OnCallSchedule, Policy EscalationPolicy, NativeWebhook OutgoingWebhook, PagerDutyWebhook PagerDutyIntegration, MicrosoftTeamsWebhook MicrosoftTeamsIntegration and ZapierWebhook ZapierZap. A Team reference's item_id now reads back as the t-prefixed ID and a User reference's as u123: write them that way, since either one without its prefix is refused. Keep owner IDs and other referenced IDs as they are.

Run `betterstack metadata list --owner-type TYPE --owner-id OWNER --team t123` with the provider's team and pick the entry whose key matches. Exactly one entry must match the owner and key: use its ID, even when it differs from the old state. Back up state, stop other applies, remove only the old state address, then import betterstack_metadata at that ID. The plan must not create, update or destroy the metadata: destroying it removes the key with all its values. Keep all values in HCL and require a plan with no changes.

### Resource: `betteruptime_monitor`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_monitor`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `expiration_policy_id` to `expiration_escalation_policy_id` in configuration and dependent expressions.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"monitor_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"expiration_escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute. Rename policy_id to escalation_policy_id and expiration_policy_id to expiration_escalation_policy_id. Its 0 for an unset monitor_group_id or expiration_policy_id becomes null, as Better Stack reads it back. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Preserve all request headers by name/value, dropping only the obsolete header row ids. Import uses the existing id.

### Resource: `betteruptime_monitor_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_monitor_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Import uses the existing group id.

### Resource: `betteruptime_new_relic_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_new_relic_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_new_relic_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_on_call_calendar`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_on_call_schedule`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `default_calendar` to `default_schedule` in configuration and dependent expressions.
- Keep the existing schedule id and name. Rename default_calendar to default_schedule. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Split the legacy on_call_rotation block into a betterstack_on_call_rotation resource addressed by the schedule id. Preserve user ordering, length, rotation_start, rotation_end, timezone and working_hours. Preserve an omitted rotation rather than creating one.

Back up state and stop other applies. Run `betterstack on-call-schedule get ID` and `betterstack on-call-rotation get ID` in the same team. Review the state-only removal of the old calendar address, the import of betterstack_on_call_schedule at the original calendar ID and, when the legacy state holds a rotation and Better Stack shows one, the import of betterstack_on_call_rotation at the same calendar ID. The plan must neither create nor destroy the rotation: creating one replaces the rotation and regenerates future shifts, and destroying it removes shifts. Keep the original schedule events and require a plan with no changes.

### Resource: `betteruptime_outgoing_webhook`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_outgoing_webhook_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"custom_webhook_template_attributes"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the NativeWebhook id, trigger_type, destination URL, notify_alongside_primary_responder and every event toggle explicitly. The generated block leaves custom_webhook_template_attributes out, which changes nothing. To keep managing it, copy it from the old block as an object instead of a single nested block, with its method, ordered headers, body template and basic-auth inputs as they are. Keep body_template as its stored string, including JSON encoding and template variables. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team.

### Resource: `betteruptime_pagerduty_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_outgoing_pagerduty_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the resource to betterstack_incidents_outgoing_pagerduty_integration. Preserve name, routing key, severity and notify_alongside_primary_responder, including the original sensitive input expression. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Import uses the existing id.

### Resource: `betteruptime_policy`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_escalation_policy`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `fallback_policy_id` to `fallback_escalation_policy_id` in configuration and dependent expressions.
- Rename `policy_group_id` to `escalation_policy_group_id` in configuration and dependent expressions.
- Conversion: `{"op":"rename","path":"steps[].urgency_id","to":"steps[].severity_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].policy_id","to":"steps[].escalation_policy_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].comment","to":"steps[].instructions_comment"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].reminder_enabled","to":"steps[].instructions_reminder_enabled"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].reminder_interval_hours","to":"steps[].instructions_reminder_interval_hours"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"steps[].instructions_reminder_enabled","value":false}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"steps[].instructions_reminder_interval_hours","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_escalation_policy and keep every policy attribute. Rename policy_group_id to escalation_policy_group_id and fallback_policy_id to fallback_escalation_policy_id. In ordered steps rename urgency_id to severity_id, policy_id to escalation_policy_id, and comment/reminder_enabled/reminder_interval_hours to instructions_comment/instructions_reminder_enabled/instructions_reminder_interval_hours. An entire_team member that names a team now reads its id as the t-prefixed team id (t123) instead of the local number. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove a step member's team_id: the old provider's member team_id was stored but never read, so the member paged the policy's own team; to page another team, use an entire_team member whose id is that team's t-prefixed id.

Run `betterstack escalation-policy get ID` and compare its ordered steps to the translated HCL. Keep wait-until timezones, time-branch windows, resolved action_type, typed metadata_values and no_value conditions and instruction reminders. Rewrite an entire_team member's numeric id to the t-prefixed team ID the policy shows. A member that pages the policy's own team keeps no id. Convert deprecated metadata_values strings or metadata_value blocks to typed metadata_values, with Team references as t-prefixed IDs and item types under their current names: Endpoint becomes Monitor, Schedule OnCallSchedule, Policy EscalationPolicy, NativeWebhook OutgoingWebhook, PagerDutyWebhook PagerDutyIntegration, MicrosoftTeamsWebhook MicrosoftTeamsIntegration and ZapierWebhook ZapierZap. Back up state, stop other applies, then review the state-only removal of the old address and the import at the same ID. The plan must neither update nor destroy the policy: an update rewrites its steps. Require a plan with no changes.

### Resource: `betteruptime_policy_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_escalation_policy_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the resource to betterstack_escalation_policy_group and keep name and sort_index. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Import uses the existing group id.

### Resource: `betteruptime_prometheus_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_prometheus_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_incidents_prometheus_integration. Keep the integration id, webhook_url, channels and pause state. Rename policy_id to escalation_policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Remove recovery_period: it has no effect for this integration, whose incident resolves as soon as the alert does.

### Resource: `betteruptime_severity`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_severity`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep the existing severity id, name and every notification channel. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team.

### Resource: `betteruptime_severity_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_severity_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Import uses the existing group id.

### Resource: `betteruptime_splunk_oncall_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incidents_splunk_on_call_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the resource to betterstack_incidents_splunk_on_call_integration. Preserve name, destination URL and notify_alongside_primary_responder. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. The id remains the existing NativeWebhook id; do not migrate it as an outgoing_webhook_integration.

### Resource: `betteruptime_status_page`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"sentinel_null","path":"status_page_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute. Its status_page_group_id of zero means no group and becomes null. Keep password in its original sensitive input expression, never copy it into migration output; read refresh must retain the write-only value. Import uses the existing page id.
- The page's betteruptime_status_page_section and betteruptime_status_page_resource blocks fold into its sections list: the refresh after the move reads them into state, and the list written with the same section names and resources keeps every id.

### Resource: `betteruptime_status_page_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page_group`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Keep name and sort_index. Drop team_name, which status page groups never used: they belong to an organization, and organization ids are not team ids.

Run `betterstack status-page-group get ID` with the token the provider will use. Back up state, stop other applies, remove only the old state address without destroying the group, then import betterstack_status_page_group at the same ID. Check that the plan changes nothing in Better Stack before applying the state operations, then require a plan with no changes.

### Resource: `betteruptime_status_page_resource`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

A status page resource becomes an entry of the resources list of its section in its page's sections list: betterstack_status_page has no standalone resource, and moved can't target a list entry. Move the page and write its sections as the betteruptime_status_page_section recipe says, then add this item to the resources of the section whose id equals status_page_section_id (or the page's first section when it is 0), in position order. Keep resource_type, resource_id, public_name, explanation, widget_type and the mark_as_down/mark_as_degraded settings. Write each metadata rule block as an object, with its metadata_value blocks as values. A Team value is t123 and a User value u123, as they read back now, where the legacy provider stored the internal number. Write item types as v5 reads them back: Endpoint becomes Monitor, Schedule OnCallSchedule, Policy EscalationPolicy, NativeWebhook OutgoingWebhook, PagerDutyWebhook PagerDutyIntegration, MicrosoftTeamsWebhook MicrosoftTeamsIntegration and ZapierWebhook ZapierZap. An entry without an id keeps the id of the item with the same resource_type and resource_id, or else the same resource_type and public_name, so the item and its subscribers are kept. Drop fixed_position and position: the list order is the position. Forget only betteruptime_status_page_resource.NAME with removed { from = betteruptime_status_page_resource.NAME lifecycle { destroy = false } }; never destroy it through the legacy provider. Require a plan with no create, update, delete or replacement.

### Resource: `betteruptime_status_page_section`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

A section becomes an entry of its page's sections list: betterstack_status_page has no standalone section resource, and moved can't target a list entry. Move the betteruptime_status_page whose id equals status_page_id first, or import the page as betterstack_status_page when the legacy configuration didn't manage it, then refresh it: the refresh reads the page's sections, with their ids, into state. Add a sections entry for this section with the same name, in the page's display order (position ascending), and nest its resources there. Sent, sections replaces the page's whole structure, so list every section and resource the refresh read, including ones added in the web app and never managed by Terraform: one left out is deleted. An entry without an id keeps the id of the section with the same name, so the section and its resources are kept. Forget only betteruptime_status_page_section.NAME with removed { from = betteruptime_status_page_section.NAME lifecycle { destroy = false } }; never destroy it through the legacy provider, since that deletes the section and the resources in it. Replace references to its id with the page's sections entry and require a plan with no create, update, delete or replacement.

### Resource: `betteruptime_team_member`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_team_member`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Change the resource type to betterstack_team_member and keep email, role and role_id. Remove member_id: `betterstack team-member list` shows the member's user ID if you need it. Move team_name to the provider's team_id (or a provider alias for that team).

Back up state and stop concurrent applies. Remove the old resource from state without destroying it (a removed block with lifecycle { destroy = false }), then import betterstack_team_member by the same e-mail address, the id both providers use. Never apply a destroy of the old resource: it removes the member or cancels the invitation. Require a no-op plan afterward, apart from deletion_protection, which starts true.

### Data source: `logtail_collector`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_collector`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_collector with the same exact name and verified team (or existing id for duplicate names). Convert configuration/custom_bucket singleton references to objects and service/namespace options to maps. The databases, databases_count, hosts_count and hosts_up_count outputs are gone: read targets through collector_target and hosts through the hosts action. Refresh only, never import or recreate the collector.

### Data source: `logtail_connection`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_connection`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_connection with the same connection ID and a token that can see its teams. The data source never reads the password: keep any credential stored elsewhere and do not recreate the connection. team_ids now read as t-prefixed IDs. Refresh only, and never import a data source.

### Data source: `logtail_dashboard`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Replace the data source type with data.betterstack_dashboard and keep the same id or exact name under the verified team. Duplicate name matches require the explicit existing ID. The data output is now a typed exported document fetched from the read-only export action: use jsonencode(data) where an external consumer needs JSON text instead of jsondecode on the old string. team_id now reads back as the t-prefixed id. Refresh only, never import the data source.

### Data source: `logtail_dashboard_alert`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_alert`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_alert and retain dashboard_id, the bare chart_id and exact name as intersecting lookup constraints. Duplicate names require the explicit returned bare alert ID. Rename variable_value accesses to variable_values and singleton escalation_target accesses to the escalation_target object, whose fields are all null for the current team. Explicit null recovery means never; zero means immediate. Refresh only; do not import or recreate an alert.

### Data source: `logtail_dashboard_chart`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_chart`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_dashboard_chart and keep dashboard_id plus exact name. Require one match; use an explicit chart ID for duplicates. Replace query block outputs with queries and old settings JSON parsing with the typed settings object. ID is the bare chart ID while dashboard_id retains parent context. Refresh only, without importing or changing charts.

### Data source: `logtail_dashboard_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_group`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_dashboard_group and retain the exact name lookup under the verified team. For duplicate names select the intended existing ID rather than accepting a first match. Refresh only; do not import a data source.

### Data source: `logtail_dashboard_section`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_section`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_dashboard_section with the same dashboard_id and exact name. Require a unique match, or select the known bare section ID. Preserve y/collapsed outputs and dashboard context. Refresh only; do not import or alter layout.

### Data source: `logtail_dashboard_template`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_template`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_dashboard_template, retaining the template id or exact catalog name (case-only differences should use the known id). Duplicate names require an explicit ID. The id is the template's backing dashboard id, as before. data is no longer an output here: read the document with data.betterstack_dashboard on the same id (its data output, from the export action) and pass it to the dashboard import, or jsonencode it for consumers needing JSON text. Refresh only and never import a data source.

### Data source: `logtail_errors_application`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_application`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_application and retain the exact name lookup under the provider's team. Duplicate names require the existing ID. Rename custom_bucket[0] references to object attributes; team_id now reads t123. Refresh the data source only; no import or application creation.

### Data source: `logtail_errors_application_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_application_group`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_application_group and retain the exact name lookup in the provider's team. A duplicate name requires the intended existing ID; refresh only and never import the data source.

### Data source: `logtail_exploration`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_exploration`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_exploration, retaining exact name lookup in the verified team (or the known ID for ambiguous names). Replace chart[0] with chart, query with queries, variable with variables, and JSON string decoding of chart.settings with its typed object. Keep settings, SQL, variable defaults/selections and group/date-range outputs. Refresh only; do not import or recreate it.

### Data source: `logtail_exploration_alert`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_alert`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_alert, keeping exploration_id plus exact name as intersecting lookup constraints; select the authoritative bare alert ID if names duplicate. An exploration's alert carries no chart_id. Rename variable_value accesses to variable_values and singleton target accesses to the escalation_target object, whose fields are all null for the current team. Null recovery retains never-auto-recover. Refresh only; never import or recreate an alert.

### Data source: `logtail_exploration_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_exploration_group`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_exploration_group and preserve exact name lookup in the verified team. Duplicate names require an explicit existing ID; refresh only, never import the data source.

### Data source: `logtail_metric`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_extracted_metric`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_extracted_metric and preserve both source_id and exact name. Require exactly one rule matching the name on that source; duplicate names fail closed. Keep SQL/type/aggregation outputs, refresh only and never import a data source.

### Data source: `logtail_source`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Replace data.logtail_source with data.betterstack_source and retain the exact table_name lookup. Select the verified team in the provider. Duplicate matches fail rather than choosing another source. Convert custom_bucket references from [0] to object attributes; team_id now reads as the t-prefixed id. Refresh only; never import a data source.

### Data source: `logtail_source_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source_group`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_source_group and keep the exact name lookup in the verified team. Duplicate names require the exact existing ID instead of an arbitrary first match. Refresh only; do not import a data source.

### Resource: `logtail_collector`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_collector`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"custom_bucket"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"configuration"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"configuration.components"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"configuration.vrl_transformation","to":"collector_logs_transformation"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"collector_logs_transformation","value":"\n."}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"collector_logs_transformation","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"source_vrl_transformation","to":"ingestion_logs_transformation"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"ingestion_logs_transformation","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"configuration.components.metrics_databases","to":"configuration.components.metrics_cluster","on_conflict":"target"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"index_by","path":"configuration.service_option","to":"configuration.services_options","key":"name"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"index_by","path":"configuration.namespace_option","to":"configuration.namespaces_options","key":"name"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"source_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"source_group_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"source_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Before moving, check the collector's team with `betterstack collector get ID`. Keep the collector ID and existing sensitive state: importing the collector cannot recover the custom_bucket secret_access_key. Turn the configuration, components and custom_bucket blocks into objects. Replace the service_option and namespace_option sets with services_options and namespaces_options maps keyed by name, keeping numeric log_sampling and boolean ingest_traces. metrics_cluster takes precedence over the deprecated metrics_databases, as in the legacy provider.
- A populated databases list becomes one betterstack_collector_target per row, imported at its existing target ID next to the moved collector, with collector_id referencing it. Leave password and api_key out: they are write-only: Better Stack keeps them and never returns them. Never apply removal of inline blocks through the legacy provider, which deletes targets.
- Preserve source_id, source_group_id, live_tail_pattern, retention, Vector configuration, component flags and both VRL programs. Migration must produce no remote create, update or delete, including credential reassertion. Remove name from custom_bucket: Better Stack reads the bucket name from its endpoint.

betterstack terraform migrate moves the collector and imports one betterstack_collector_target per legacy databases row when the collector's targets, as `betterstack collector-target list --collector-id COLLECTOR_ID --team t123` lists them, hold exactly one target with the same ID, kind (legacy service_type), host and port for each row. Otherwise the collector stays on the legacy provider: fix missing, duplicate or mismatched targets and run it again, and never match targets by their position. By hand: stop other applies and keep a raw `terraform state pull` backup outside the repository. Move logtail_collector.NAME to betterstack_collector.NAME with a moved block. Never import or recreate the collector, whose bucket credentials live only in state. Configure one betterstack_collector_target per database with collector_id = betterstack_collector.NAME.id and the existing kind, host, port, username, ssl_mode and tls, leaving password and api_key out (Better Stack keeps them and never returns them), and import each at its target ID with an import block. Never apply a removal of the inline databases through the legacy provider: it deletes the targets. Require a plan with no create, update, delete or credential change.

### Resource: `logtail_collector_target`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_collector_target`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"sentinel_null","path":"host","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"collector_host","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"collector_host","to":"host"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Before moving, check the target with `betterstack collector-target get ID --collector-id COLLECTOR_ID` and confirm its kind. Keep collector_id, kind, enabled, port, service, listen_ip, endpoint, scheme, username, password, api_key, SSL and TLS settings, status and container. For nginx, apache, kafka, traefik and prometheus targets, rename collector_host to host. A database target keeps host. Move the existing state to keep the write-only credentials, and never recreate a target or duplicate one already split from a collector's inline databases.

### Resource: `logtail_connection`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_connection`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"sentinel_null","path":"valid_until","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"data_sources[].source_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Run `betterstack connection get ID` with a token that can see every team of the connection: its team_ids are the t-prefixed IDs of the teams the legacy numeric team_ids named. Replace a legacy team_ids or team_names configuration with exactly that set, since team_names is only read back now. Never use one selected team in place of a multi-team connection, and never prefix the old numbers yourself.
- Move the existing state. Never remove and import or recreate this connection: its password is shown only when the connection is created and must stay in sensitive state. Keep host, port, username, password, sample_query, data_sources, client_type, the IP allowlist, expiry, note and creator. Remove data_region from a ClickHouse connection's configuration: the connection stays limited to its cluster, which is only read back in cluster, and Better Stack refuses cluster on a ClickHouse connection. For a Grafana connection, set cluster to the cluster name `betterstack connection get ID` shows. Keep valid_until as configured, not the normalized value Better Stack reads back. Check that the plan neither rotates nor replaces the credential.

### Resource: `logtail_dashboard`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `variable` to `variables` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"dashboard_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"dashboard_group_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"variables[].sql_definition","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"variables[].default_values","value":[]}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute. Rename variable blocks to the variables object list, and a dashboard_group_id of 0 to null. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team; team_id now reads back as the t-prefixed id. Import uses the existing id.
- A dashboard the legacy provider created from data (an imported dashboard) keeps its id, charts, queries, alerts, sections and variables: drop data from the configuration, since this resource takes no exported document and must not import it again. Its charts and sections stay unmanaged unless imported as their own resources, and no child may be imported twice.

### Resource: `logtail_dashboard_alert`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_alert`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `variable_value` to `variable_values` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_omit","path":"paused","value":false}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"escalation_target"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.

Back up state with `terraform state pull`. Find the alert with `betterstack alert list --dashboard-id DASHBOARD_ID --chart-id CHART_ID --team t123`, using the bare chart ID, and check that exactly one alert has the old name. Rename to betterstack_alert, keeping the same dashboard_id, bare chart_id and alert ID. Remove the old address from state only, with a `removed` block, and import the same alert at its ID. Never let a plan destroy or create the alert: destroying it resolves its open incidents and removes its link to past ones. Every attribute keeps its name. Keep recovery_period = -1: the provider stores it as null, which means the incident never resolves on its own. Keep 0 for immediate recovery and every positive duration as it is. Rename the variable_value blocks to variable_values, keeping selected_label, values and additional_conditions. Turn the escalation_target block into an object with policy_id or policy_name, or with team_id set to the team's Better Stack ID, such as t123, as `betterstack alert list` shows it (never add t to the legacy number, which is the app's own and can differ), or {} when no block was set, which sends to the current team. A legacy team_name becomes that team's team_id, since team_name is only read back now. When the old target is unknown or stale, find the real target first and never fall back to the current team. Write each metadata value as a list of strings: a legacy env = "prod" becomes env = ["prod"]. Keep the source modes, platforms and variables, incident cause and per-series setting, notification flags, anomaly settings, missing-data behaviour and paused state. Run terraform plan, check every value it reads back, and after applying require a second plan with no changes.

### Resource: `logtail_dashboard_chart`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_chart`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `query` to `queries` in configuration and dependent expressions.
- Conversion: `{"op":"composite_id","path":"id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].name","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].sql_query","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].where_condition","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].static_text","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].source_variable","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute. Convert query blocks to queries preserving SQL, source variables and the x/y/w/h layout. The legacy settings JSON string becomes the typed settings object: write each setting you manage as an attribute (for example unit, decimal_places, legend). The move leaves settings out of the moved state, and the first refresh reads them, so nothing on the chart changes. The legacy dashboard_id/chart_id composite id becomes the bare chart id with dashboard_id beside it. created_at and updated_at are gone: Better Stack never returned them. Never recreate a chart: that deletes its alerts. Keep read-only query-builder/PQL/funnel payloads unchanged.

### Resource: `logtail_dashboard_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute. Move team_name to the provider's team_id, or to the resource's own team_id for an object in another team. Import uses the existing id.

### Resource: `logtail_dashboard_section`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_section`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"composite_id","path":"id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep dashboard_id, name, y, collapsed and explanation. The legacy dashboard_id/section_id composite id becomes the bare section id with dashboard_id beside it. created_at and updated_at are gone: Better Stack never returned them. Never create, delete or move a section during the migration: a new or moved section that overlaps others makes Better Stack rearrange the dashboard's charts.

### Resource: `logtail_errors_application`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_application`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"custom_bucket"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"application_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"correlate_with_source_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"application_group_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"correlate_with_source_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute. Turn the one-element custom_bucket block into an object, keeping the existing secret_access_key input. Configure the application's team on the provider and remove obsolete team_name inputs. The computed team_id refreshes as t123. Keep the existing resource state so create-only values are not lost. Import uses the existing ID. Remove name from custom_bucket: Better Stack reads the bucket name from its endpoint.

### Resource: `logtail_errors_application_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_application_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename to betterstack_application_group, keeping the existing group ID, name, sort_index and application references. Replace team_name with the provider's team.

### Resource: `logtail_errors_jira_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_errors_jira_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"notifications"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"json_decode","path":"jira_fields"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"applications[].application_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_errors_jira_integration, keeping the integration ID, which the legacy provider stored as both id and integration_id. Drop integration_id, convert the single notifications block to an object, decode the jira_fields JSON string into its object and replace team_name with the provider's team.

### Resource: `logtail_errors_linear_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_errors_linear_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"notifications"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"applications[].application_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_errors_linear_integration, keeping the integration ID, which the legacy provider stored as both id and integration_id. Drop integration_id, convert the single notifications block to an object and replace team_name with the provider's team.

### Resource: `logtail_errors_microsoft_teams_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_errors_microsoft_teams_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"notifications"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"wrap","path":"routes[].channel_ids","to":"routes[].channels","key":"id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"routes[].applications[].application_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_errors_microsoft_teams_integration, keeping the integration ID, which the legacy provider stored as both id and integration_id. Drop integration_id, convert the single notifications block to an object and replace team_name with the provider's team. Each route's channel_ids become channels entries of {id = ...}, in the same order.

### Resource: `logtail_errors_release`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_release`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename to betterstack_release and preserve the existing release ID, application_id, version and configured environments (Better Stack adds environments to the ones the release already has, and Terraform manages only the ones it sets).
- This resource registers exactly one application and version. Changing the version, application or environment registers a replacement and keeps the earlier releases. Destroying it only removes it from Terraform state: the release stays in Better Stack.

### Resource: `logtail_errors_slack_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_errors_slack_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"notifications"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"applications[].application_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_errors_slack_integration, keeping the integration ID, which the legacy provider stored as both id and integration_id. Drop integration_id, convert the single notifications block to an object and replace team_name with the provider's team.

### Resource: `logtail_exploration`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_exploration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `query` to `queries` in configuration and dependent expressions.
- Rename `variable` to `variables` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"chart"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].name","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].sql_query","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].where_condition","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].static_text","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].source_variable","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"exploration_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute under its current name. Check the exploration's team with `betterstack exploration get ID`, move the existing exploration ID and select that team on the provider instead of team_name. Turn the chart block into an object, write the settings you manage as attributes of the typed chart.settings object (the move leaves them out of the moved state and the first refresh reads them), and rename the query and variable blocks to queries and variables. Keep every SQL, source variable, date range, group and selection setting. Leave out only the empty optional query strings that the legacy provider left out too. chart.name stays a computed output.
- Preserve existing query-builder/PQL/funnel payloads without rewriting them; name-only updates must not replace the queries. Migration does not recreate the exploration or its alert history.

### Resource: `logtail_exploration_alert`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_alert`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `variable_value` to `variable_values` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_omit","path":"paused","value":false}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"escalation_target"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.

Back up state with `terraform state pull`. Find the alert with `betterstack alert list --exploration-id EXPLORATION_ID --team t123`, check that exactly one alert has the old name and note its bare alert ID. Rename to betterstack_alert, keeping exploration_id and that alert ID. Remove the old address from state only, with a `removed` block, and import the same alert at its ID. Never let a plan destroy and recreate the alert: destroying it resolves its open incidents and removes its link to past ones. Every attribute keeps its name. Keep recovery_period = -1: the provider stores it as null, which means the incident never resolves on its own. Keep 0 for immediate recovery and every positive duration as it is. Rename variable_value to variable_values and turn the escalation_target block into an object, with a legacy team_name replaced by that team's team_id, or {} for the current team. When the old target is unknown, find it and never fall back to the current team. Run terraform plan and check that every value it reads back matches the configuration.

### Resource: `logtail_exploration_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_exploration_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename to betterstack_exploration_group, retaining the same ID, name and exploration references. Verify the authoritative global team before moving and replace team_name with that provider team.

### Resource: `logtail_metric`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_extracted_metric`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute and the existing prefixed id.

### Resource: `logtail_source`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"custom_bucket"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"source_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"string_id","path":"source_group_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"vrl_transformation","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"vrl_transformation_logs","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"vrl_transformation","to":"vrl_transformation_logs"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"live_tail_pattern","to":"live_tail_pattern_spans","when":{"path":"platform","in":["claude_platform"]}}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"scrape_frequency_secs","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"scrape_request_basic_auth_user","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"scrape_request_basic_auth_password","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute. Run `betterstack source get ID` with the token the provider will use and check its t-prefixed team_id before moving. The legacy numeric team_id and team_name are not IDs to prefix: put the checked team on the provider, or on the source's own team_id for a source in another team.
- Convert custom_bucket from a one-element block to an object, retaining secret_access_key from the existing sensitive input. Keep the original source id. Convert the historical vrl_transformation alias to vrl_transformation_logs; conflicting nonempty values require operator resolution. A claude_platform source receives spans and no logs, so its live_tail_pattern moves to live_tail_pattern_spans. A source that receives neither logs nor spans, such as one on prometheus_scrape, has no Live tail format: drop live_tail_pattern from its configuration. Drop code_mapping_stack_root and code_mapping_source_root: a telemetry source refuses them, they belong to errors applications. Remove name from custom_bucket: Better Stack reads the bucket name from its endpoint.

### Resource: `logtail_source_aws_account`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source_aws_account`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `auto_subscribe_log_groups` to `aws_auto_sub_log_groups` in configuration and dependent expressions.
- Rename to betterstack_source_aws_account, keeping the source ID: the legacy resource used the source ID as its own id and as source_id. Rename auto_subscribe_log_groups to aws_auto_sub_log_groups. aws_account_id, aws_role_arn and aws_external_id move with the state, so keep the configuration that sets them.
- If an aws_cloudformation_stack passes the source's data_region as ClusterId, pass betterstack_source.<name>.cluster instead: data_region reads back as the region, such as us, whose ingest host doesn't resolve. Add lifecycle { ignore_changes = [parameters["SourceToken"]] } to the stack, because AWS returns that parameter hidden and every plan would update the stack.

### Resource: `logtail_source_aws_log_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_aws_log_group_subscription`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the type to betterstack_aws_log_group_subscription and keep source_id, region, name and subscribed unchanged. If subscribed was omitted, set it explicitly to its old default true. Preserve the subscription row ID, not the source_id/subscription_id legacy import string. Destroying this resource removes only this explicit setting: the log group then stays subscribed only when aws_auto_sub_log_groups is on for its betterstack_source_aws_account or it is a default log group.

### Resource: `logtail_source_gcp_project`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source_gcp_project`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename to betterstack_source_gcp_project, keeping the source ID: the legacy resource used the source ID as its own id and as source_id. gcp_account_id, gcp_project_id and gcp_project_number move with the state, so keep the configuration that sets them.

### Resource: `logtail_source_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute: the same row ID, name, sort_index and references. Check the group's team with `betterstack source-group get ID` before moving, and replace team_name with that team on the provider.
