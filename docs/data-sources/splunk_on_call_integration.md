---
page_title: "betterstack_splunk_on_call_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.
---

# betterstack_splunk_on_call_integration (Data Source)

Forwards Better Stack incidents to Splunk On-Call, formerly VictorOps.

## Example Usage

```terraform
data "betterstack_splunk_on_call_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the splunk_on_call_integration to look up.

### Read-Only

- `name` (String) Integration name.
- `team_name` (String) Name of the team that owns this integration.
- `url` (String, Sensitive) Splunk On-Call REST endpoint URL.
- `notify_alongside_primary_responder` (Boolean) Also notify the primary on-call responder.

## Endpoints

- `GET /api/v2/splunk-on-calls`
- `GET /api/v2/splunk-on-calls/:id`
- `POST /api/v2/splunk-on-calls`
- `PATCH /api/v2/splunk-on-calls/:id`
- `DELETE /api/v2/splunk-on-calls/:id`
