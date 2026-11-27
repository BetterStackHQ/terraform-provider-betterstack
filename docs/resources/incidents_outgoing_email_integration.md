---
page_title: "betterstack_incidents_outgoing_email_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Sends incident-update e-mails to external addresses, with per-stage subject, title, and body templates.
---

# betterstack_incidents_outgoing_email_integration (Resource)

Sends incident-update e-mails to external addresses, with per-stage subject, title, and body templates.

## Example Usage

### E-mail a client's team with a custom subject

```terraform
resource "betterstack_incidents_outgoing_email_integration" "e_mail_a_client_s_team_with_a_custom_subject" {
  name                 = "Client distribution list"
  recipients           = [{ email = "noc@client.com" }, { email = "ops@client.com" }]
  on_incident_started  = true
  on_incident_resolved = true
  templates            = [{ incident_stage = "started", subject = "Incident: $NAME" }]
}
```

### E-mail executives only when incidents start and are resolved

```terraform
resource "betterstack_incidents_outgoing_email_integration" "e_mail_executives_only_when_incidents_start_and_are_resolved" {
  name                     = "Executive updates"
  recipients               = [{ email = "cto@example.com" }, { email = "vp-engineering@example.com" }]
  on_incident_acknowledged = false
  on_incident_reopened     = false
  on_incident_comment      = false
  templates                = [{ incident_stage = "started", subject = "[Outage] $NAME", body = "We detected a problem: $CAUSE. Follow along at $INCIDENT_URL." }, { incident_stage = "resolved", subject = "[Resolved] $NAME" }]
}
```

## Schema

### Required

- `recipients` (Attributes List) The people the integration e-mails. Each recipient is an object with an `email` key and up to 10 extra text keys, such as `first_name`, which templates render as `$RECIPIENT.<key>`. (see [below for nested schema](#nestedatt--recipients))

### Optional

- `name` (String) The integration's name, shown in the dashboard.
- `team_id` (String) Team that owns this integration, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `templates` (Attributes List, Computed) Custom subject, title and body for stages of an incident, one entry per stage. A blank subject, title or body uses the default template. (see [below for nested schema](#nestedatt--templates))
- `on_incident_started` (Boolean) E-mail the recipients when an incident starts. Defaults to `true`.
- `on_incident_acknowledged` (Boolean) E-mail the recipients when an incident is acknowledged. Defaults to `true`.
- `on_incident_resolved` (Boolean) E-mail the recipients when an incident is resolved. Defaults to `true`.
- `on_incident_reopened` (Boolean) E-mail the recipients when an incident is reopened. Defaults to `true`.
- `on_incident_comment` (Boolean) E-mail the recipients when a new comment is posted on an incident. Defaults to `true`.
- `notify_alongside_primary_responder` (Boolean) Whether to e-mail the recipients alongside the primary responder on incidents with no escalation policy. Defaults to `true`.

### Read-Only

- `id` (String) The ID of this outgoing e-mail integration.
- `paused` (Boolean) Whether the integration is paused. A paused integration sends no e-mails.
- `team_name` (String) Name of the team that owns this integration.

<a id="nestedatt--recipients"></a>
### Nested Schema for `recipients`

Optional:

- `email` (String) The recipient's e-mail address.

<a id="nestedatt--templates"></a>
### Nested Schema for `templates`

Optional:

- `incident_stage` (String) Incident stage this template applies to. Valid values: `started`, `acknowledged`, `resolved`, `reopened`, `comment`.
- `subject` (String) The e-mail subject, up to 500 characters. Supports incident variables such as `$NAME`, `$STATUS`, `$CAUSE` and `$INCIDENT_URL`, a metadata value as `$METADATA.Key`, all metadata as `$METADATA_ARRAY` and a recipient's attributes as `$RECIPIENT.<key>`. `null` uses the default template.
- `title` (String) The e-mail title, up to 500 characters. Supports the same variables as `subject`. `null` uses the default template.
- `body` (String) The e-mail body in Markdown, up to 20,000 characters. Supports the same variables as `subject`. `null` uses the default template.

## Import

Import an existing outgoing e-mail integration by its ID. Get the ID via [List outgoing e-mail integrations](https://betterstack.com/docs/uptime/api/list-all-outgoing-email-integrations/).

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_incidents_outgoing_email_integration.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_incidents_outgoing_email_integration.example 12345
```
