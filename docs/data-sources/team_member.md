---
page_title: "betterstack_team_member Data Source - terraform-provider-betterstack"
subcategory: "Account"
description: |-
  A member of a team, or a pending invitation to join it, with the role they hold or would land on.
---

# betterstack_team_member (Data Source)

A member of a team, or a pending invitation to join it, with the role they hold or would land on.

## Example Usage

```terraform
data "betterstack_team_member" "example" {
  email = "alice@example.com"
}
```

## Schema

### Optional

- `email` (String) The e-mail address of the team member to look up. Set it, or `id` to the same value.
- `id` (String) The same value as `email`.
- `team_id` (String) Team the membership is in, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `first_name` (String) The member's first name, or `null` when they haven't set one. Absent on a pending invitation.
- `last_name` (String) The member's last name, or `null` when they haven't set one. Absent on a pending invitation.
- `role` (String) The member's built-in role in this team. A user can have a different role in each team. `null` on an invitation to a custom role. With neither `role` nor `role_id`, an invitation gets the responder role.
  - `admin`: Can access every feature in every team, including billing, organization settings and API tokens. The API never assigns it.
  - `billing_admin`: Can manage subscriptions, invoices and payment settings for the whole organization, without access to operational features.
  - `team_lead`: Can access monitors, incidents, on-call, status pages, telemetry and AI SRE, invite team members and change their roles.
  - `responder`: Can access monitors, incidents, on-call, status pages, telemetry and AI SRE, and invite team members.
  - `member`: Can access sources, Live tail, dashboards, error tracking, session replays and AI SRE, without monitors, incidents or on-call.
  - `custom`: The member has a custom role. Assign one with `role_id` instead.
- `role_id` (String) ID of the member's role, or of the role an invitation gives, custom roles included. Get the IDs via [List roles](https://betterstack.com/docs/uptime/api/roles/list/). When `role` is also given, both must name the same role.
- `mobile_app_platforms` (List of String) The platforms the member's mobile apps receive push notifications on, `ios` and `android`, or empty when none. Absent on a pending invitation.
- `invited_at` (String (RFC 3339)) When the invitation was created. Present only on a pending invitation.
- `created_at` (String (RFC 3339)) When the user account was created. Absent on a pending invitation.
- `deletion_protection` (Boolean) Whether destroying or replacing this team member is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.
