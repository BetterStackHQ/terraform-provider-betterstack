---
page_title: "betterstack_status_page Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A public or private status page that communicates service health to your users.
---

# betterstack_status_page (Resource)

A public or private status page that communicates service health to your users.

~> Destroying a status page deletes its data for good, so `deletion_protection` is on by default: Terraform refuses to destroy or replace it until you set `deletion_protection = false` and apply.

-> Replaces `betteruptime_status_page`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a status page

```terraform
resource "betterstack_status_page" "example" {
  company_url  = "https://best-company.com"
  company_name = "Best company"
  subdomain    = "best-company-tokyo"
  timezone     = "Tokyo"
}
```

### In a status page group

```terraform
resource "betterstack_status_page_group" "example" {
  name       = "Production Status Pages"
  sort_index = 1
}

resource "betterstack_status_page" "in_status_page_group" {
  company_url          = "https://best-company.com"
  company_name         = "Best company"
  subdomain            = "best-company-tokyo"
  timezone             = "Tokyo"
  status_page_group_id = betterstack_status_page_group.example.id
}
```

## Schema

### Required

- `company_name` (String) Company name shown on the page.
- `subdomain` (String) Subdomain of the page's default address. It must be unique across Better Stack, up to 63 characters of letters, digits, dashes and underscores.
- `timezone` (String) Time zone the page shows times in, as a Rails time zone name such as `Tokyo` or an IANA ID such as `Europe/Prague`. Returned as the Rails name.

### Optional

