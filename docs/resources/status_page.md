---
page_title: "betterstack_status_page Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A public or private status page that communicates service health to your users.
---

# betterstack_status_page (Resource)

A public or private status page that communicates service health to your users.

-> Replaces `betteruptime_status_page`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_status_page" "example" {
  company_name = "example"
  subdomain    = "example-page"
  timezone     = "UTC"
}
```

## Schema

### Required

- `company_name` (String) Company name shown on the page.
- `subdomain` (String) betterstack.com subdomain for the page.
- `timezone` (String) Timezone for displayed times.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `organization_id` (Number) Organization the page belongs to, as its Better Stack organization id. Defaults to the first organization the token can access. An organization the token can't access returns a 422 error that lists the ones it can. Ignored on update. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `company_url` (String) Company website URL.
- `custom_domain` (String) Custom domain for the page. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `contact_url` (String) Contact URL shown on the page. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `logo_url` (String) The logo shown in light mode. Send an image URL and the logo is fetched from it. Responses return the URL of the stored logo. Send `null` to remove the logo. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `dark_logo_url` (String) The logo shown in dark mode. Send an image URL and the logo is fetched from it. Responses return the URL of the stored logo. Send `null` to remove the logo. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `logo_remote_url` (String) The URL the light mode logo was fetched from. Send `null` to remove the logo. When both are sent, `logo_url` wins. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `dark_logo_remote_url` (String) The URL the dark mode logo was fetched from. Send `null` to remove the logo. When both are sent, `dark_logo_url` wins. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `custom_css` (String) Custom CSS added to the page. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `custom_javascript` (String) Custom JavaScript added to the page. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `google_analytics_id` (String) Google Analytics measurement ID. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `min_incident_length` (Number) Shortest incident the page shows, in seconds. Must be less than a year. Defaults to `0`.
- `subscribable` (Boolean) Whether visitors can subscribe to updates. Defaults to `false`.
- `hide_from_search_engines` (Boolean) Whether to hide the page from search engines with a `noindex` tag. Defaults to `false`.
- `password_enabled` (Boolean) Whether visitors need a password to view the page. Defaults to `false`.
- `password` (String, Sensitive) Password visitors enter to view the page. Never returned. Setting a password also turns on `password_enabled`, unless you send `password_enabled` too. An empty password removes the password and turns protection off. Write-only: the API never returns it; Terraform keeps the configured value.
- `require_sso` (Boolean) Whether visitors must sign in with SSO to view the page. Defaults to `false`.
- `published` (Boolean) Whether the page is published and publicly visible. An unpublished page is a draft. Publishing a draft fails when the organization already has as many published status pages as its plan allows. Defaults to `true`.
- `history` (Number) Number of days of history the page shows, from 7 to 365. Defaults to `90`.
- `design` (String) Page design: `v1` for the classic look or `v2` for the modern one. Valid values: `v1`, `v2`. Defaults to `v1`.
- `theme` (String) Color theme. Valid values: `light`, `dark`, `system`. Defaults to `light`.
- `layout` (String) How resources are laid out on the page. Valid values: `vertical`, `horizontal`. Defaults to `vertical`.
- `automatic_reports` (Boolean) Whether maintenance reports are created automatically. Defaults to `false`.
- `status_page_group_id` (Number) Status page group the page belongs to. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `whitelabeled` (Boolean) Whether to remove Better Stack branding from the page. Defaults to `false`.
- `announcement` (String) Announcement banner text. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `announcement_embed_visible` (Boolean) Whether the embeddable widget shows the announcement. Defaults to `true`.
- `announcement_embed_link` (String) Link the announcement points to on the embeddable widget. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `announcement_embed_css` (String) Custom CSS for the announcement on the embeddable widget. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `navigation_links` (Attributes List, see [nested schema](#nestedatt--navigation_links)) Up to four links in the page's navigation, each with `text` and `href`. Only the `v2` design shows them. Send `[]` to remove them all.
- `ip_allowlist` (List of String) IP addresses or CIDR ranges allowed to view the page. Send `[]` to remove them all.
- `sections` (Attributes List, see [nested schema](#nestedatt--sections)) The page's sections in display order, each with its resources in display order. Sent, the list replaces the page's whole structure: sections and resources are matched by id, ones without an id are created, and ones left out are removed from the page. The monitors and other records behind them aren't affected. An attribute a resource leaves out keeps its value. Left out, the structure stays as it is. Writable on update only, so a new page gets its sections in a follow-up update. Not accepted by the create request; the provider applies it in an immediate follow-up call once the resource exists.

### Read-Only

- `id` (String) The id of this status_page.
- `aggregate_state` (String) The page's overall status: `operational`, `degraded`, `downtime` or `maintenance`.
- `created_at` (String (RFC 3339)) When the status page was created.
- `updated_at` (String (RFC 3339)) When the status page was last updated.

<a id="nestedatt--navigation_links"></a>
### Nested Schema for `navigation_links`

- `text` (String)
- `href` (String)

<a id="nestedatt--sections"></a>
### Nested Schema for `sections`

- `id` (String) Section id. Leave it out for a new section. An id keeps the section, even when it moves.
- `name` (String) Section name.
- `resources` (Attributes List, see [nested schema](#nestedatt--sections--resources)) The section's resources, in display order. Left out, the section keeps its resources as they are, and a new section starts empty.

<a id="nestedatt--sections--resources"></a>
### Nested Schema for `sections.resources`

- `id` (String) Resource id. Leave it out for a new resource. An id keeps the resource and its subscribers, even when it moves to another section.
- `resource_type` (String) Type of the underlying resource. Required for a new resource. ThirdPartyService is added in the web app only: an item read back keeps its type and services when listed by id. Valid values: `Monitor`, `MonitorGroup`, `Heartbeat`, `HeartbeatGroup`, `WebhookIntegration`, `EmailIntegration`, `IncomingWebhook`, `AmazonSnsIntegration`, `ManuallyTrackedItem`, `LogsChart`, `CatalogReference`, `ThirdPartyService`.
- `resource_id` (Number) Id of the underlying monitor, heartbeat, group or other record. Required for a new resource, unless resource_type is ManuallyTrackedItem, which creates its backing record.
- `public_name` (String) Name shown publicly. Required for a new resource.
- `explanation` (String) Tooltip explanation shown to visitors.
- `widget_type` (String) How this resource is rendered. Valid values: `plain`, `history`, `intraday_history`, `response_times`, `intraday_response_times`, `chart_only`.
- `mark_as_down_for` (String) When an incident marks this resource as down. Valid values: `no_incident`, `any_incident`, `incident_matching_metadata`.
- `mark_as_down_metadata_rule` (Attributes, see [nested schema](#nestedatt--sections--resources--mark_as_down_metadata_rule)) Metadata rule used when mark_as_down_for is incident_matching_metadata, as on a status page resource.
- `mark_as_degraded_for` (String) When an incident marks this resource as degraded. Valid values: `no_incident`, `any_incident`, `incident_matching_metadata`.
- `mark_as_degraded_metadata_rule` (Attributes, see [nested schema](#nestedatt--sections--resources--mark_as_degraded_metadata_rule)) Metadata rule used when mark_as_degraded_for is incident_matching_metadata, in the shape of mark_as_down_metadata_rule.

<a id="nestedatt--sections--resources--mark_as_down_metadata_rule"></a>
### Nested Schema for `sections.resources.mark_as_down_metadata_rule`

- `key` (String)
- `values` (Attributes List, see [nested schema](#nestedatt--sections--resources--mark_as_down_metadata_rule--values))

<a id="nestedatt--sections--resources--mark_as_down_metadata_rule--values"></a>
### Nested Schema for `sections.resources.mark_as_down_metadata_rule.values`

- `type` (String)
- `value` (String)
- `item_id` (String)
- `name` (String)
- `email` (String)

<a id="nestedatt--sections--resources--mark_as_degraded_metadata_rule"></a>
### Nested Schema for `sections.resources.mark_as_degraded_metadata_rule`

- `key` (String)
- `values` (Attributes List, see [nested schema](#nestedatt--sections--resources--mark_as_degraded_metadata_rule--values))

<a id="nestedatt--sections--resources--mark_as_degraded_metadata_rule--values"></a>
### Nested Schema for `sections.resources.mark_as_degraded_metadata_rule.values`

- `type` (String)
- `value` (String)
- `item_id` (String)
- `name` (String)
- `email` (String)

## Endpoints

- `GET /api/v5/status-pages`
- `GET /api/v5/status-pages/:id`
- `POST /api/v5/status-pages`
- `PATCH /api/v5/status-pages/:id`
- `DELETE /api/v5/status-pages/:id`

-> Sending `null` for a setting that can't be empty keeps its current value. That covers every true or false setting, `min_incident_length`, `history`, `design`, `theme`, `layout` and `navigation_links`.

-> `logo_url`, `logo_remote_url`, `dark_logo_url` and `dark_logo_remote_url` must be absolute URLs, or the request returns a 422 error. Leaving out a logo's fields keeps the logo.

## Import

```shell
terraform import betterstack_status_page.example 12345
```
