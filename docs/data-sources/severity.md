---
page_title: "betterstack_severity Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity, a named alerting level with its notification-channel settings.
---

# betterstack_severity (Data Source)

A severity, a named alerting level with its notification-channel settings.

## Example Usage

```terraform
data "betterstack_severity" "example" {
  name = "Low Severity"
}
```

## Schema

### Lookup

Set `id`, or filter attributes matching exactly one record.

- `id` (String) Look up by id.
- `name` (String) The severity's name, unique within the team, ignoring case.

### Optional

- `team_id` (String) Team that owns this severity, such as `t123`. Defaults to the provider's `team_id`. Set it to look the object up in another team.

### Read-Only

- `call` (Boolean) Whether escalation policy steps with this severity alert people by phone call. Each person can override this in their own notification settings.
- `sms` (Boolean) Whether escalation policy steps with this severity alert people by SMS. Each person can override this in their own notification settings.
- `email` (Boolean) Whether escalation policy steps with this severity alert people by e-mail. Each person can override this in their own notification settings.
- `push` (Boolean) Whether escalation policy steps with this severity alert people by push notification. Each person can override this in their own notification settings.
- `critical_alert` (Boolean) Whether escalation policy steps with this severity send a critical alert, a push notification that ignores the mute switch and Do Not Disturb mode.
- `severity_group_id` (String) The ID of the severity group this severity belongs to, or `null` when it's in none. Get the ID via [List severity groups](https://betterstack.com/docs/uptime/api/list-all-existing-severity-groups/).
- `team_name` (String) Name of the team that owns this severity.
