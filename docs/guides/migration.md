# Migrate to the unified Better Stack Terraform provider

> **Not available yet.** `betterstack terraform migrate` is not in the current `betterstack` CLI. It was part of the previous CLI and has to be ported before the steps below can be followed. The provider's state movers do not depend on it.

The unified provider uses `BetterStackHQ/betterstack`, flat `betterstack_*` types, and the header-versioned Better Stack API. The legacy `BetterStackHQ/better-uptime` and `BetterStackHQ/logtail` providers and APIs remain available, receiving bug fixes. There is no migration deadline.

Migration changes Terraform ownership, not remote infrastructure. Do not accept a migration plan that creates, updates, deletes or replaces remote objects. This guide does not authorize a public provider release; use an explicitly supplied release-ready build until publication is separately approved.

## Before starting

1. Use Terraform 1.8 or newer and a current `betterstack` CLI with migration metadata. Keep the old configuration, `.terraform.lock.hcl`, legacy provider versions and a backend-supported state backup. State and backups contain secrets; keep them private.
2. Stop concurrent Terraform operations and resolve unrelated drift first. `terraform show -json` does not reliably expose every tainted/deposed object, pending replacement, provider alias or configuration expression. The final Terraform plan remains mandatory.
3. Keep the existing API token's organization/team scope. The unified provider uses `api_token` (or `BETTERSTACK_API_TOKEN`) and global `team_id` values such as `t123`. A legacy application's numeric team ID is not necessarily the global ID. Never obtain a team ID by adding `t` to that number.
4. Preserve provider aliases for distinct credentials or teams. CLI OAuth credentials are only for read-only discovery; never put them into Terraform configuration or substitute them for the provider's API token.

## Ask the CLI for migration instructions

```sh
terraform show -json | betterstack terraform migrate --output migration.tf
```

A private saved **show-JSON state document** can be passed with `--state state.json`. This is not a raw backend state file or saved-plan JSON. `--team-id t123` can disambiguate a target team, but cannot override evidence of another owning team.

The helper reads the entire state and performs only read-only API lookups. It does not execute Terraform, rewrite your HCL, edit state, upload state, or apply changes. It resolves ownership using the selected credentials, not a user's default team. It refuses the whole input, writing nothing, for an unknown legacy type/schema, conflicting target addresses, scope or ownership it cannot prove, two addresses resolving to one remote object, missing required values or credential problems.

For proven one-to-one migrations it writes sorted `moved` blocks into a new `migration.tf`; for reviewed guided resources whose current ID a scoped read-only lookup verified it writes a state-only `removed` block plus an `import`. Collector `databases` rows and a calendar's rotation become separate target resources, each imported at the existing child a scoped lookup matched to exactly one row, next to the moved or imported parent. An AWS/GCP source link folds into its migrated source: a `removed` block forgets only the virtual link address, and the source's block carries its readable settings. Generated target configuration goes to a second new file (`--config-output`, default `betterstack.tf` beside it): one `betterstack_*` block per migrated resource, converted from state as the provider's state mover does, with sensitive values replaced by declared sensitive variables, write-only values the API never reads back (source link IDs and credentials, collector target passwords) left out with a comment because the server keeps them, count/for_each values indexed per instance and module resources commented out for their module source. It never overwrites an existing path, and a failure leaves no partial final artifact. Secret values are not copied from state into either file or diagnostics. No legacy objects produces an explicit success message and no file; missing migration metadata instead requires an upgrade or bundle refresh.

A guided object it cannot hand off (no reviewed import, a failed, empty or ambiguous lookup, or state content only its recipe can carry) does not stop the run: the files cover every proven object, and the `migration.tf` header and stderr list each left-over address with its reason and recipe. Those objects stay on the legacy provider, so keep their legacy HCL, provider blocks and credentials, and point their references at the migrated addresses. The command then exits with status 3 instead of 0.

When the file contains `removed` blocks, keep the legacy provider blocks and credentials configured until the migration is applied: Terraform needs them to plan the removals. Remove them afterwards, unless objects stay on the legacy provider.

Read the file before using it. **Moved blocks do not update resource declarations or expressions.** Apply its per-address instructions to your existing configuration, preserving variable references and sensitive inputs. Merge the commented provider snippet into the appropriate existing provider blocks; do not add a second default provider configuration.

Use the target provider minimum version printed by the helper. The bundle and provider must agree on the mappings; an older provider cannot execute mappings from a newer bundle simply because Terraform understands `moved` blocks.

## Configuration changes

Use the generated provider migration guide for the complete source-type inventory and each conversion/recipe. It is generated from the same schema metadata used by the mover and CLI, rather than a second name map in this document.

- Keep effective timing units, defaults, null/sentinel values, nested settings and create-only constants. For example, consolidated vendor integrations require their original `service`; an alert's legacy recovery value `-1` means explicit null, not the default recovery interval.
- Write every setting you want to keep. The unified provider plans each attribute's documented default, so an attribute your configuration leaves out is planned at that default, and a value set only in the web app (a monitor paused there, for example) shows as a change in the first plan. Add it to the configuration before applying.
- Preserve references to computed outputs. A legacy team member's `id` was an e-mail address, whereas the new `id` identifies the member or invitation. Use `email` when that was the intended value.
- Preserve provider HTTP controls: `api_retry_max`, `api_retry_wait_min`, `api_retry_wait_max`, `api_timeout`, `api_rate_limit` and `api_rate_burst`. Provider configuration and alias wiring cannot be recovered from show-JSON state.
- Rewrite data-source declarations and lookup expressions, but do not import data sources. Check that their returned identities and useful outputs still refer to the same objects.
- Do not run global search-and-replace commands over HCL. Strings, module wiring, renamed fields, split resources and expressions require address-specific review.

