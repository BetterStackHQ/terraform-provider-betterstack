---
page_title: "betterstack_ip_list Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  List the IP addresses Better Stack monitors check from, grouped by cluster. Use this public list to keep firewall allowlists up to date.
---

# betterstack_ip_list (Data Source)

List the IP addresses Better Stack monitors check from, grouped by cluster. Use this public list to keep firewall allowlists up to date.

-> The API describes no create/delete for this object, so the data source is its only Terraform surface.

## Example Usage

```terraform
data "betterstack_ip_list" "example" {
}
```

## Schema

### Optional

- `filter_clusters` (List of String) Restrict values to these collection keys. Omitted or empty selects all keys. Unknown keys are rejected.

`id` is computed. Every refresh performs a live query; no API token is needed when the declared endpoint is public.

### Read-Only

- `all_clusters` (List of String)
- `ips` (List of String)
