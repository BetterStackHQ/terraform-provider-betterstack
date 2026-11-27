---
page_title: "betterstack_source Resource - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A telemetry source, a stream of logs, metrics, or traces ingested into Better Stack.
---

# betterstack_source (Resource)

A telemetry source, a stream of logs, metrics, or traces ingested into Better Stack.

~> Destroying a source deletes its data for good, so `deletion_protection` is on by default: Terraform refuses to destroy or replace it until you set `deletion_protection = false` and apply.

-> Replaces `logtail_source`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a source

```terraform
resource "betterstack_source" "example" {
  name        = "My new HTTP source"
  platform    = "http"
  data_region = "us"
}
```

### In a source group

```terraform
resource "betterstack_source_group" "example" {
  name = "Backend sources"
}

resource "betterstack_source" "in_source_group" {
  name            = "My new HTTP source"
  platform        = "http"
  data_region     = "us"
  source_group_id = betterstack_source_group.example.id
}
```

## Schema

### Required

- `name` (String) Source name.
- `platform` (String, Forces replacement) The platform the source ingests from. For a collector, use [Create a collector](https://betterstack.com/docs/logs/api/create-a-collector/), which creates its own source. Can't be changed after the source is created. See the [57 valid values](#values--platform).

### Optional

- `data_region` (String, Forces replacement) The data region, `europe`, `us`, `singapore` or a dedicated cluster's name from [List data regions](https://betterstack.com/docs/logs/api/data-regions/). It reads back as that name. It can't be changed after the source is created. Defaults to the organization's default data region.
- `source_group_id` (String) Source group this source belongs to, or `null` when it's in no group. Get the IDs via [List source groups](https://betterstack.com/docs/logs/api/list-all-existing-source-groups/).
- `ingesting_paused` (Boolean) Whether ingestion is paused for this source. Defaults to `false`.
- `logs_retention` (Number, Computed) Days logs are kept, one of the retention periods your plan offers.
- `metrics_retention` (Number, Computed) Days metrics are kept, one of the retention periods your plan offers.
- `live_tail_pattern` (String, Computed) The pattern Live tail shows each log line with. Columns are wrapped in braces, such as `PID: {message_json.pid} {level} {message}`. It's `null` when the source doesn't receive logs. See [Formatting logs](https://betterstack.com/docs/logs/using-logtail/formatting-logs/). A source that doesn't receive logs refuses it.
- `live_tail_pattern_spans` (String, Computed) The pattern Live tail shows each span with. Columns are wrapped in braces, such as `{span.name} {duration}`. It's `null` when the source doesn't receive spans. See [Formatting logs](https://betterstack.com/docs/logs/using-logtail/formatting-logs/). A source that doesn't receive spans refuses it.
- `team_id` (String) Team that owns the source, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `vrl_transformation_logs` (String, Computed) VRL program Better Stack runs on each log line as it's ingested. `null` restores the platform's default program, and `""` is an empty program. Read more about [VRL transformations](https://betterstack.com/docs/logs/using-logtail/transforming-ingested-data/logs-vrl/).
- `vrl_transformation_spans` (String, Computed) VRL program Better Stack runs on each span as it's ingested. `null` restores the platform's default program, and `""` is an empty program. Read more about [VRL transformations](https://betterstack.com/docs/logs/using-logtail/transforming-ingested-data/logs-vrl/).
- `blocked_metrics` (List of String) Metric names marked as spam, which are dropped when they're ingested. Read more about [marking metrics as spam](https://betterstack.com/docs/logs/using-logtail/transforming-ingested-data/mark-metrics-as-spam/). Defaults to `[]`.
- `scrape_urls` (List of String) URLs Better Stack scrapes. Only for the `prometheus_scrape` and `cloudflare_prometheus_exporter` platforms. Defaults to `[]`.
- `scrape_frequency_secs` (Number) Seconds between scrapes. Only for the `prometheus_scrape` and `cloudflare_prometheus_exporter` platforms. Valid values: `15`, `30`, `60`, `120`, `300`. Defaults to `15`.
- `scrape_request_headers` (Attributes List, Computed) Custom headers sent with each scrape request, each with a `name` and a `value`. Only for the `prometheus_scrape` and `cloudflare_prometheus_exporter` platforms. (see [below for nested schema](#nestedatt--scrape_request_headers))
- `scrape_request_basic_auth_user` (String) Username for basic authentication on scrape requests. Only for the `prometheus_scrape` and `cloudflare_prometheus_exporter` platforms.
- `scrape_request_basic_auth_password` (String, Sensitive) Password for basic authentication on scrape requests. It isn't returned after it's saved. Only for the `prometheus_scrape` and `cloudflare_prometheus_exporter` platforms. Better Stack never returns it, so Terraform keeps the value you set.
- `skip_ssl_verify` (Boolean) Whether to skip TLS certificate verification when scraping. Only for the `prometheus_scrape` and `cloudflare_prometheus_exporter` platforms. Defaults to `false`.
- `gcp_metrics_filters` (Attributes List, Computed) Which Google Cloud metrics the source collects. Each rule has one key, `include` or `exclude`, holding a regular expression that must match the whole metric type, or `*` for every metric. All rules use the same key. An empty list collects no metrics, which is the default. The source's list of GCP metrics shows the metrics and whether these rules collect each. Only for the `gcp` platform. (see [below for nested schema](#nestedatt--gcp_metrics_filters))
- `custom_bucket` (Attributes) Your own S3-compatible bucket to archive data in, or `null` when the source uses Better Stack storage. Read more about [hosting data in your own bucket](https://betterstack.com/docs/logs/host-data-in-your-own-bucket/). The bucket can't be changed after the source is created. An update that sets only `access_key_id` and `secret_access_key` rotates its key pair. We check the new pair against the bucket first and refuse the whole request when it fails. A bucket that several sources share is rotated for all of them. Changing `name`, `endpoint` or `keep_data_after_retention`, or adding or removing it, replaces the resource. (see [below for nested schema](#nestedatt--custom_bucket))
- `connected_source_ids` (List of String, Computed) IDs of the sources connected to this one to correlate their logs and traces. A connection goes both ways. The source joins the listed sources and the sources they're already connected to. An update removes no existing connection, and `[]` disconnects the source.
- `service_fields` (Attributes List, Computed) Log fields that break data down by service. Each entry needs a `name_key`, and `type_key` can be `null` to skip it. An update replaces the existing list, and `[]` removes them all. (see [below for nested schema](#nestedatt--service_fields))
- `deletion_protection` (Boolean) Whether destroying or replacing this source is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.

### Read-Only

- `id` (String) The ID of this source.
- `cluster` (String) The cluster the source's data is stored on, by its name. To connect Grafana to this data, send it as the `cluster` of [Create a connection](https://betterstack.com/docs/logs/api/connections/create/).
- `source_types` (List of String) Row types the source holds, such as `logs`, `spans`, `metrics`, `exceptions` or `web_events`, sorted by name.
- `parent_source_id` (String) Source this source shares its table with when it's connected to one, whose collections `query_collections` names, or `null` otherwise.
- `query_collections` (Attributes) ClickHouse names to query the source's data by. A connected source reads its parent's collections, so `app_filter` for the raw collections and `metrics_app_filter` for the metrics collection hold the condition that keeps only its rows. Both are `null` otherwise. (see [below for nested schema](#nestedatt--query_collections))
- `token` (String, Sensitive) Ingestion token. Returned on every read. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `table_name` (String) Underlying table name.
- `ingesting_host` (String) Host to send data to. `null` for the `prometheus_scrape` and `cloudflare_prometheus_exporter` platforms, whose data Better Stack scrapes.
- `created_at` (String (RFC 3339)) When the source was created.
- `updated_at` (String (RFC 3339)) When the source was last updated.
- `team_name` (String) Name of the team that owns the source.

<a id="values--platform"></a>
### Valid values for `platform`

`apache2`, `aws`, `aws_cloudwatch`, `aws_ecs`, `aws_elb`, `aws_fargate`, `azure`, `azure_logs`, `claude_platform`, `cloudflare_logpush`, `cloudflare_prometheus_exporter`, `cloudflare_worker`, `datadog_agent`, `digitalocean`, `docker`, `dokku`, `dotnet`, `elasticsearch`, `erlang`, `fastly`, `filebeat`, `flights`, `fluentbit`, `fluentd`, `fly_io`, `gcp`, `go`, `google_cloud_pubsub`, `haproxy`, `heroku`, `http`, `java`, `javascript`, `kubernetes`, `logstash`, `minio`, `mongodb`, `mysql`, `nginx`, `open_telemetry`, `oracle_cloud`, `php`, `postgresql`, `prometheus`, `prometheus_scrape`, `python`, `rabbitmq`, `redis`, `render`, `rsyslog`, `ruby`, `syslog-ng`, `traefik`, `ubuntu`, `vector`, `vercel_integration`, `winlogbeat`

<a id="nestedatt--scrape_request_headers"></a>
### Nested Schema for `scrape_request_headers`

Optional:

- `name` (String) The header's name.
- `value` (String) The header's value.

<a id="nestedatt--gcp_metrics_filters"></a>
### Nested Schema for `gcp_metrics_filters`

Optional:

- `include` (String) A regular expression for metric types to collect, or `*` for every metric.
- `exclude` (String) A regular expression for metric types to leave out. Every other metric is collected.

<a id="nestedatt--custom_bucket"></a>
### Nested Schema for `custom_bucket`

Optional:

- `endpoint` (String) The S3-compatible endpoint URL of the bucket. It can't be changed after the bucket is set.
- `access_key_id` (String) The access key ID Better Stack uses to write to the bucket.
- `secret_access_key` (String) The secret access key for `access_key_id`. Write-only, so it's never returned.
- `keep_data_after_retention` (Boolean) Whether data stays in the bucket after the retention period. It can't be changed after the bucket is set.

Read-Only:

- `name` (String) The bucket's name, which Better Stack reads from `endpoint`.

<a id="nestedatt--service_fields"></a>
### Nested Schema for `service_fields`

Optional:

- `name_key` (String) The field to extract the service name from, such as `service.name`.
- `type_key` (String) The field to extract the service's platform from, such as `service.type`. Better Stack matches platforms like Nginx, PostgreSQL or Redis in its value.

<a id="nestedatt--query_collections"></a>
### Nested Schema for `query_collections`

Optional:

- `database` (String) The ClickHouse database that holds the source's data.
- `table` (String) The physical table, as `database.table`.
- `recent` (Map of String) The collection of the last 30 minutes for each row type the source holds, keyed by row type, such as `logs` or `metrics`.
- `s3` (String) The collection that holds older rows of every type.
- `app_filter` (String) The condition that keeps only the source's rows in its parent's raw collections. `null` unless the source is connected to a source.
- `metrics_app_filter` (String) The condition that keeps only the source's rows in its parent's metrics collection. `null` unless the source is connected to a source.

<a id="values--platform"></a>
### Valid values for `platform`

`apache2`, `aws`, `aws_cloudwatch`, `aws_ecs`, `aws_elb`, `aws_fargate`, `azure`, `azure_logs`, `claude_platform`, `cloudflare_logpush`, `cloudflare_prometheus_exporter`, `cloudflare_worker`, `datadog_agent`, `digitalocean`, `docker`, `dokku`, `dotnet`, `elasticsearch`, `erlang`, `fastly`, `filebeat`, `flights`, `fluentbit`, `fluentd`, `fly_io`, `gcp`, `go`, `google_cloud_pubsub`, `haproxy`, `heroku`, `http`, `java`, `javascript`, `kubernetes`, `logstash`, `minio`, `mongodb`, `mysql`, `nginx`, `open_telemetry`, `oracle_cloud`, `php`, `postgresql`, `prometheus`, `prometheus_scrape`, `python`, `rabbitmq`, `redis`, `render`, `rsyslog`, `ruby`, `syslog-ng`, `traefik`, `ubuntu`, `vector`, `vercel_integration`, `winlogbeat`

-> Lists and reads include sources other teams share with the token's teams. Update and delete work only on the teams' own sources.

## Import

Import an existing source by its ID. Get the ID from the [`betterstack_source` data source](../data-sources/source.md), which finds it by `name` or `table_name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_source.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_source.example 12345
```
