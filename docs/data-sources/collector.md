---
page_title: "betterstack_collector Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A Better Stack Collector agent that ingests logs, metrics, and traces from a host or cluster.
---

# betterstack_collector (Data Source)

A Better Stack Collector agent that ingests logs, metrics, and traces from a host or cluster.

## Example Usage

```terraform
data "betterstack_collector" "example" {
  name = "Production Kubernetes"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Collector name.

### Read-Only

- `platform` (String) Where the collector runs. Can't be changed after the collector is created. Valid values: `docker`, `swarm`, `kubernetes`.
- `note` (String) A note about the collector. Send `null` to clear it.
- `data_region` (String) The data region, as the source picker names it: `europe`, `us`, `singapore` or a dedicated cluster's name. It reads back as that name. The older ids `germany`, `us_east` and `us_west` are also accepted. Set on create only. Omit it to use the organization's default.
- `logs_retention` (Number) Days to keep logs. Refused with 422 on a plan with fixed retention, or for a number of days the plan doesn't offer.
- `metrics_retention` (Number) Days to keep metrics. Refused with 422 on a plan with fixed retention, or for a number of days the plan doesn't offer.
- `ingesting_paused` (Boolean) Pause ingestion for this collector.
- `source_vrl_transformation` (String) VRL program Better Stack runs on each log line of the collector's source as it's ingested.
- `source_group_id` (String) Source group the collector's source belongs to. Send `null` or `"0"` to remove it from its group.
- `live_tail_pattern` (String) The pattern Live tail shows each log line with.
- `user_vector_config` (String) Your own Vector configuration in YAML, merged into the configuration the collector runs with.
- `configuration` (Attributes, see [nested schema](#nestedatt--configuration)) Collector configuration: which components are on, as `true` or `false` flags such as `logs_host`, `logs_docker`, `logs_kubernetes`, `metrics_cluster`, `ebpf_metrics` and `traces_opentelemetry`, plus sampling rates, batching and buffering. `when_full` is `drop_newest` or `block`. The per-namespace and per-service options each hold `log_sampling` and `ingest_traces`. An update merges the components you send into the stored ones and keeps the previous configuration in history.
- `custom_bucket` (Attributes, see [nested schema](#nestedatt--custom_bucket)) Your own S3-compatible bucket to archive data in. Set it on create. `secret_access_key` is never returned.
- `connected_source_ids` (List of String) IDs of the sources connected to this collector's source, for example to link spans and logs. An update replaces all connected sources. A connection goes both ways.
- `service_fields` (Attributes List, see [nested schema](#nestedatt--service_fields)) Log fields that break data down by service, each with a `name_key` and a `type_key`. An update replaces all of them. `type_key` can be `null`.
- `secret` (String, Sensitive) Secret the deployed collector authenticates with. Returned on every read.
- `status` (String) Collector status.
- `source_id` (String) ID of the source the collector sends data to.
- `team_id` (String) Team that owns the collector, as its `t`-prefixed Better Stack team ID, such as `t123`.
- `team_name` (String) Name of the team that owns the collector.
- `pinged_at` (String (RFC 3339)) When the collector last checked in.
- `created_at` (String (RFC 3339)) When the collector was created.
- `updated_at` (String (RFC 3339)) When the collector was last updated.
- `deletion_protection` (Boolean) While true (the default), destroying or replacing this collector fails, in a plan or in a pulumi destroy. Kept in Terraform or Pulumi state only: the API never sees it. Set it to false and apply that before deleting or replacing it on purpose.

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