For managed charts and explorations, `queries` holds the configured query inputs and `queries_result` exposes the complete returned queries, including their read-only IDs. Move references to those IDs to `queries_result`; do not configure server-assigned IDs. Data sources continue to expose the complete returned collection as `queries`.

## Guided cases

Some states need imports, splits, merges, provider-alias wiring or information available only in configuration. For each object the helper cannot hand off it reports the corresponding recipe and leaves the object on the legacy provider. A left-over object is not permission to omit a resource or lose its capabilities.

Follow the generated recipe with a stable backed-up state. Resolve canonical IDs by scoped read-only API requests, then configure the same objects under their new addresses. Use declarative non-destructive state operations where supported; otherwise use the recipe's explicit `terraform state rm` / `terraform import` ordering. `terraform state rm` forgets ownership only; never run a remote deletion to free an address. Do not let old and new managed addresses own the same object simultaneously.

Exported dashboards, metadata owner/key identities, append-only release history, and calendars, collector databases or AWS/GCP source links the helper could not match need their specific preservation steps. Write-only values stay on the server: leave them out of the configuration rather than sending them again, and never fabricate credentials a read cannot recover. Removing a virtual source-facet address must not disconnect the source, and forgetting a release address must not erase release history.

### State-only credential and request-policy handoff

Imports cannot recover write-only credentials or request-only flags. Do not accept a plan that sends
the same credential again: that is still a remote update and can reconnect a collector target.
The generated `scripts/transfer-state.py` carries outgoing webhook integration passwords and the legacy
status-page resource's fixed-position policy. It also still merges AWS/GCP source-facet and
collector inline-target state, but only for a hand migration that must keep those write-only
values in state: `betterstack terraform migrate` leaves them on the server instead and needs no
transfer. It is separate from `betterstack terraform migrate`.

1. Set `umask 077`, stop concurrent runs and save an original raw backup with
   `terraform state pull > before.tfstate` before moving, forgetting or importing anything.
2. Follow the address-specific recipe to move the owner with refresh-only state operations and
   import existing target IDs or forget virtual facet addresses. Do not run a normal apply.
   Pull the resulting raw state to `current.tfstate`.
3. Run the offline handoff with the exact verified addresses and current serial. For example:

   ```sh
   python3 scripts/transfer-state.py \
     --backup before.tfstate --current current.tfstate \
     --source logtail_source_aws_account.link --target betterstack_source.app \
     --expected-serial 12 --output candidate.tfstate
   ```

   Collector splits also require `--child-id` with the verified existing target ID. The script never
   calls Terraform or an API, modifies its inputs, or uploads anything. It creates a new private
   candidate file; conflicting values, mismatched identities and unsupported transfers are errors.
4. Review the candidate locally as secret-bearing state. Confirm the backend still matches the
   current snapshot, then use `terraform state push candidate.tfstate` without `-force`. Never push
   over a concurrent change. Restore the complete equivalent configuration and require a normal
   no-op plan before resuming applies.

## Verify and recover

Run `terraform init` with the reviewed required-provider declarations, then inspect a saved plan. Moves, imports, read-only refreshes and explicit state-only forget operations may appear; the first plan need not be empty. **There must be no migration-induced remote create, update, delete or replacement.** Stop if the plan contains unrelated drift or intended infrastructure changes.

After applying only the reviewed state operations, require a second no-op plan. Verify the existing remote IDs, children, ordering, memberships, linkage and behavior remain unchanged. Keep the backup until that verification is complete.

On lookup/import failure, stop before remote changes. Check the current backend state before recovering through Terraform/backend-supported operations. Do not overwrite state that another process may have changed, and never hand-edit a backend state file.

## Source-type migration recipes

Minimum target provider version: **0.1.0**. Each entry names its supported state schema versions. Data sources require configuration changes, not state imports.

### Data source: `betteruptime_amazon_sns_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_amazon_sns_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_amazon_sns_integration with the same exact name and selected provider team. Update single extraction outputs from list indexing to objects. Rename references to policy_id to escalation_policy_id.

Replace the data-source type and refresh. Exact lookup must find one SNS receiver, not an IncomingWebhook; missing or duplicate names fail. Retain topic_arn, subscription_state and URL output references. Do not import data-source state.

