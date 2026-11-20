---
page_title: "betterstack_dashboard_template Data Source - terraform-provider-betterstack"
subcategory: "Telemetry"
description: |-
  A prebuilt dashboard template that can be exported and imported to create a dashboard.
---

# betterstack_dashboard_template (Data Source)

A prebuilt dashboard template that can be exported and imported to create a dashboard.

-> The API describes no create/delete for this object, so the data source is its only Terraform surface.

## Example Usage

```terraform
data "betterstack_dashboard_template" "example" {
  name = "Better Stack Collector"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Template name.

### Read-Only

- `description` (String) What the template visualizes.
- `categories` (List of String) Categories the template is filed under.

## Endpoints

- `GET /api/v4/dashboards/templates`
