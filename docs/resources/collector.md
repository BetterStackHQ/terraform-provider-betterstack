---
page_title: "betterstack_collector Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A Better Stack Collector agent that ingests logs, metrics, and traces from a host or cluster.
---

# betterstack_collector (Resource)

A Better Stack Collector agent that ingests logs, metrics, and traces from a host or cluster.

~> While `deletion_protection` is `true`, destroying this resource or changing an attribute that replaces it fails with an error, in a plan or in a `pulumi destroy`: deleting it permanently deletes everything stored in it. To delete it on purpose, set `deletion_protection = false` and apply that first. To stop managing it without deleting it, use a `removed` block with `lifecycle { destroy = false }`.

-> Replaces `logtail_collector`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_collector" "example" {
  name     = "Production Kubernetes"
  platform = "kubernetes"
}
```

## Schema

### Required

- `name` (String) Collector name.
- `platform` (String) Where the collector runs. Can't be changed after the collector is created. Valid values: `docker`, `swarm`, `kubernetes`. Can only be set when creating the resource; changing it afterwards replaces the resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `note` (String) A note about the collector. Send `null` to clear it. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `data_region` (String) The data region, as the source picker names it: `europe`, `us`, `singapore` or a dedicated cluster's name. It reads back as that name. The older ids `germany`, `us_east` and `us_west` are also accepted. Set on create only. Omit it to use the organization's default. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `logs_retention` (Number) Days to keep logs. Refused with 422 on a plan with fixed retention, or for a number of days the plan doesn't offer.
- `metrics_retention` (Number) Days to keep metrics. Refused with 422 on a plan with fixed retention, or for a number of days the plan doesn't offer.
- `ingesting_paused` (Boolean) Pause ingestion for this collector. Defaults to `false`.
- `source_vrl_transformation` (String) VRL program Better Stack runs on each log line of the collector's source as it's ingested. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `source_group_id` (Number) Source group the collector's source belongs to. Send `null` or `"0"` to remove it from its group. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `live_tail_pattern` (String) The pattern Live tail shows each log line with.
- `user_vector_config` (String) Your own Vector configuration in YAML, merged into the configuration the collector runs with. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `configuration` (Attributes, see [nested schema](#nestedatt--configuration)) Collector configuration: which components are on, as `true` or `false` flags such as `logs_host`, `logs_docker`, `logs_kubernetes`, `metrics_cluster`, `ebpf_metrics` and `traces_opentelemetry`, plus sampling rates, batching and buffering. `when_full` is `drop_newest` or `block`. The per-namespace and per-service options each hold `log_sampling` and `ingest_traces`. An update merges the components you send into the stored ones and keeps the previous configuration in history.
- `custom_bucket` (Attributes, see [nested schema](#nestedatt--custom_bucket)) Your own S3-compatible bucket to archive data in. Set it on create. `secret_access_key` is never returned. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `connected_source_ids` (List of Number) IDs of the sources connected to this collector's source, for example to link spans and logs. An update replaces all connected sources. A connection goes both ways. Not accepted by the create request; the provider applies it in an immediate follow-up call once the resource exists.
- `service_fields` (Attributes List, see [nested schema](#nestedatt--service_fields)) Log fields that break data down by service, each with a `name_key` and a `type_key`. An update replaces all of them. `type_key` can be `null`. Not accepted by the create request; the provider applies it in an immediate follow-up call once the resource exists.
- `deletion_protection` (Boolean) While true (the default), destroying or replacing this collector fails, in a plan or in a pulumi destroy. Kept in Terraform or Pulumi state only: the API never sees it. Set it to false and apply that before deleting or replacing it on purpose.

### Read-Only

- `id` (String) The id of this collector.
- `secret` (String, Sensitive) Secret the deployed collector authenticates with. Returned on every read. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `status` (String) Collector status.
- `source_id` (Number) ID of the source the collector sends data to.
- `team_id` (String) Team that owns the collector, as its `t`-prefixed Better Stack team ID, such as `t123`.
- `team_name` (String) Name of the team that owns the collector.
- `pinged_at` (String (RFC 3339)) When the collector last checked in.
- `created_at` (String (RFC 3339)) When the collector was created.
- `updated_at` (String (RFC 3339)) When the collector was last updated.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

- `components` (Dynamic)
- `logs_sample_rate` (Number)
- `traces_sample_rate` (Number)
- `vrl_transformation` (String)
- `merge_logs` (Boolean)
- `merge_logs_config` (String)
- `disk_batch_size_mb` (Number)
- `memory_batch_size_mb` (Number)
- `buffer_max_events` (Number)
- `when_full` (String)
- `log_line_length_limit_kb` (Number)
- `namespaces_options` (Dynamic)
- `services_options` (Dynamic)

<a id="nestedatt--custom_bucket"></a>
### Nested Schema for `custom_bucket`

- `name` (String)
- `endpoint` (String)
- `access_key_id` (String)
- `secret_access_key` (String)
- `keep_data_after_retention` (Boolean)

<a id="nestedatt--service_fields"></a>
### Nested Schema for `service_fields`

- `name_key` (String)
- `type_key` (String)

## Endpoints

- `GET /api/v4/collectors`
- `GET /api/v4/collectors/:id`
- `POST /api/v4/collectors`
- `PATCH /api/v4/collectors/:id`
- `DELETE /api/v4/collectors/:id`

-> A failed check returns `errors` as a list of messages, next to `invalid_attributes`.

-> Create and update also accept a deprecated `databases` list of `{id, service_type, host, port, username, password, ssl_mode, tls, _destroy}` objects, which creates, edits or removes database targets. Manage targets through collector targets instead.

-> With an OAuth token, the user needs the sources permission.

## Import

```shell
terraform import betterstack_collector.example 12345
```
