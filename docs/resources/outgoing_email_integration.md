---
page_title: "betterstack_outgoing_email_integration Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Sends incident-update e-mails to external addresses, with per-stage subject, title, and body templates.
---

# betterstack_outgoing_email_integration (Resource)

Sends incident-update e-mails to external addresses, with per-stage subject, title, and body templates.

## Example Usage

```terraform
resource "betterstack_outgoing_email_integration" "example" {
  recipients = [{ email = "test@example.com" }]
}
```

## Schema

### Required

- `recipients` (Attributes List, see [nested schema](#nestedatt--recipients)) Recipients. Each entry is an e-mail address, or an object with an email key plus any other text attributes to use in templates. At least one recipient is required.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `name` (String) Integration name. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `templates` (Attributes List, see [nested schema](#nestedatt--templates)) Custom subject, title and body for each incident stage, as `[{incident_stage, subject, title, body}]`. Stages in the list are created or updated, and stages left out are removed. Leave the list out to keep the current templates. A blank subject, title or body uses the default template. Removing this argument keeps its current value in Better Stack, since it has no default to reset to, and the plan warns about it.
- `on_incident_started` (Boolean) E-mail the recipients when an incident starts. Defaults to `true`.
- `on_incident_acknowledged` (Boolean) E-mail the recipients when an incident is acknowledged. Defaults to `true`.
- `on_incident_resolved` (Boolean) E-mail the recipients when an incident is resolved. Defaults to `true`.
- `on_incident_reopened` (Boolean) E-mail the recipients when an incident is reopened. Defaults to `true`.
- `on_incident_comment` (Boolean) E-mail the recipients when a new comment is posted on an incident. Defaults to `true`.
- `notify_alongside_primary_responder` (Boolean) Also e-mail the recipients alongside the primary on-call responder. Defaults to `true`.

### Read-Only

- `id` (String) The id of this outgoing_email_integration.
- `team_name` (String) Name of the team that owns this integration.
- `paused` (Boolean) Whether the integration is paused. A paused integration sends no e-mails.

<a id="nestedatt--recipients"></a>
### Nested Schema for `recipients`

- `email` (String)

<a id="nestedatt--templates"></a>
### Nested Schema for `templates`

- `incident_stage` (String) Incident stage this template applies to. Valid values: `started`, `acknowledged`, `resolved`, `reopened`, `comment`.
- `subject` (String)
- `title` (String)
- `body` (String)

## Endpoints

- `GET /api/v2/outgoing-emails`
- `GET /api/v2/outgoing-emails/:id`
- `POST /api/v2/outgoing-emails`
- `PATCH /api/v2/outgoing-emails/:id`
- `DELETE /api/v2/outgoing-emails/:id`

-> Each recipients entry is a plain e-mail address, or an object with an `email` key plus any other text attributes to use in templates.

## Import

```shell
terraform import betterstack_outgoing_email_integration.example 12345
```
