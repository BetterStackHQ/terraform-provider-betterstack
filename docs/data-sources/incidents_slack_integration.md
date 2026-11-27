---
page_title: "betterstack_incidents_slack_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A connected Slack channel that receives incident notifications.
---

# betterstack_incidents_slack_integration (Data Source)

A connected Slack channel that receives incident notifications.

-> The API describes no create/delete for this object, so the data source is its only Terraform surface.

## Example Usage

```terraform
data "betterstack_incidents_slack_integration" "example" {
  slack_channel_name = "#better-stack-incidents"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `slack_channel_name` (String) Slack channel name.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `slack_team_id` (String) Slack workspace id.
- `slack_team_name` (String) Slack workspace name.
- `slack_channel_id` (String) Slack channel id.
- `slack_status` (String) Whether the connected Slack account is active.
  - `active`: The connected Slack account is active.
  - `account_inactive`: Slack reports the connected account as inactive.
- `integration_type` (String) How incidents are posted to Slack.
  - `legacy`: Posts every incident to the connected channel, as `verbose` does. Slack users can't alert escalation policies from it.
  - `verbose`: Posts every incident to the connected channel.
  - `thread`: Starts a thread for each incident in the connected channel. Needs a paid plan.
  - `channel`: Creates a new channel for each incident. Needs a paid plan.
- `channel_display_name` (String) Display name for the channel in Better Stack. Blank clears it.
- `channel_invitees` (Attributes List) Who is invited to incident channels. Replaces the whole list. (see [below for nested schema](#nestedatt--channel_invitees))
- `alert_options` (Attributes List) Teams and escalation policies whose incidents notify this channel. An update replaces the whole list. An alert channel left out of a team entry is off. (see [below for nested schema](#nestedatt--alert_options))
- `status_page_update_resources` (Attributes List) Resources whose status page updates are posted to this channel. Empty means every resource of the integration's team. Needs `post_status_page_updates` turned on. An update replaces the whole list. (see [below for nested schema](#nestedatt--status_page_update_resources))
- `private_channel` (Boolean) Create incident channels as private channels. Needs at least one entry in `channel_invitees`.
- `alert_all_teams_and_policies` (Boolean) Whether Slack users can alert every team and escalation policy in the organization, or only the ones in `alert_options`.
- `invite_new_slack_users` (Boolean) Whether people who join the Slack workspace are invited into the integration's Better Stack team as responders.
- `on_call_notifications` (Boolean) Whether the channel is notified when the primary responder changes.
- `notify_on_pause` (Boolean) Post when a monitor, heartbeat, group or integration is paused.
- `notify_on_resolve` (Boolean) Post when an incident is resolved.
- `direct_message_on_mention` (Boolean) Direct-message users mentioned in incidents.
- `ai_sre_on_mention` (Boolean) Whether tagging Better Stack in Slack starts an AI SRE conversation.
- `restrict_incident_creation_to_channel` (Boolean) Whether only the members of this channel can create incidents.
- `post_timeline_events` (Boolean) Post incident timeline events about e-mails, phone calls and integrations.
- `post_incident_metadata` (Boolean) Post incident metadata.
- `post_status_page_updates` (Boolean) Post status page updates. Turning it off clears `status_page_update_resources`.
- `notify_alongside_primary_responder` (Boolean) Whether this channel is notified alongside the primary responder on incidents with no escalation policy.
- `team_name` (String) Name of the team that owns this integration.

<a id="nestedatt--channel_invitees"></a>
### Nested Schema for `channel_invitees`

Read-Only:

- `type` (String) What is invited. Valid values: `user`, `on_call_calendar`, `slack_user_group`.
- `id` (String) ID of the on-call calendar or Slack user group. A user is `u123` or their e-mail, and reads back as `u123`. Get a user's ID via [List team members](https://betterstack.com/docs/uptime/api/team-members/list/).

<a id="nestedatt--alert_options"></a>
### Nested Schema for `alert_options`

Read-Only:

- `type` (String) Type of routing target.
  - `team`: A team, alerted through the channels turned on in this entry.
  - `policy`: An escalation policy. Responses use this value.
  - `escalation_policy`: An escalation policy. Requests accept it as well as `policy`.
- `id` (String) A team's ID from [List teams](https://betterstack.com/docs/uptime/api/teams/list/), such as `t123`, or an escalation policy's ID from [List escalation policies](https://betterstack.com/docs/uptime/api/list-all-escalation-policies/).
- `name` (String) Team or escalation policy name.
- `team_id` (String) The escalation policy's team, as its team ID such as `t123`. Only on policy entries.
- `team_name` (String) The escalation policy's team name. Only on policy entries.
- `call` (Boolean) Alert by phone call. Only on team entries.
- `sms` (Boolean) Alert by SMS. Only on team entries.
- `email` (Boolean) Alert by e-mail. Only on team entries.
- `push` (Boolean) Alert by push notification. Only on team entries.
- `critical_alert` (Boolean) Alert by a critical push notification. Only on team entries.

<a id="nestedatt--status_page_update_resources"></a>
### Nested Schema for `status_page_update_resources`

Read-Only:

- `type` (String) Resource type. Valid values: `status_page`, `monitor`, `heartbeat`, `monitor_group`, `heartbeat_group`, `webhook_integration`, `email_integration`, `incoming_webhook`, `amazon_sns_integration`, `manually_tracked_item`, `logs_chart`, `catalog_reference`.
- `id` (String) Resource id.