### Data source: `betteruptime_incoming_webhook`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incoming_webhook_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_incoming_webhook_integration and preserve the exact name lookup. Single extraction-field outputs are objects rather than one-element lists; update output indexing. Rename references to policy_id to escalation_policy_id.

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
- Rename to data.betterstack_monitor and retain the exact url filter. Rename references to policy_id to escalation_policy_id.

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
- Rename to data.betterstack_escalation_policy and retain exact name lookup under the same provider team. Rewrite outputs that read policy_group_id, a step's urgency_id or policy_id, or the instructions comment/reminder fields to escalation_policy_group_id, severity_id, escalation_policy_id and their instructions_ names.

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
- Destination: `betterstack_slack_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename to data.betterstack_slack_integration and retain the exact slack_channel_name filter and selected provider team. Zero or multiple matching channels must fail, not pick an arbitrary channel.

Replace the data source type, retain slack_channel_name and all computed-output references, and refresh under the same selected team. Data sources are not imported or moved.

### Data source: `betteruptime_team_member`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: none yet; keep it on the legacy provider.
- Supported state schema versions: `0`.
- Handling: stays on the legacy provider; do not migrate it yet.

Keep this team member data source on the legacy better-uptime provider for now. The unified provider does not manage team members: a pending invitation and a member are looked up by ids from different tables, so a resource holding an invitation id could read, and on destroy remove, a different member with the same id. It moves once Better Stack serves invitations on their own path.

### Resource: `betteruptime_amazon_sns_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_amazon_sns_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"cause_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"title_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"started_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"acknowledged_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"resolved_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the AmazonSnsIntegration id, receiving URL, topic ARN and subscription state; do not recreate the receiver or restart AWS confirmation. Keep every attribute: the legacy provider wrote the same fields on /api/v2/amazon-sns, where escalation_policy_id was policy_id. Convert the five single extraction blocks to objects while keeping rules/custom fields as arrays. Preserve the SNS Subject/title and Message/body extraction defaults and every configured override explicitly. Move team_name to the provider's team_id (or a provider alias for that team). topic_arn and subscription_state remain computed, never writable.

### Resource: `betteruptime_aws_cloudwatch_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_aws_cloudwatch_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_aws_cloudwatch_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/aws-cloudwatch-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_azure_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_azure_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_azure_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/azure-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_catalog_attribute`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_catalog_attribute`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"rename","path":"relation_id","to":"catalog_relation_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename relation_id to catalog_relation_id, retaining the existing parent, name, primary flag, position, and column values. Import uses catalog_relation_id/id. Never delete/recreate the column, which deletes its values.

### Resource: `betteruptime_catalog_record`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_catalog_record`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Conversion: `{"op":"rename","path":"relation_id","to":"catalog_relation_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"group_by","path":"attribute","to":"attributes","key":"attribute","into":"values"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the relation and record ids; relation_id becomes catalog_relation_id. Group legacy attribute blocks by their attribute_id or attribute_name into attributes = [{attribute = {id/name = ...}, values = [...]}], retaining every value and its type/reference. Item types keep their wire names (Policy, Schedule, Endpoint). A Team reference's item_id now reads back as the t-prefixed public id: write it that way.

Read GET /api/v3/catalog/relations/RELATION/records/ID with the legacy relation_id, retaining every per-column value. Back up state, stop concurrent applies, remove only the old address, then import betterstack_catalog_record as RELATION/ID. Review translated whole-set attributes carefully: entries PATCH does not name are cleared. Never call a batch verb or delete/recreate rows or columns; require a no-op plan.

### Resource: `betteruptime_catalog_relation`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_catalog_relation`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Retain the relation id, name, description and match_mode. The relation and all of its columns and records remain organization-owned; import uses the existing relation id.

### Resource: `betteruptime_datadog_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_datadog_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_datadog_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/datadog-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_elastic_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_elastic_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_elastic_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/elastic-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_email_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_email_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"cause_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"acknowledged_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"resolved_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"started_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"title_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute: the legacy provider wrote the same fields on /api/v2/email-integrations, where escalation_policy_id was policy_id. Convert the five single extraction blocks to objects; keep custom fields and rules as ordered arrays. Legacy target_field belongs to the shared webhook schema but has no meaning for an e-mail's from_email/subject/body targets; remove only those placeholders, not match content. Move team_name to the provider's team_id (or a provider alias for that team).

Back up state and stop concurrent applies. Read GET /api/v3/email-integrations/:id under the same team and compare every extraction/rule/channel setting to the translated HCL, preserving effective defaults. Review a state-only old-address removal plus import of betterstack_email_integration with the same id; never call DELETE or POST. Require a no-op plan and unchanged receiving email_address afterward.

### Resource: `betteruptime_google_monitoring_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_google_monitoring_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_google_monitoring_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/google-monitoring-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_grafana_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_grafana_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_grafana_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/grafana-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_heartbeat`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_heartbeat`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Keep every attribute: the legacy provider wrote the same fields on /api/v2/heartbeats, where escalation_policy_id was policy_id. Move team_name to the provider's team_id (or a provider alias for that team). Keep the existing heartbeat id, so the ping URL stays the same.

### Resource: `betteruptime_heartbeat_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_heartbeat_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute: the legacy provider wrote the same /api/v2/heartbeat-groups fields. Move team_name to the provider's team_id (or a provider alias for that team). Import uses the existing group id.

### Resource: `betteruptime_incoming_webhook`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_incoming_webhook_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"cause_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"title_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"started_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"acknowledged_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"resolved_alert_id_field"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the existing receiver id and URL; do not recreate its secret receiving token. Keep every attribute: the legacy provider wrote the same fields on /api/v2/incoming-webhooks, where escalation_policy_id was policy_id. Convert the five single extraction blocks to nullable objects, preserving extraction target, match type and content; keep custom fields and event rules as ordered arrays. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_jira_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_jira_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `jira_fields_json` to `jira_fields` in configuration and dependent expressions.
- Conversion: `{"op":"json_decode","path":"jira_fields"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename better_stack_id in HCL to the required existing-integration id input; legacy state already stores this same id. Decode jira_fields_json into the typed jira_fields object, retaining all nested Jira values. This resource adopts a connected Jira integration by GET and manages it by PATCH; create never provisions a new OAuth connection and destroy only removes Terraform state. Keep name unchanged.

### Resource: `betteruptime_metadata`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_metadata`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `metadata_value` to `values` in configuration and dependent expressions.
- Identity is the scoped owner_type/owner_id/key group the legacy provider already read through /api/v3/metadata. Rename metadata_value blocks to values without losing order, string contents or item references; a deprecated scalar value becomes one {type = "String", value = ...} entry. Item types keep their wire names (Policy, Schedule, Endpoint). A Team reference's item_id now reads back as the t-prefixed public id: write it that way. Keep owner ids and non-Team referenced ids as they are.

