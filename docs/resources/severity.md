---
page_title: "betterstack_severity Resource - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A severity, a named alerting level with its notification-channel settings.
---

# betterstack_severity (Resource)

A severity, a named alerting level with its notification-channel settings.

-> Replaces `betteruptime_severity`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create a severity

```terraform
resource "betterstack_severity" "example" {
  name  = "New severity"
  sms   = false
  call  = true
  email = false
  push  = true
}
```

### In a severity group

```terraform
resource "betterstack_severity_group" "example" {
  name = "Customer-facing"
}

resource "betterstack_severity" "in_severity_group" {
  name              = "New severity"
  sms               = false
  call              = true
  email             = false
  push              = true
  severity_group_id = betterstack_severity_group.example.id
}
```

## Schema

### Required

- `name` (String) The severity's name, unique within the team, ignoring case.

### Optional

- `call` (Boolean) Whether escalation policy steps with this severity alert people by phone call. Each person can override this in their own notification settings. Defaults to `false`.
- `sms` (Boolean) Whether escalation policy steps with this severity alert people by SMS. Each person can override this in their own notification settings. Defaults to `false`.
- `email` (Boolean) Whether escalation policy steps with this severity alert people by e-mail. Each person can override this in their own notification settings. Defaults to `true`.
- `push` (Boolean) Whether escalation policy steps with this severity alert people by push notification. Each person can override this in their own notification settings. Defaults to `true`.
- `critical_alert` (Boolean) Whether escalation policy steps with this severity send a critical alert, a push notification that ignores the mute switch and Do Not Disturb mode. Defaults to `false`.
- `severity_group_id` (String) The ID of the severity group this severity belongs to, or `null` when it's in none. Get the ID via [List severity groups](https://betterstack.com/docs/uptime/api/list-all-existing-severity-groups/). Remove the argument to clear it from its group.
- `team_id` (String) Team that owns this severity, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.

### Read-Only

- `id` (String) The ID of this severity.
- `team_name` (String) Name of the team that owns this severity.

## Import

Import an existing severity by its ID. Get the ID via [List severities](https://betterstack.com/docs/uptime/api/list-all-severities/). The [`betterstack_severity` data source](../data-sources/severity.md) also finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_severity.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_severity.example 12345
```
