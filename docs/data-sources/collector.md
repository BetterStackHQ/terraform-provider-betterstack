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

### Optional

- `team_id` (String) Team that owns the collector, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `platform` (String) Where the collector runs. Valid values: `docker`, `swarm`, `kubernetes`.
- `note` (String) A note about the collector.
- `data_region` (String) The data region, `europe`, `us`, `singapore` or a dedicated cluster's name from [List data regions](https://betterstack.com/docs/logs/api/data-regions/). It reads back as that name.
- `logs_retention` (Number) Days logs are kept, one of the retention periods your plan offers.
- `metrics_retention` (Number) Days metrics are kept, one of the retention periods your plan offers.
- `ingesting_paused` (Boolean) Whether ingestion is paused for this collector.
- `collector_logs_transformation` (String) VRL program the collector runs on each log line on your hardware, before sending it to Better Stack.
- `collector_metrics_transformation` (String) VRL program the collector runs on each metric on your hardware, before sending it to Better Stack.
- `collector_traces_transformation` (String) VRL program the collector runs on each span on your hardware, before sending it to Better Stack.
- `ingestion_logs_transformation` (String) VRL program Better Stack runs on each log line from the collector as it's ingested.
- `ingestion_traces_transformation` (String) VRL program Better Stack runs on each span from the collector as it's ingested.
- `source_group_id` (String) Source group the collector's source belongs to, or `null` when it's in no group. Get the IDs via [List source groups](https://betterstack.com/docs/logs/api/list-all-existing-source-groups/).
- `live_tail_pattern` (String) The pattern Live tail shows each log line with. Columns are wrapped in braces, such as `PID: {message_json.pid} {level} {message}`. It's `null` when the collector's source doesn't receive logs. See [Formatting logs](https://betterstack.com/docs/logs/using-logtail/formatting-logs/).
- `live_tail_pattern_spans` (String) The pattern Live tail shows each span with. Columns are wrapped in braces, such as `{span.name} {duration}`. It's `null` when the collector's source doesn't receive spans. See [Formatting logs](https://betterstack.com/docs/logs/using-logtail/formatting-logs/).
- `blocked_metrics` (List of String) Metric names marked as spam, which are dropped when they're ingested. Read more about [marking metrics as spam](https://betterstack.com/docs/logs/using-logtail/transforming-ingested-data/mark-metrics-as-spam/).
- `user_vector_config` (String) Your own Vector configuration in YAML, merged into the configuration the collector runs with.
- `configuration` (Attributes) Collector configuration. It says which components are on, plus sampling rates, batching and buffering. An update merges the components it includes into the stored ones. (see [below for nested schema](#nestedatt--configuration))
- `custom_bucket` (Attributes) Your own S3-compatible bucket to archive data in, or `null` when the collector uses Better Stack storage. Read more about [hosting data in your own bucket](https://betterstack.com/docs/logs/host-data-in-your-own-bucket/). Set it on create. An update that sets only `access_key_id` and `secret_access_key` rotates its key pair. (see [below for nested schema](#nestedatt--custom_bucket))
- `connected_source_ids` (List of String) IDs of the sources connected to this collector's source to correlate their logs and traces. A connection goes both ways. The collector's source joins the listed sources and the sources they're already connected to. No existing connection is removed, and `[]` disconnects the collector's source.
- `service_fields` (Attributes List) Log fields that break data down by service. Each entry needs a `name_key`, and `type_key` can be `null` to skip it. The list replaces the existing one, and `[]` removes them all. (see [below for nested schema](#nestedatt--service_fields))
- `secret` (String, Sensitive) Secret the deployed collector authenticates with. Returned on every read.
- `status` (String) The collector's status, worked out from its hosts, one of `pending`, `up`, `degraded`, `unknown`, `partially_up`, `unsupported_version` or `down`.
- `source_id` (String) ID of the source the collector sends data to.
- `installation_guide_url` (String) Link to the installation guide for the collector's platform, with the install command that uses its `secret`.
- `pinged_at` (String (RFC 3339)) When a host of the collector last checked in, or `null` when no host has checked in yet.
- `created_at` (String (RFC 3339)) When the collector was created.
- `updated_at` (String (RFC 3339)) When the collector was last updated.
- `team_name` (String) Name of the team that owns the collector.
- `deletion_protection` (Boolean) Whether destroying or replacing this collector is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.

<a id="nestedatt--configuration"></a>
### Nested Schema for `configuration`

Read-Only:

- `components` (Attributes) Which components the collector runs, each `true` or `false`. (see [below for nested schema](#nestedatt--configuration--components))
- `logs_sample_rate` (Number) Percentage of logs to keep, from 0 to 100. Defaults to 100.
- `traces_sample_rate` (Number) Percentage of traces to keep, from 0 to 100. Defaults to 100, and 0 drops every trace.
- `merge_logs` (Boolean) Whether the collector merges lines that continue a log entry, such as a stack trace, into the entry before them.
- `merge_logs_config` (String) The VRL condition `merge_logs` uses to tell whether a line starts a new log entry. Any other line joins the entry before it.
- `disk_batch_size_mb` (Number) Batching on disk, the size of the collector's disk buffer in MB. Defaults to 3000, and it can't be less than 256.
- `memory_batch_size_mb` (Number) Batching in memory, how much data the collector batches before sending it, in MB. Defaults to 20.
- `buffer_max_events` (Number) How many events the collector's memory buffer holds before it moves them to disk. Defaults to 10000.
- `when_full` (String) What the collector does when its disk buffer is full, `drop_newest` to drop new events or `block` to wait until there's room.
- `log_line_length_limit_kb` (Number) Maximum log line length in kB, from 4 to 128. Longer lines are dropped. Defaults to 8.
- `namespaces_options` (Attributes Map) Options per namespace, keyed by namespace name. A map replaces the stored one, and an empty map clears it. (see [below for nested schema](#nestedatt--configuration--namespaces_options))
- `services_options` (Attributes Map) Options per service, keyed by service name. A map replaces the stored one, and an empty map clears it. (see [below for nested schema](#nestedatt--configuration--services_options))

<a id="nestedatt--configuration--components"></a>
### Nested Schema for `configuration.components`

Read-Only:

- `logs_host` (Boolean) Collect logs from the host's /var/log directory.
- `logs_docker` (Boolean) Collect logs of Docker containers that use the default json-file logging driver.
- `logs_kubernetes` (Boolean) Collect Kubernetes container logs and events.
- `logs_collector_internals` (Boolean) Collect the collector's own logs.
- `metrics_cluster` (Boolean) Collect Kubernetes object state metrics, such as pods, workloads, nodes and volumes.
- `metrics_databases` (Boolean) Former name of `metrics_cluster`. It reads back as `metrics_cluster` does. Write `metrics_cluster` instead.
- `metrics_other_databases` (Boolean) Collect Redis, MongoDB, Memcached and Kafka metrics.
- `metrics_host_os` (Boolean) Collect CPU, memory, disk, filesystem and network metrics of the host.
- `metrics_postgres` (Boolean) Collect PostgreSQL metrics.
- `metrics_mysql` (Boolean) Collect MySQL metrics.
- `metrics_elasticsearch` (Boolean) Collect Elasticsearch metrics.
- `metrics_pgbouncer` (Boolean) Collect PgBouncer metrics.
- `metrics_nginx` (Boolean) Collect Nginx metrics.
- `metrics_apache` (Boolean) Collect Apache metrics.
- `metrics_traefik` (Boolean) Collect Traefik metrics.
- `ebpf_metrics` (Boolean) Collect host metrics with eBPF.
- `ebpf_tracing_basic` (Boolean) Basic distributed tracing with eBPF.
- `ebpf_tracing_full` (Boolean) Full distributed tracing with eBPF.
- `ebpf_java_agent` (Boolean) Trace Java services with eBPF.
- `traces_opentelemetry` (Boolean) Accept OpenTelemetry traces from your SDKs over OTLP.
- `ebpf_red_metrics` (Boolean) Service map and request rate, error and duration metrics with eBPF.
- `metrics_process_statistics` (Boolean) Collect the host's process and thread counts. It can use significant CPU on a host that runs very many processes.

<a id="nestedatt--configuration--namespaces_options"></a>
### Nested Schema for `configuration.namespaces_options`

Read-Only:

- `log_sampling` (Number) Percentage of the namespace's logs to keep, from 0 to 100.
- `ingest_traces` (Boolean) Whether to ingest the namespace's traces.

<a id="nestedatt--configuration--services_options"></a>
### Nested Schema for `configuration.services_options`

Read-Only:

- `log_sampling` (Number) Percentage of the service's logs to keep, from 0 to 100.
- `ingest_traces` (Boolean) Whether to ingest the service's traces.

<a id="nestedatt--custom_bucket"></a>
### Nested Schema for `custom_bucket`

Read-Only:

- `name` (String) The bucket's name, which Better Stack reads from `endpoint`.
- `endpoint` (String) The S3-compatible endpoint URL of the bucket. It can't be changed after the bucket is set.
- `access_key_id` (String) The access key ID Better Stack uses to write to the bucket.
- `secret_access_key` (String) The secret access key for `access_key_id`. Write-only, so it's never returned.
- `keep_data_after_retention` (Boolean) Whether data stays in the bucket after the retention period. It can't be changed after the bucket is set.

<a id="nestedatt--service_fields"></a>
### Nested Schema for `service_fields`

Read-Only:

- `name_key` (String) The field to extract the service name from, such as `service.name`.
- `type_key` (String) The field to extract the service's platform from, such as `service.type`. Better Stack matches platforms like Nginx, PostgreSQL or Redis in its value.