Read GET /api/v4/metadata?owner_type=TYPE&owner_id=OWNER with the proved provider team and pick the entry whose key matches. Require exactly one matching owner/key group and use its returned id, even when it differs from old state. Back up state, stop concurrent applies, remove only the old state address, then import betterstack_metadata at that id. No DELETE/POST/PATCH during migration: deleting removes the entire key and all its values. Retain all values in HCL and require a no-op plan.

### Resource: `betteruptime_monitor`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_monitor`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"monitor_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"expiration_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute: the legacy provider wrote the same fields on /api/v2/monitors, where escalation_policy_id was policy_id. Its 0 for an unset monitor_group_id or expiration_policy_id becomes null, as the API reads it. Move team_name to the provider's team_id (or a provider alias for that team). Preserve all request headers by name/value, dropping only the obsolete header row ids. Import uses the existing id.

### Resource: `betteruptime_monitor_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_monitor_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute: the legacy provider wrote the same /api/v2/monitor-groups fields. Move team_name to the provider's team_id (or a provider alias for that team). Import uses the existing group id.

### Resource: `betteruptime_new_relic_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_new_relic_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_new_relic_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/new-relic-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_on_call_calendar`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_on_call_schedule`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `default_calendar` to `default_schedule` in configuration and dependent expressions.
- Keep the existing schedule id and name: the legacy provider wrote the same fields on /api/v2/on-calls, where default_schedule was default_calendar. Move team_name to the provider's team_id or alias. Split the legacy on_call_rotation block into a betterstack_on_call_rotation resource addressed by the schedule id. Preserve user ordering, length, rotation_start, rotation_end, timezone and working_hours. Preserve an omitted rotation rather than creating one.

Back up state and stop concurrent applies. Read GET /api/v3/on-call-schedules/:id and GET /api/v3/on-call-schedules/:id/rotation under the same selected team. Review state-only removal of the old calendar address, import betterstack_on_call_schedule at the original calendar id, and, when the legacy state holds a rotation and the read returns one, import betterstack_on_call_rotation at the same calendar id. POST would replace the rotation and regenerate future shifts, and DELETE would remove shifts: neither is a migration operation. Keep the original schedule events and require a no-op plan.

### Resource: `betteruptime_outgoing_webhook`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_outgoing_webhook_integration`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Conversion: `{"op":"singleton","path":"custom_webhook_template_attributes"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep the NativeWebhook id, trigger_type, destination URL, notify_alongside_primary_responder and every event toggle explicitly. Unwrap the single custom_webhook_template_attributes block into an object, preserving method, ordered headers, body template and basic-auth input expressions. Preserve body_template as its stored string, including JSON encoding and template variables; do not drop non-JSON template strings. Move team_name to the provider's team_id (or a provider alias for that team).
- Keep any custom_webhook_template_attributes auth_password in HCL as its existing sensitive input expression; never copy it from state or show-JSON. A webhook whose legacy state holds a template password is not auto-imported and needs the state-transfer recipe.

Stop concurrent applies and save an ORIGINAL terraform state pull backup privately before removing any address. Read GET /api/v2/outgoing-webhooks/:id and compare every event toggle and request-template value. Keep the original sensitive auth_password expression in HCL so an ordinary apply does not reassert a different password. Review terraform state rm of the legacy address, then CLI terraform import of the new resource at the existing id (read-only; do not apply an import block). Save a CURRENT terraform state pull. For a configured legacy template password, run python3 scripts/transfer-state.py --backup ORIGINAL_RAW_PULL.json --current CURRENT_RAW_PULL.json --source EXACT_LEGACY_ADDRESS --target EXACT_TARGET_ADDRESS --expected-serial INTEGER --output NEW_CANDIDATE.json. This fixed recipe copies only the original singleton template auth_password into the already-imported template and marks its nested state path sensitive; missing original credentials, changed identities or conflicting target values are refused. Review the candidate privately, re-pull and compare backend lineage/serial, then explicitly install it with Terraform's supported state push and locking, never force. Require an ordinary no-op plan before any apply. The helper never contacts the API or writes the backend. No DELETE/POST/PATCH or credential reassertion is part of this migration.

### Resource: `betteruptime_pagerduty_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_pagerduty_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the resource to betterstack_pagerduty_integration. Preserve name, routing key, severity and notify_alongside_primary_responder, including the original sensitive input expression. Move team_name to the provider's team_id (or a provider alias for that team). Import uses the existing id.

