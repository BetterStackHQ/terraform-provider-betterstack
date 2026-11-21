---
page_title: "betterstack_status_page_resource Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A monitor, heartbeat, or group displayed as a line item on a status page.
---

# betterstack_status_page_resource (Resource)

A monitor, heartbeat, or group displayed as a line item on a status page.

## Example Usage

```terraform
resource "betterstack_status_page_resource" "example" {
  status_page_id = "12345"
  resource_type  = "Monitor"
  public_name    = "https://uptime.betterstack.com/"
}
```

## Schema

### Address

This object lives under a parent collection, so its parent is part of its address.

- `status_page_id` (String, Required) Id of the parent this status page belongs to. Changing it replaces the resource.

### Required

- `resource_type` (String) Type of the resource to show. A group reads back as `MonitorGroup` or `HeartbeatGroup`, depending on what it groups. Valid values: `Monitor`, `MonitorGroup`, `Heartbeat`, `HeartbeatGroup`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`, `AmazonSnsIntegration`, `ManuallyTrackedItem`, `LogsChart`, `CatalogReference`, `ThirdPartyService`.
- `public_name` (String) Name shown publicly for this item.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `status_page_section_id` (String) Section to place this item in. Omit to use the page's first section, or a newly created default section on a page with none. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `resource_id` (String) ID of the monitor, heartbeat or group to show. Required unless `resource_type` is `ManuallyTrackedItem`: leave it out there to create a new manually tracked item. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `catalog_reference` (Attributes, see [nested schema](#nestedatt--catalog_reference)) For a `CatalogReference` item, the catalog value it shows. `key` is the catalog relation's name, and `value` is a string or a catalog item given as `{type, item_id | name | email}`. The value is found or created when you send it, in place of `resource_id`. A Team `item_id` reads back as a `t`-prefixed ID. `null` for every other type. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `explanation` (String) Tooltip explanation shown to visitors. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `history` (Boolean) Whether `widget_type` is `history`. Sending `true` sets `widget_type` to `history`, and `false` sets it to `plain`, unless `widget_type` is also sent. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `widget_type` (String) How this item is rendered. Valid values: `plain`, `history`, `intraday_history`, `response_times`, `intraday_response_times`, `chart_only`. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `position` (Number) 0-based display order within the section. Inserting at a position shifts the items at or after it down one. With fixed_position true, which Terraform always sends, the position is written as given and nothing shifts. A position past the end lands last. Left out, the new one is added at the end. Must be a non-negative integer. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `mark_as_down_for` (String) When an incident marks this item as down. Valid values: `no_incident`, `any_incident`, `incident_matching_metadata`. Defaults to `any_incident`.
- `mark_as_down_metadata_rule` (Attributes, see [nested schema](#nestedatt--mark_as_down_metadata_rule)) The incident metadata rule used when `mark_as_down_for` is `incident_matching_metadata`. It has a `key` and `values`, each either `{type: String, value}` or a catalog item `{type, item_id, name, email}`. A Team `item_id` reads back as a `t`-prefixed ID, and other item IDs are numbers. Returned only when `mark_as_down_for` is `incident_matching_metadata`. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `mark_as_degraded_for` (String) When an incident marks this item as degraded. Valid values: `no_incident`, `any_incident`, `incident_matching_metadata`. Defaults to `no_incident`.
- `mark_as_degraded_metadata_rule` (Attributes, see [nested schema](#nestedatt--mark_as_degraded_metadata_rule)) The incident metadata rule used when `mark_as_degraded_for` is `incident_matching_metadata`, in the same shape as `mark_as_down_metadata_rule`. Returned only when `mark_as_degraded_for` is `incident_matching_metadata`. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.

### Read-Only

- `id` (String) The id of this status_page_resource.
- `availability` (Number) Availability over the page's history window, from 0.0 to 1.0.
- `status` (String) Current reported status.
- `status_history` (Attributes List, see [nested schema](#nestedatt--status_history)) Daily status over the page's history window. Each day has `day`, `status`, and `downtime_duration` and `maintenance_duration` in seconds.

<a id="nestedatt--catalog_reference"></a>
### Nested Schema for `catalog_reference`

- `key` (String)
- `value` (Attributes, see [nested schema](#nestedatt--catalog_reference--value))

<a id="nestedatt--catalog_reference--value"></a>
### Nested Schema for `catalog_reference.value`

- `type` (String)
- `value` (String)
- `item_id` (String)
- `name` (String)
- `email` (String)

<a id="nestedatt--mark_as_down_metadata_rule"></a>
### Nested Schema for `mark_as_down_metadata_rule`

- `key` (String)
- `values` (Attributes List, see [nested schema](#nestedatt--mark_as_down_metadata_rule--values))

<a id="nestedatt--mark_as_down_metadata_rule--values"></a>
### Nested Schema for `mark_as_down_metadata_rule.values`

- `type` (String)
- `value` (String)
- `item_id` (String)
- `name` (String)
- `email` (String)

<a id="nestedatt--mark_as_degraded_metadata_rule"></a>
### Nested Schema for `mark_as_degraded_metadata_rule`

- `key` (String)
- `values` (Attributes List, see [nested schema](#nestedatt--mark_as_degraded_metadata_rule--values))

<a id="nestedatt--mark_as_degraded_metadata_rule--values"></a>
### Nested Schema for `mark_as_degraded_metadata_rule.values`

- `type` (String)
- `value` (String)
- `item_id` (String)
- `name` (String)
- `email` (String)

<a id="nestedatt--status_history"></a>
### Nested Schema for `status_history`

- `day` (String)
- `status` (String)
- `downtime_duration` (Number)
- `maintenance_duration` (Number)

## Endpoints

- `GET /api/v3/status-pages/:status_page_id/resources`
- `GET /api/v3/status-pages/:status_page_id/resources/:id`
- `POST /api/v3/status-pages/:status_page_id/resources`
- `PATCH /api/v3/status-pages/:status_page_id/resources/:id`
- `DELETE /api/v3/status-pages/:status_page_id/resources/:id`

## Import

```shell
terraform import betterstack_status_page_resource.example <status_page_id>/12345
```
