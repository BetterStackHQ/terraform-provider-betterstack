---
page_title: "betterstack_slack_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A connected Slack channel that receives incident notifications. Connect Slack in the Better Stack web app. The API reads and updates connected channels.
---

# betterstack_slack_integration (Data Source)

A connected Slack channel that receives incident notifications. Connect Slack in the Better Stack web app. The API reads and updates connected channels.

-> The API describes no create/delete for this object, so the data source is its only Terraform surface.

## Example Usage

```terraform
data "betterstack_slack_integration" "example" {
  slack_channel_name = "#better-stack-incidents"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `slack_channel_name` (String) Slack channel name.

### Read-Only

- `team_name` (String) Name of the team that owns this integration.
- `slack_team_id` (String) Slack workspace id.
- `slack_team_name` (String) Slack workspace name.
- `slack_channel_id` (String) Slack channel id.
- `slack_status` (String) Connection status. Valid values: `active`, `account_inactive`.
- `integration_type` (String) Notification style. Valid values: `legacy`, `verbose`, `thread`, `channel`.
- `channel_display_name` (String) Display name for the channel in Better Stack. Blank clears it.
- `channel_invitees` (Attributes List, see [nested schema](#nestedatt--channel_invitees)) Who is invited to incident channels. Replaces the whole list.
- `alert_options` (Attributes List, see [nested schema](#nestedatt--alert_options)) Teams and escalation policies whose incidents notify this channel. Sending it replaces the whole list. An alert channel left out of a team entry is off.
- `status_page_update_resources` (Attributes List, see [nested schema](#nestedatt--status_page_update_resources)) Resources whose status page updates are posted to this channel. Needs post_status_page_updates turned on. Sending it replaces the whole list.
- `private_channel` (Boolean) Create incident channels as private channels.
- `alert_all_teams_and_policies` (Boolean) Notify this channel for every team and escalation policy.
- `invite_new_slack_users` (Boolean) Invite new Slack users automatically.
- `on_call_notifications` (Boolean) Whether on-call changes notify this channel.
- `notify_on_pause` (Boolean) Post when a monitor is paused.
- `notify_on_resolve` (Boolean) Post when an incident is resolved.
- `direct_message_on_mention` (Boolean) Direct-message users mentioned in incidents.
- `ai_sre_on_mention` (Boolean) Let AI SRE answer when mentioned.
- `restrict_incident_creation_to_channel` (Boolean) Only create incidents from this channel.
- `post_timeline_events` (Boolean) Post incident timeline events.
- `post_incident_metadata` (Boolean) Post incident metadata.
- `post_status_page_updates` (Boolean) Post status page updates. Turning it off clears status_page_update_resources.
- `notify_alongside_primary_responder` (Boolean) Whether this channel is also notified alongside the primary on-call responder.

<a id="nestedatt--channel_invitees"></a>
### Nested Schema for `channel_invitees`

- `type` (String) What is invited. Valid values: `user`, `on_call_calendar`, `slack_user_group`.
- `id` (String) Id of the user, on-call calendar or Slack user group.

<a id="nestedatt--alert_options"></a>
### Nested Schema for `alert_options`

- `type` (String) Type of routing target. Responses use policy. Requests accept policy or escalation_policy. Valid values: `team`, `policy`, `escalation_policy`.
- `id` (String) A team's t-prefixed id (t123), or an escalation policy's numeric id.
- `name` (String) Team or escalation policy name. Ignored in requests.
- `team_id` (String) The escalation policy's team, as a t-prefixed id. Only on policy entries, and ignored in requests.
- `team_name` (String) The escalation policy's team name. Only on policy entries, and ignored in requests.
- `call` (Boolean) Alert by phone call. Only on team entries.
- `sms` (Boolean) Alert by SMS. Only on team entries.
- `email` (Boolean) Alert by e-mail. Only on team entries.
- `push` (Boolean) Alert by push notification. Only on team entries.
- `critical_alert` (Boolean) Alert by a critical push notification. Only on team entries.

<a id="nestedatt--status_page_update_resources"></a>
### Nested Schema for `status_page_update_resources`

- `type` (String) Resource type. Valid values: `status_page`, `monitor`, `heartbeat`, `monitor_group`, `heartbeat_group`, `webhook_integration`, `email_integration`, `incoming_webhook`, `amazon_sns_integration`, `manually_tracked_item`, `logs_chart`, `catalog_reference`.
- `id` (String) Resource id.

## Endpoints

- `GET /api/v3/slack-integrations`
- `GET /api/v3/slack-integrations/:id`
- `PATCH /api/v3/slack-integrations/:id`