### Resource: `betteruptime_policy`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_escalation_policy`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `policy_group_id` to `escalation_policy_group_id` in configuration and dependent expressions.
- Conversion: `{"op":"rename","path":"steps[].urgency_id","to":"steps[].severity_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].policy_id","to":"steps[].escalation_policy_id"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].comment","to":"steps[].instructions_comment"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].reminder_enabled","to":"steps[].instructions_reminder_enabled"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"steps[].reminder_interval_hours","to":"steps[].instructions_reminder_interval_hours"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"steps[].instructions_reminder_enabled","value":false}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"steps[].instructions_reminder_interval_hours","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_escalation_policy and keep every policy attribute: the legacy provider wrote the same policy fields, on /api/v3. Rename policy_group_id to escalation_policy_group_id. In ordered steps rename urgency_id to severity_id, policy_id to escalation_policy_id, and comment/reminder_enabled/reminder_interval_hours to instructions_comment/instructions_reminder_enabled/instructions_reminder_interval_hours. An entire_team member that names a team now reads its id as the t-prefixed team id (t123) instead of the local number. Move team_name to the provider's team_id or alias.

Read GET /api/v5/escalation-policies/:id and compare the full ordered steps to translated HCL. Keep wait-until timezones, time-branch windows, resolved action_type, typed metadata_values/no_value conditions and instructions reminders. Rewrite an entire_team member's numeric id to the t-prefixed team id the read returns; the own-team member keeps no id. Convert deprecated metadata_values strings or metadata_value blocks to typed metadata_values, Team references as t-prefixed ids. Back up state, stop concurrent applies, review state-only old-address removal plus same-id import; no PATCH that rewrites steps and no DELETE. Require a no-op plan.

### Resource: `betteruptime_policy_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_escalation_policy_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the resource to betterstack_escalation_policy_group and keep name and sort_index: the legacy provider wrote the same fields on /api/v2/policy-groups. Move team_name to the provider's team_id or provider alias. Import uses the existing group id.

### Resource: `betteruptime_prometheus_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_prometheus_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `policy_id` to `escalation_policy_id` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"escalation_policy_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Rename to betterstack_prometheus_integration. Keep the integration id, webhook_url, channels and pause state: the legacy provider wrote the same fields on /api/v2/prometheus-integrations, where escalation_policy_id was policy_id. A policy_id of 0 becomes a null escalation_policy_id. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_severity`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_severity`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep the existing severity id, name and every notification channel. Move team_name to the provider's team_id (or a provider alias for that team).

### Resource: `betteruptime_severity_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_severity_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute: the legacy provider wrote the same fields on /api/v2/urgency-groups. Move team_name to the provider's team_id (or a provider alias for that team). Import uses the existing group id.

### Resource: `betteruptime_splunk_oncall_integration`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_splunk_on_call_integration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the resource to betterstack_splunk_on_call_integration. Preserve name, destination URL and notify_alongside_primary_responder. Move team_name to the provider's team_id (or a provider alias for that team). The id remains the existing NativeWebhook id; do not migrate it as an outgoing_webhook_integration.

### Resource: `betteruptime_status_page`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"sentinel_null","path":"status_page_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute: the legacy provider wrote the same /api/v2/status-pages fields. Its status_page_group_id of zero means no group and becomes null. Keep password in its original sensitive input expression, never copy it into migration output; read refresh must retain the write-only value. Import uses the existing page id.

### Resource: `betteruptime_status_page_group`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page_group`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Keep name and sort_index: the legacy provider wrote the same /api/v2/status-page-groups fields. Drop team_name, which status page groups never used: they belong to an organization, and organization ids are not team ids.

Read GET /api/v2/status-page-groups/:id using the selected credential. Back up state, stop concurrent applies, remove only the old state address without destroying the group, then import betterstack_status_page_group with the same id. Review a no-remote-mutation plan before applying state operations, then require a no-op plan.

### Resource: `betteruptime_status_page_resource`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page_resource`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Conversion: `{"op":"singleton","path":"mark_as_degraded_metadata_rule"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"mark_as_degraded_metadata_rule.metadata_value","to":"mark_as_degraded_metadata_rule.values"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"singleton","path":"mark_as_down_metadata_rule"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"mark_as_down_metadata_rule.metadata_value","to":"mark_as_down_metadata_rule.values"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"status_page_section_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep status_page_id, status_page_section_id, resource_id/type, public_name, explanation, widget_type and position: the legacy provider wrote the same fields under /api/v2. Drop fixed_position: the target always sends fixed_position = true, as the legacy provider did, and has no attribute for it, so a position never shifts siblings. Convert single metadata-rule blocks to objects and their metadata_value blocks to values, retaining typed references. Import keeps the legacy status_page_id/id form.

Read GET /api/v3/status-pages/:status_page_id/resources/:id, verifying parent page, underlying resource and metadata rules. v3 reads Team metadata values as t-prefixed global team ids where the legacy provider stored the internal number: rewrite each such reference to its t-prefixed id; preserve all condition values. Stop concurrent applies and save an ORIGINAL terraform state pull backup. Review state-only terraform state rm of the legacy address, then CLI terraform import of the target at status_page_id/id; do not apply an import block. The target always sends fixed_position = true, as the reviewed legacy SDK did, so import alone keeps the item's position policy and no state needs seeding. Require an ordinary no-op plan before any apply. No API writes are part of this handoff: never remove/re-add the item, PATCH it to reassert policy, or shift sibling positions; the underlying resource remains unchanged.

### Resource: `betteruptime_status_page_section`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: `betterstack_status_page_section`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep status_page_id, name and position. Drop fixed_position: the target always sends fixed_position = true, as the legacy provider did, and has no attribute for it. Import keeps the legacy status_page_id/id form. Never shift siblings as a migration side effect.

