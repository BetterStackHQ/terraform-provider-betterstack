---
page_title: "betterstack_pagerduty_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to a PagerDuty service.
---

# betterstack_pagerduty_integration (Data Source)

Forwards Better Stack incidents to a PagerDuty service.

## Example Usage

```terraform
data "betterstack_pagerduty_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the pagerduty_integration to look up.

### Read-Only

- `name` (String) Integration name.
- `team_name` (String) Name of the team that owns this integration.
- `key` (String, Sensitive) PagerDuty integration key, also called the routing key.
- `severity` (String) PagerDuty severity sent with the event. Valid values: `info`, `warning`, `error`, `critical`.
- `notify_alongside_primary_responder` (Boolean) Also notify the primary on-call responder.

## Endpoints

- `GET /api/v2/pager-duty-webhooks`
- `GET /api/v2/pager-duty-webhooks/:id`
- `POST /api/v2/pager-duty-webhooks`
- `PATCH /api/v2/pager-duty-webhooks/:id`
- `DELETE /api/v2/pager-duty-webhooks/:id`
