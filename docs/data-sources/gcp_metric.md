---
page_title: "betterstack_gcp_metric Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A Google Cloud metric a GCP source can collect, and whether its metric filters collect it.
---

# betterstack_gcp_metric (Data Source)

A Google Cloud metric a GCP source can collect, and whether its metric filters collect it.

-> The API describes no create/delete for this object, so the data source is its only Terraform surface.

## Example Usage

```terraform
data "betterstack_gcp_metric" "example" {
  source_id   = "12345"
  metric_type = "compute.googleapis.com/instance/cpu/utilization"
}
```

## Schema

### Address

- `source_id` (String, Required) The ID of the parent this source belongs to. Changing it replaces the resource.

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `metric_type` (String) Google Cloud metric type, such as `compute.googleapis.com/instance/cpu/utilization`.

### Read-Only

- `display_name` (String) Name Google Cloud shows for the metric.
- `description` (String) What the metric measures, as Google Cloud describes it.
- `kind` (String) Google Cloud metric kind, `GAUGE`, `DELTA` or `CUMULATIVE`.
- `service_name` (String) Google Cloud service the metric belongs to, such as `compute.googleapis.com`.
- `collected` (Boolean) Whether the source's `gcp_metrics_filters` collect this metric. Change them via [Update a source](https://betterstack.com/docs/logs/api/update-source/).