### Resource: `betteruptime_team_member`

- Source provider: `registry.terraform.io/BetterStackHQ/better-uptime`.
- Destination: none yet; keep it on the legacy provider.
- Supported state schema versions: `0`.
- Handling: stays on the legacy provider; do not migrate it yet.

Keep this team member on the legacy better-uptime provider for now. The unified provider does not manage team members: a pending invitation and a member are looked up by ids from different tables, so a resource holding an invitation id could read, and on destroy remove, a different member with the same id. It moves once Better Stack serves invitations on their own path.

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

Rename to data.betterstack_connection with the same connection id and verified access scope. GET never reveals password; retain any independently stored credential and do not recreate the connection. team_ids now read as t-prefixed ids. Refresh only; do not import data sources.

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

Rename to data.betterstack_chart_alert and retain dashboard_id, the bare chart_id and exact name as intersecting lookup constraints. Duplicate names require the explicit returned bare alert ID. Rename variable_value accesses to variable_values and singleton escalation_target accesses to the polymorphic string/object output. Explicit null recovery means never; zero means immediate. Refresh only; do not import or recreate an alert.

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

Rename to data.betterstack_chart_alert, keeping exploration_id plus exact name as intersecting lookup constraints; select the authoritative bare alert ID if names duplicate. An exploration's alert carries no chart_id. Rename variable_value accesses to variable_values and singleton target accesses to the polymorphic escalation_target. Null recovery retains never-auto-recover. Refresh only; never import or recreate an alert.

### Data source: `logtail_exploration_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_exploration_group`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_exploration_group and preserve exact name lookup in the verified team. Duplicate names require an explicit existing ID; refresh only, never import the data source.

### Data source: `logtail_metric`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_metric`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

Rename to data.betterstack_metric_expression and preserve both source_id and exact name. Require exactly one rule matching the name on that source; duplicate names fail closed. Keep SQL/type/aggregation outputs, refresh only and never import a data source.

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
- Conversion: `{"op":"rename","path":"configuration.components.metrics_databases","to":"configuration.components.metrics_cluster","on_conflict":"target"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"index_by","path":"configuration.service_option","to":"configuration.services_options","key":"name"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"index_by","path":"configuration.namespace_option","to":"configuration.namespaces_options","key":"name"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"source_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Read GET /api/v4/collectors/:id to verify the authoritative global team before moving. Keep the collector ID and existing sensitive state: the custom_bucket secret_access_key cannot be recovered by importing the collector. Turn configuration, components and custom_bucket singleton blocks into objects. Replace service_option and namespace_option sets with services_options/namespaces_options maps keyed by name, preserving numeric log_sampling and boolean ingest_traces. metrics_cluster takes precedence over deprecated metrics_databases exactly as in the legacy provider.
- A populated databases list becomes one betterstack_collector_target per row, imported at its existing target ID next to the moved collector, with collector_id referencing it. Leave password and api_key out: they are write-only, the API keeps them server-side and never returns them. Never apply removal of inline blocks through the legacy provider, which deletes targets.
- Preserve source_id, source_group_id, live_tail_pattern, retention, Vector configuration, component flags and both VRL programs. Migration must produce no remote create, update or delete, including credential reassertion.

betterstack terraform migrate moves the collector and imports one betterstack_collector_target per legacy databases row when GET /api/v1/collectors/<existing collector ID>/targets in the verified team lists exactly one target with the same ID, kind (legacy service_type), host and port for each row. Otherwise the collector stays on the legacy provider: resolve missing, duplicate or mismatched targets and rerun it, never inferring identity from position. By hand: stop concurrent applies and keep a raw terraform state pull backup outside the repository. Move logtail_collector.NAME to betterstack_collector.NAME with a moved block; never import or recreate the collector, whose bucket credentials live only in state. Configure one betterstack_collector_target per database with collector_id = betterstack_collector.NAME.id and the existing kind, host, port, username, ssl_mode and tls, leaving password and api_key out (write-only: the API keeps them server-side), and import each at its verified target ID with an import block. Never apply removal of the inline databases through the legacy provider: it deletes the targets. Require a plan with no remote create, update, delete or credential change.

### Resource: `logtail_collector_target`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_collector_target`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"sentinel_null","path":"host","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"collector_host","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"collector_host","to":"host"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Verify GET /api/v1/collectors/:collector_id/targets/:id under the authoritative team and confirm kind before moving. Keep collector_id, kind, enabled, port, service, listen_ip, endpoint, scheme, username, password, api_key, SSL/TLS, status and container. For nginx/apache/kafka/traefik/prometheus rename collector_host to host; database host stays host. Move existing state to retain write-only credentials and never recreate a target or duplicate one already split from a collector's inline databases.

### Resource: `logtail_connection`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_connection`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"sentinel_null","path":"valid_until","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Read GET /api/v2/connections/:id with credentials authorized for the full connection: its team_ids are the t-prefixed ids of the teams the legacy numeric team_ids named. Replace a legacy team_ids configuration with precisely that set (or keep team_names as it is); never use a single selected team as a substitute for a multi-team connection or prefix app-local IDs yourself.
- Move the existing state; never remove/import or recreate this credential: password is only returned on creation and must remain in sensitive state. Keep host, port, username, password, sample_query, data_sources, client_type, data_region, IP allowlist, expiry, note and creator. Preserve configured data_region and valid_until rather than API-normalized variants. Verify no credential rotation or replacement in the plan.

