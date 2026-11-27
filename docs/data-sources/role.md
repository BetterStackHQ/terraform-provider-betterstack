---
page_title: "betterstack_role Data Source - terraform-provider-betterstack"
subcategory: "Account"
description: |-
  A role in the organization, either a system role or a custom role built in the role editor.
---

# betterstack_role (Data Source)

A role in the organization, either a system role or a custom role built in the role editor.

-> The API describes no create/delete for this object, so the data source is its only Terraform surface.

## Example Usage

```terraform
data "betterstack_role" "example" {
  name = "On-call lead"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) Role name.

### Read-Only

- `role` (String) Which built-in role this is. A team member's `role` has the same values.
  - `admin`: Can access every feature in every team, including billing, organization settings and API tokens.
  - `billing_admin`: Can manage subscriptions, invoices and payment settings for the whole organization, without access to operational features.
  - `team_lead`: Can access monitors, incidents, on-call, status pages, telemetry and AI SRE, invite team members and change their roles.
  - `responder`: Can access monitors, incidents, on-call, status pages, telemetry and AI SRE, and invite team members.
  - `member`: Can access sources, Live tail, dashboards, error tracking, session replays and AI SRE, without monitors, incidents or on-call.
  - `custom`: A role the organization made itself in the role editor.