- `deletion_protection` (Boolean) Whether destroying or replacing this status page is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.
- `organization_id` (String, Forces replacement) The ID of the organization the page belongs to, such as `123`. Get the ID from `organization_id` in [List teams](https://betterstack.com/docs/uptime/api/teams/list/). Required on create when you sign in with OAuth and can access more than one organization.
- `company_url` (String) Company website URL.
- `custom_domain` (String) A host name you own to serve the page from, such as `status.example.com`. Point its CNAME record to `statuspage.betteruptime.com`.
- `contact_url` (String) URL visitors can use to get in touch, such as a support page or a `mailto:` address like `mailto:support@example.com`.
- `logo_url` (String) The logo shown in light mode. Responses return the URL of the stored logo. It accepts the URL of an image under 20 MB, and we fetch the logo from it. Remove the argument to remove the logo.
- `dark_logo_url` (String) The logo shown in dark mode. Responses return the URL of the stored logo. It accepts the URL of an image under 20 MB, and we fetch the logo from it. Remove the argument to remove the logo.
- `logo_remote_url` (String) The URL the light mode logo was fetched from. Remove the argument to remove the logo. When both are sent, `logo_url` wins.
- `dark_logo_remote_url` (String) The URL the dark mode logo was fetched from. Remove the argument to remove the logo. When both are sent, `dark_logo_url` wins.
- `favicon_remote_url` (String) The URL the favicon was fetched from. It accepts the URL of an image under 20 MB, and we fetch the favicon from it. Remove the argument to remove the favicon.
- `custom_css` (String) Custom CSS added to the page.
- `custom_javascript` (String) Custom JavaScript added to the page.
- `google_analytics_id` (String) Your Google Analytics ID, such as `G-XXXXXXXX`. It must start with `G-`, `UA-` or `GTM-`.
- `min_incident_length` (Number) Shortest incident the page shows, in seconds. `0` shows every incident. Must be less than a year. Defaults to `0`.
- `subscribable` (Boolean) Whether visitors can subscribe to updates. Defaults to `false`.
- `hide_from_search_engines` (Boolean) Whether to hide the page from search engines with a `noindex` tag. Defaults to `false`.
- `password_enabled` (Boolean, Computed) Whether visitors need a password to view the page. Off unless you turn it on or set `password`.
- `password` (String, Sensitive) Password visitors enter to view the page. Never returned. Setting a password also turns on `password_enabled`, unless `password_enabled` is set too. An empty password removes the password and turns protection off. Better Stack never returns it, so Terraform keeps the value you set.
- `require_sso` (Boolean) Whether visitors must sign in with SSO to view the page. Needs SSO set up for your organization, and can't be on together with `password_enabled`. Defaults to `false`.
- `published` (Boolean) Whether the page is published and publicly visible. An unpublished page is a draft. Publishing a draft fails when the organization already has as many published status pages as its plan allows. Defaults to `true`.
- `history` (Number) Number of days of history the page shows, from 7 to 365. Defaults to `90`.
- `treat_not_monitored_as_operational` (Boolean) Whether the days a resource wasn't monitored show as operational in its history and count toward its uptime, instead of showing as not monitored. Defaults to `false`.
- `design` (String) Page design. Defaults to `v1`.
  - `v1`: The classic design.
  - `v2`: The modern design.
- `theme` (String) Color theme. Only the `v2` design uses it. Valid values: `light`, `dark`, `system`. Defaults to `light`.
- `layout` (String) Header layout of the page. Only the `v2` design uses it. Valid values: `vertical`, `horizontal`. Defaults to `vertical`.
- `automatic_reports` (Boolean) Whether status page reports are created automatically for new incidents. Defaults to `false`.
- `status_page_group_id` (String) The ID of the status page group the page belongs to. Get the ID via [List status page groups](https://betterstack.com/docs/uptime/api/list-all-status-page-groups/).
- `whitelabeled` (Boolean) Whether to remove Better Stack branding from the page. Defaults to `false`.
- `announcement` (String) Announcement banner text.
- `announcement_embed_visible` (Boolean) Whether the embeddable widget shows the announcement. Defaults to `true`.
- `announcement_embed_link` (String) Link the announcement points to on the embeddable widget. Defaults to the status page.
- `announcement_embed_css` (String) Your own CSS for the announcement on the embeddable widget.
- `navigation_links` (Attributes List, Computed) Up to four links in the page's navigation, each with `text` and `href`. Only the `v2` design shows them. An empty list removes them all. (see [below for nested schema](#nestedatt--navigation_links))
- `ip_allowlist` (List of String, Computed) IP addresses or CIDR ranges allowed to view the page. Entries starting with `#` are comments. Set it to `[]` to let anyone view it.
- `sections` (Attributes List, Computed) The page's sections and their resources, in display order. Setting it replaces the whole structure. Sections and resources are matched by `id`. Ones without an `id` are created. Ones left out are removed from the page. The monitors and other records behind them aren't affected. An attribute a resource leaves out keeps its value. Leave it out to keep the structure as it is. Only an update accepts it. In Terraform, an item without an id keeps the id of the existing item with the same `name`, so a reorder or a move updates it in place. An item that matches none keeps the id of the existing item at its position when no other item took that one, so a rename updates it in place too. (see [below for nested schema](#nestedatt--sections))

### Read-Only

- `id` (String) The ID of this status page.
- `favicon_url` (String) The icon browsers show in the page's tab, as the URL of the stored favicon. Only the `v2` design shows it.
- `aggregate_state` (String) The page's overall status.
  - `operational`: No resource on the page is down, degraded or in maintenance.
  - `degraded`: A resource is degraded and none is down.
  - `downtime`: A resource is down.
  - `maintenance`: A resource is in maintenance and none is down or degraded.
- `public_url` (String) The address visitors open the page at. It is the `custom_domain` when the page has one, otherwise the betterstack.com `subdomain`.
- `created_at` (String (RFC 3339)) When the status page was created.
- `updated_at` (String (RFC 3339)) When the status page was last updated.

<a id="nestedatt--navigation_links"></a>
### Nested Schema for `navigation_links`

Optional:

- `text` (String) The link's text, such as `Maintenance`.
- `href` (String) Where the link goes, a full URL, or `/`, `/maintenance` or `/incidents` for the status page's own pages.

<a id="nestedatt--sections"></a>
### Nested Schema for `sections`

Optional:

- `id` (String) Section id. Leave it out for a new section. An id keeps the section, even when it moves.
- `name` (String) Section name.
- `resources` (Attributes List) The section's resources, in display order. Leave it out to keep the section's resources as they are. A new section starts empty. In Terraform, an item without an id keeps the id of the existing item with the same `resource_type` and `resource_id`, or else the same `resource_type` and `public_name`, so a reorder or a move updates it in place. (see [below for nested schema](#nestedatt--sections--resources))

<a id="nestedatt--sections--resources"></a>
### Nested Schema for `sections.resources`

Optional:

- `id` (String) Resource id. Leave it out for a new resource. An id keeps the resource and its subscribers, even when it moves to another section.
- `resource_type` (String) Type of the underlying resource. Required for a new resource.
  - `Monitor`: Shows a monitor.
  - `MonitorGroup`: Shows a monitor group.
  - `Heartbeat`: Shows a heartbeat.
  - `HeartbeatGroup`: Shows a heartbeat group.
  - `WebhookIntegration`: Shows an integration that opens incidents from another tool's alerts, such as Datadog or Grafana.
  - `EmailIntegration`: Shows an e-mail integration.
  - `IncomingWebhook`: Shows an incoming webhook.
  - `AmazonSnsIntegration`: Shows an Amazon SNS integration.
  - `ManuallyTrackedItem`: Shows a manually tracked item. Better Stack creates it when you leave out `resource_id`.
  - `LogsChart`: Shows a dashboard chart, given in `chart_id`.
  - `CatalogReference`: Shows a catalog value, given in `catalog_reference`.
  - `ThirdPartyService`: Shows the status of a third-party service. It can only be added in the status page editor, and sent back with its `id` it keeps the services it shows.
- `resource_id` (String) ID of the underlying monitor, heartbeat, group or other record. Required for a new resource. A `ManuallyTrackedItem` doesn't need it, since it creates its backing record. Neither does a resource whose `catalog_reference` names a catalog value or whose `chart_id` names a dashboard chart.
- `catalog_reference` (Attributes) For a `CatalogReference` resource, the catalog value it shows, as on a status page resource. When it's set, the value is found or created and decides the resource's `resource_id`, so a `resource_id` set with it is ignored. `null` for every other type. (see [below for nested schema](#nestedatt--sections--resources--catalog_reference))
- `chart_id` (String) For a `LogsChart` resource, the ID of the dashboard chart it shows, as on a status page resource. The chart must be on one of your organization's dashboards. It decides the resource's `resource_id`, so a `resource_id` sent with it is ignored. `null` for every other type.
- `public_name` (String) Name shown publicly. Required for a new resource.
- `explanation` (String) Tooltip explanation shown to visitors.
- `widget_type` (String) How this resource is rendered.
  - `plain`: Shows only the current status.
  - `history`: Shows the current status and a history of it.
  - `intraday_history`: Shows the current status and a detailed history of it.
  - `response_times`: Shows a history of the status and a response times chart, for a monitor or a logs chart.
  - `intraday_response_times`: Shows a detailed history of the status and a response times chart, for a monitor or a logs chart.
  - `chart_only`: Shows only the chart, for a logs chart.
- `mark_as_down_for` (String) When an incident marks this resource as down. A `ManuallyTrackedItem` has no incidents, so this has no effect on one.
  - `no_incident`: No incident marks the resource down.
  - `any_incident`: Every open incident of the resource marks it down.
  - `incident_matching_metadata`: An open incident marks the resource down when its metadata matches `mark_as_down_metadata_rule`.
- `mark_as_down_metadata_rule` (Attributes) Metadata rule used when `mark_as_down_for` is `incident_matching_metadata`, as on a status page resource. A user is `u123` or their e-mail, and a team is `t123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/), and a team's ID via [List teams](https://betterstack.com/docs/uptime/api/teams/list/). A `ManuallyTrackedItem` has no incidents, so this has no effect on one. (see [below for nested schema](#nestedatt--sections--resources--mark_as_down_metadata_rule))
- `mark_as_degraded_for` (String) When an incident marks this resource as degraded. A `ManuallyTrackedItem` has no incidents, so this has no effect on one.
  - `no_incident`: No incident marks the resource degraded.
  - `any_incident`: Every open incident of the resource marks it degraded.
  - `incident_matching_metadata`: An open incident marks the resource degraded when its metadata matches `mark_as_degraded_metadata_rule`.
- `mark_as_degraded_metadata_rule` (Attributes) Metadata rule used when `mark_as_degraded_for` is `incident_matching_metadata`, in the shape of `mark_as_down_metadata_rule`. A `ManuallyTrackedItem` has no incidents, so this has no effect on one. (see [below for nested schema](#nestedatt--sections--resources--mark_as_degraded_metadata_rule))

<a id="nestedatt--sections--resources--catalog_reference"></a>
### Nested Schema for `sections.resources.catalog_reference`

Optional:

- `key` (String) The catalog relation's name.
- `value` (Attributes) The catalog value to show, a plain string or a catalog item. (see [below for nested schema](#nestedatt--sections--resources--catalog_reference--value))

<a id="nestedatt--sections--resources--catalog_reference--value"></a>
### Nested Schema for `sections.resources.catalog_reference.value`

Optional:

- `type` (String) `String` for a plain value, or the type of the catalog item, such as `User` or `Team`.
- `value` (String) The plain value, when `type` is `String`.
- `item_id` (String) The catalog item's ID, when `type` isn't `String`. A user is `u123` or their e-mail, and a team is `t123`.
- `name` (String) The catalog item's name. Set it instead of `item_id` to find the item.
- `email` (String) The user's e-mail, when `type` is `User`. Set it instead of `item_id` to find the user.

<a id="nestedatt--sections--resources--mark_as_down_metadata_rule"></a>
### Nested Schema for `sections.resources.mark_as_down_metadata_rule`

Optional:

- `key` (String) The incident metadata key to match.
- `values` (Attributes List) The values to match. An incident matches when its metadata under `key` holds one of them. (see [below for nested schema](#nestedatt--sections--resources--mark_as_down_metadata_rule--values))

<a id="nestedatt--sections--resources--mark_as_down_metadata_rule--values"></a>
### Nested Schema for `sections.resources.mark_as_down_metadata_rule.values`

Optional:

- `type` (String) `String` for a plain value, or the type of the catalog item, such as `User` or `Team`.
- `value` (String) The plain value, when `type` is `String`.
- `item_id` (String) The catalog item's ID, when `type` isn't `String`. A user is `u123` or their e-mail, and a team is `t123`. Get the user ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/).
- `name` (String) The catalog item's name. Set it instead of `item_id` to find the item.
- `email` (String) The user's e-mail, when `type` is `User`. Set it instead of `item_id` to find the user.

<a id="nestedatt--sections--resources--mark_as_degraded_metadata_rule"></a>
### Nested Schema for `sections.resources.mark_as_degraded_metadata_rule`

Optional:

- `key` (String) The incident metadata key to match.
- `values` (Attributes List) The values to match. An incident matches when its metadata under `key` holds one of them. (see [below for nested schema](#nestedatt--sections--resources--mark_as_degraded_metadata_rule--values))

<a id="nestedatt--sections--resources--mark_as_degraded_metadata_rule--values"></a>
### Nested Schema for `sections.resources.mark_as_degraded_metadata_rule.values`

Optional:

- `type` (String) `String` for a plain value, or the type of the catalog item, such as `User` or `Team`.
- `value` (String) The plain value, when `type` is `String`.
- `item_id` (String) The catalog item's ID, when `type` isn't `String`. A user is `u123` or their e-mail, and a team is `t123`. Get the user ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/).
- `name` (String) The catalog item's name. Set it instead of `item_id` to find the item.
- `email` (String) The user's e-mail, when `type` is `User`. Set it instead of `item_id` to find the user.

-> The `logo_url`, `logo_remote_url`, `dark_logo_url`, `dark_logo_remote_url` and `favicon_remote_url` attributes must be absolute URLs, or the change is refused. Leaving out an image's fields keeps the image.

## Import

Import an existing status page by its ID.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_status_page.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_status_page.example 12345
```