### Resource: `logtail_dashboard`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `variable` to `variables` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"dashboard_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"variables[].sql_definition","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"variables[].default_values","value":[]}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute: the legacy provider wrote the same /api/v2/dashboards fields this /api/v3 resource takes. Rename variable blocks to the variables object list, and a dashboard_group_id of 0 to null. Move team_name to the provider's team_id (or a provider alias for that team); team_id now reads back as the t-prefixed id. Import uses the existing id.
- A dashboard the legacy provider created from data (the import API) keeps its id, charts, queries, alerts, sections and variables: drop data from the configuration, since this resource takes no exported document and must not call import again. Its charts and sections stay unmanaged unless imported as their own resources, and no child may be imported twice.

### Resource: `logtail_dashboard_alert`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_alert`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `variable_value` to `variable_values` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_omit","path":"paused","value":false}`. Follow the configuration instructions below; state conversion does not rewrite HCL.

Back up state, verify the global team, and resolve the old dashboard/chart/alert composite identity with GET /api/v2/alerts?dashboard_id=<dashboard_id>&chart_id=<bare chart_id>; require the returned alert's name, dashboard_id and chart_id to match. Rename to betterstack_chart_alert, keeping the same dashboard_id, bare chart_id and alert id. State-only forget the old address and import the same alert; never create/delete it or discard incident history. Every attribute keeps its wire name: the legacy provider wrote the same /api/v2 alert fields. Keep recovery_period=-1: it still means never auto-recover, retaining 0 as immediate and every positive duration unchanged. Rename the variable_value blocks to variable_values, preserving selected_label, values and additional_conditions. Convert escalation_target from a singleton block to its wire form: an object with policy_id/policy_name or team_id/team_name, or the literal current_team when no block was set. Never convert unknown/stale routing to current_team: resolve the actual target first. Keep metadata values as strings or lists of strings; keep source modes/platforms/variables, incident cause/per-series, notification flags, anomaly settings, missing-data behavior and paused state. Review full readback and require a subsequent no-op plan.

### Resource: `logtail_dashboard_chart`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_chart`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `query` to `queries` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_null","path":"settings","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"json_decode","path":"settings"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].name","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].sql_query","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].where_condition","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].static_text","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].source_variable","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute: the legacy provider wrote the same /api/v2/dashboards/:dashboard_id/charts fields. Convert query blocks to queries and jsondecode the legacy settings JSON string, preserving booleans/numbers/arrays, SQL, source variables and the x/y/w/h layout. The legacy dashboard_id/chart_id composite id becomes the bare chart id with dashboard_id beside it. created_at and updated_at are gone: the API never returned them. Never recreate a chart: that deletes its alerts. Keep read-only query-builder/PQL/funnel payloads unchanged.

### Resource: `logtail_dashboard_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute: the legacy provider wrote the same /api/v2/dashboard-groups fields. Move team_name to the provider's team_id (or a provider alias for that team). Import uses the existing id.

### Resource: `logtail_dashboard_section`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_dashboard_section`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep dashboard_id, name, y, collapsed and explanation: the legacy provider wrote the same /api/v2/dashboards/:dashboard_id/sections fields. The legacy dashboard_id/section_id composite id becomes the bare section id with dashboard_id beside it. created_at and updated_at are gone: the API never returned them. No create, delete or reposition is part of migration: these shift chart layout.

### Resource: `logtail_errors_application`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_application`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"custom_bucket"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"application_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"correlate_with_source_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute: the legacy provider wrote the same application fields. Turn the one-element custom_bucket block into an object, preserving the existing secret_access_key input. Configure the application's team on the provider and remove obsolete team_name inputs; computed team_id refreshes from the API as t123. Retain the existing resource state so create-only values are not lost. Import uses the existing id.

### Resource: `logtail_errors_application_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_application_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename to betterstack_application_group, keeping the existing group ID, name, sort_index and application references. Replace team_name with the provider's team.

### Resource: `logtail_errors_release`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_release`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename to betterstack_release and preserve the existing release ID, application_id, version and configured environments (the API merges environments, Terraform does not take ownership of the server union).
- This resource registers exactly one application/version. Version/application/environment changes register a replacement while retaining historical releases. Destroy only forgets state and never sends DELETE; no cleanup call should be added.

### Resource: `logtail_exploration`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_exploration`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename `query` to `queries` in configuration and dependent expressions.
- Rename `variable` to `variables` in configuration and dependent expressions.
- Conversion: `{"op":"singleton","path":"chart"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"chart.settings","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"json_decode","path":"chart.settings"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].name","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].sql_query","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].where_condition","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].static_text","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_omit","path":"queries[].source_variable","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"exploration_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute's wire name: the legacy provider wrote the same /api/v2/explorations fields. Verify GET /api/v2/explorations/:id under the authoritative global team, move the existing exploration ID and select that team on the provider instead of team_name. Convert chart from a singleton block into an object, jsondecode chart.settings without stringifying booleans/numbers/arrays, and rename the query/variable blocks to queries/variables. Keep every SQL/source variable/date-range/group/selection setting. Omit only empty optional query strings that the legacy provider omitted on the wire. chart.name remains a computed output.
- Preserve existing query-builder/PQL/funnel payloads without rewriting them; name-only updates must not replace the queries. Migration does not recreate the exploration or its alert history.

### Resource: `logtail_exploration_alert`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_alert`.
- Supported state schema versions: `0`.
- Handling: guided configuration; `betterstack terraform migrate` verifies the current ID and emits `removed` (state-only) plus `import` blocks.
- Rename `variable_value` to `variable_values` in configuration and dependent expressions.
- Conversion: `{"op":"sentinel_omit","path":"paused","value":false}`. Follow the configuration instructions below; state conversion does not rewrite HCL.

