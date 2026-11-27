---
page_title: "betterstack_incidents_outgoing_email_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Sends incident-update e-mails to external addresses, with per-stage subject, title, and body templates.
---

# betterstack_incidents_outgoing_email_integration (Data Source)

Sends incident-update e-mails to external addresses, with per-stage subject, title, and body templates.

## Example Usage

```terraform
data "betterstack_incidents_outgoing_email_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The ID of the outgoing e-mail integration to look up.

### Optional

- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `name` (String) The integration's name, shown in the dashboard.
- `recipients` (Attributes List) The people the integration e-mails. Each recipient is an object with an `email` key and up to 10 extra text keys, such as `first_name`, which templates render as `$RECIPIENT.<key>`. (see [below for nested schema](#nestedatt--recipients))
- `templates` (Attributes List) Custom subject, title and body for stages of an incident, one entry per stage. A blank subject, title or body uses the default template. (see [below for nested schema](#nestedatt--templates))
- `on_incident_started` (Boolean) E-mail the recipients when an incident starts.
- `on_incident_acknowledged` (Boolean) E-mail the recipients when an incident is acknowledged.
- `on_incident_resolved` (Boolean) E-mail the recipients when an incident is resolved.
- `on_incident_reopened` (Boolean) E-mail the recipients when an incident is reopened.
- `on_incident_comment` (Boolean) E-mail the recipients when a new comment is posted on an incident.
- `notify_alongside_primary_responder` (Boolean) Whether to e-mail the recipients alongside the primary responder on incidents with no escalation policy.
- `paused` (Boolean) Whether the integration is paused. A paused integration sends no e-mails.
- `team_name` (String) Name of the team that owns this integration.

<a id="nestedatt--recipients"></a>
### Nested Schema for `recipients`

Read-Only:

- `email` (String) The recipient's e-mail address.

<a id="nestedatt--templates"></a>
### Nested Schema for `templates`

Read-Only:

- `incident_stage` (String) Incident stage this template applies to. Valid values: `started`, `acknowledged`, `resolved`, `reopened`, `comment`.
- `subject` (String) The e-mail subject, up to 500 characters. Supports incident variables such as `$NAME`, `$STATUS`, `$CAUSE` and `$INCIDENT_URL`, a metadata value as `$METADATA.Key`, all metadata as `$METADATA_ARRAY` and a recipient's attributes as `$RECIPIENT.<key>`. `null` uses the default template.
- `title` (String) The e-mail title, up to 500 characters. Supports the same variables as `subject`. `null` uses the default template.
- `body` (String) The e-mail body in Markdown, up to 20,000 characters. Supports the same variables as `subject`. `null` uses the default template.
