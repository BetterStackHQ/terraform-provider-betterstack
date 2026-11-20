---
page_title: "betterstack_source Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A telemetry source: a stream of logs, metrics, or traces ingested into Better Stack.
---

# betterstack_source (Data Source)

A telemetry source: a stream of logs, metrics, or traces ingested into Better Stack.

## Example Usage

```terraform
data "betterstack_source" "example" {
  name = "Nginx"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Source name.
- `table_name` (String) Underlying table name.

### Read-Only

- `platform` (String) The platform the source ingests from. Can't be changed after the source is created. `collector` is refused: create a collector with POST /api/v4/collectors, which creates its own source. Valid values: `apache2`, `aws`, `aws_cloudwatch`, `aws_ecs`, `aws_elb`, `aws_fargate`, `azure`, `azure_logs`, `cloudflare_logpush`, `cloudflare_prometheus_exporter`, `cloudflare_worker`, `datadog_agent`, `digitalocean`, `docker`, `dokku`, `dotnet`, `elasticsearch`, `erlang`, `fastly`, `filebeat`, `flights`, `fluentbit`, `fluentd`, `fly_io`, `gcp`, `go`, `google_cloud_pubsub`, `haproxy`, `heroku`, `http`, `java`, `javascript`, `kubernetes`, `logstash`, `minio`, `mongodb`, `mysql`, `nginx`, `open_telemetry`, `php`, `postgresql`, `prometheus`, `prometheus_scrape`, `python`, `rabbitmq`, `redis`, `render`, `rsyslog`, `ruby`, `syslog-ng`, `traefik`, `ubuntu`, `vector`, `vercel_integration`, `winlogbeat`.
- `data_region` (String) The data region, as the source picker names it: `europe`, `us`, `singapore` or a dedicated cluster's name. It reads back as that name. The older ids `germany`, `us_east` and `us_west` are also accepted. Set on create only. Omit it to use the organization's default.
- `source_group_id` (Number) Source group this source belongs to. Send `null` or `"0"` to remove it from its group.
- `ingesting_paused` (Boolean) Pause ingestion for this source.
- `logs_retention` (Number) Days to keep logs. Refused with 422 on a plan with fixed retention, or for a number of days the plan doesn't offer.
- `metrics_retention` (Number) Days to keep metrics. Refused with 422 on a plan with fixed retention, or for a number of days the plan doesn't offer.
- `live_tail_pattern` (String) The pattern Live tail shows each log line with.
- `vrl_transformation_logs` (String) VRL program Better Stack runs on each log line as it's ingested. Send `null` to reset it to the platform's default, or `""` for an empty program.
- `vrl_transformation_spans` (String) VRL program Better Stack runs on each span as it's ingested. Send `null` to reset it to the platform's default, or `""` for an empty program.
- `blocked_metrics` (List of String) Metric names to drop when they're ingested.
- `scrape_urls` (List of String) URLs to scrape. Only for platforms that scrape, such as `prometheus_scrape`.
- `scrape_frequency_secs` (Number) Seconds between scrapes.
- `scrape_request_headers` (Attributes List, see [nested schema](#nestedatt--scrape_request_headers)) Custom headers sent with each scrape request, each with a `name` and a `value`.
- `scrape_request_basic_auth_user` (String) Username for basic authentication on scrape requests.
- `scrape_request_basic_auth_password` (String, Sensitive) Password for basic authentication on scrape requests. Returned on reads, except on a source another team shares with yours.
- `skip_ssl_verify` (Boolean) Skip TLS verification when scraping.
- `aws_account_id` (String) ID of an AWS account already connected to Better Stack, to reuse it. Only for the `aws` platform.
- `aws_role_arn` (String, Sensitive) ARN of the IAM role from Better Stack's CloudFormation stack. Only for the `aws` platform.
- `aws_external_id` (String, Sensitive) External ID from Better Stack's CloudFormation stack. Only for the `aws` platform.
- `aws_auto_sub_log_groups` (Boolean) Whether to subscribe new CloudWatch log groups automatically when they're discovered. Only for the `aws` platform, and returned only on AWS sources.
- `gcp_account_id` (String) ID of a GCP account already connected to Better Stack, to reuse it. Only for the `gcp` platform.
- `gcp_project_id` (String) ID of the GCP project to connect. Only for the `gcp` platform.
- `gcp_project_number` (String) Number of the GCP project to connect. Only for the `gcp` platform.
- `gcp_metrics_filters` (Attributes List, see [nested schema](#nestedatt--gcp_metrics_filters)) Which Google Cloud metrics the source collects. Each rule has one key, `include` or `exclude`, holding a regular expression that must match the whole metric type, or `*` for every metric. All rules use the same key. An empty list collects no metrics, which is the default. `GET /api/v4/sources/:source_id/gcp-metrics` lists the metrics and whether these rules collect each. Only for the `gcp` platform.
- `gcp_log_filter` (String) Google Cloud Logging query that selects which logs the source receives. `null` forwards every log, which is the default. Google Cloud checks the query when it's saved and refuses an invalid one. Set it once the source's GCP project is connected, in a later request than the one that connects it. Only for the `gcp` platform.
- `custom_bucket` (Attributes, see [nested schema](#nestedatt--custom_bucket)) Your own S3-compatible bucket to archive data in. Set it on create. An update can send only `access_key_id` and `secret_access_key`, both together, to rotate the key pair. `secret_access_key` is never returned.
- `connected_source_ids` (List of Number) IDs of the sources connected to this one, for example to link spans and logs. An update replaces all connected sources. A connection goes both ways.
- `service_fields` (Attributes List, see [nested schema](#nestedatt--service_fields)) Log fields that break data down by service, each with a `name_key` and a `type_key`. An update replaces all of them. `type_key` can be `null`.
- `token` (String, Sensitive) Ingestion token. Returned on every read, so you can configure log shippers with it.
- `ingesting_host` (String) Host to send data to.
- `team_id` (String) Team that owns the source, as its `t`-prefixed Better Stack team ID, such as `t123`.
- `team_name` (String) Name of the team that owns the source.
- `created_at` (String (RFC 3339)) When the source was created.
- `updated_at` (String (RFC 3339)) When the source was last updated.
- `deletion_protection` (Boolean) While true (the default), destroying or replacing this source fails, in a plan or in a pulumi destroy. Kept in Terraform or Pulumi state only: the API never sees it. Set it to false and apply that before deleting or replacing it on purpose.

<a id="nestedatt--scrape_request_headers"></a>
### Nested Schema for `scrape_request_headers`

- `name` (String)
- `value` (String)

<a id="nestedatt--gcp_metrics_filters"></a>
### Nested Schema for `gcp_metrics_filters`

- `include` (String)
- `exclude` (String)

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

- `GET /api/v4/sources`
- `GET /api/v4/sources/:id`
- `POST /api/v4/sources`
- `PATCH /api/v4/sources/:id`
- `DELETE /api/v4/sources/:id`