Back up state and resolve the old exploration/alert composite identity with GET /api/v2/alerts?exploration_id=<existing exploration_id> in the verified global team; require matching exploration_id and name and keep the returned bare alert ID. The managed betterstack_chart_alert creates through a dashboard chart (POST /api/v2/dashboards/:dashboard_id/charts/:chart_id/alerts), which an exploration's chart is not, so an exploration alert stays on the legacy provider or is managed through the API directly; never delete and recreate it, which discards its incident history. Every attribute keeps its wire name. recovery_period=-1 becomes explicit null, 0 remains immediate and positive durations remain unchanged. Rename variable_value to variable_values and convert the singleton escalation_target to its wire object or the literal current_team; resolve unknown routing rather than defaulting it. Verify exact readback.

### Resource: `logtail_exploration_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_exploration_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename to betterstack_exploration_group, retaining the same ID, name and exploration references. Verify the authoritative global team before moving and replace team_name with that provider team.

### Resource: `logtail_metric`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_metric`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute and the existing prefixed id: the legacy provider called the same /api/v2/sources/:source_id/metrics endpoints.

### Resource: `logtail_source`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Conversion: `{"op":"singleton","path":"custom_bucket"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"source_group_id","value":0}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"vrl_transformation","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"sentinel_null","path":"vrl_transformation_logs","value":""}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Conversion: `{"op":"rename","path":"vrl_transformation","to":"vrl_transformation_logs"}`. Follow the configuration instructions below; state conversion does not rewrite HCL.
- Keep every attribute: the legacy provider wrote the same source fields to /api/v2/sources. Resolve the source with a read of /api/v4/sources/:id under the intended credential and verify its t-prefixed team_id before moving. Legacy numeric team_id and team_name are obsolete scope representations, not ids to prefix: put the verified team on the provider (using aliases for separate teams); computed team_id refreshes from the API.
- Convert custom_bucket from a one-element block to an object, retaining secret_access_key from the existing sensitive input. Keep the original source id. Convert the historical vrl_transformation alias to vrl_transformation_logs; conflicting nonempty values require operator resolution. Drop code_mapping_stack_root and code_mapping_source_root: a telemetry source refuses them, they belong to errors applications.
- A logtail_source_aws_account or logtail_source_gcp_project on this source folds into this block: its readable settings join it, its write-only linkage stays on the server; never import a second owner.

### Resource: `logtail_source_aws_account`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.
- Rename `auto_subscribe_log_groups` to `aws_auto_sub_log_groups` in configuration and dependent expressions.

betterstack terraform migrate folds this link into its owning source when the logtail_source whose id equals source_id is migrated from the same state and module: a removed block with destroy = false forgets only this virtual address and the source's generated betterstack_source block carries its readable settings. Otherwise it stays on the legacy provider; migrate it together with its owning source. Configure that betterstack_source with aws_auto_sub_log_groups (renamed from auto_subscribe_log_groups, value unchanged) and leave aws_account_id, aws_role_arn and aws_external_id out: they are write-only, the API keeps the existing linkage server-side and never returns it. Forget only logtail_source_aws_account.NAME with removed { from = logtail_source_aws_account.NAME lifecycle { destroy = false } }; never delete or disconnect it through the legacy provider, and never import a second owner of the source. Replace references to its ID with the source's id and require a plan with no remote create, update, delete or replacement.

### Resource: `logtail_source_aws_log_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source_aws_log_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Rename the type to betterstack_source_aws_log_group and keep source_id, region, name and subscribed unchanged: the legacy provider called the same /api/v2/sources/:source_id/aws-log-group-subscriptions endpoints. If subscribed was omitted, set it explicitly to its old default true. Preserve the subscription row ID, not the source_id/subscription_id legacy import string. Destroying this resource removes only the explicit override and restores automatic behavior.

### Resource: `logtail_source_gcp_project`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source`.
- Supported state schema versions: `0`.
- Handling: guided configuration/state transition.

betterstack terraform migrate folds this link into its owning source when the logtail_source whose id equals source_id is migrated from the same state and module: a removed block with destroy = false forgets only this virtual address. Otherwise it stays on the legacy provider; migrate it together with its owning source. Configure that betterstack_source without gcp_account_id, gcp_project_id and gcp_project_number: they are write-only, the API keeps the existing linkage server-side and never returns it. Forget only logtail_source_gcp_project.NAME with removed { from = logtail_source_gcp_project.NAME lifecycle { destroy = false } }; never delete or disconnect it through the legacy provider, and never import a second owner of the source. Replace references to its ID with the source's id and require a plan with no remote create, update, delete or replacement.

### Resource: `logtail_source_group`

- Source provider: `registry.terraform.io/BetterStackHQ/logtail`.
- Destination: `betterstack_source_group`.
- Supported state schema versions: `0`.
- Handling: one-to-one moved state, subject to whole-state and ownership validation.
- Keep every attribute: the legacy provider wrote the same /api/v1/source-groups fields. Keep the same row id, name, sort_index and references; verify the selected global team with an authoritative read before moving, and replace team_name with that provider team.
