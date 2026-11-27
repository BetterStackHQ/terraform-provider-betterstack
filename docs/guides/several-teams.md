---
page_title: "Manage several teams"
description: |-
  Set team_id on each resource to manage several teams with one provider block.
---

# Manage several teams

The provider's `team_id` is the team a resource is created in when the resource doesn't name one. To manage resources in several teams with one API token, set `team_id` on the resources that belong to another team. Changing a resource's `team_id` replaces it, because an object can't move to another team.

Write the team as its ID from Better Stack URLs, such as `t123` in `https://uptime.betterstack.com/team/t123/monitors`.

```terraform
provider "betterstack" {
  team_id = "t123"
}

resource "betterstack_heartbeat" "nightly_backup" {
  name   = "Nightly backup"
  period = 86400
  grace  = 3600
}

resource "betterstack_heartbeat" "settlement_job" {
  team_id = "t456"
  name    = "Settlement job"
  period  = 3600
  grace   = 600
}
```

Because the team is an argument, one `for_each` block or one module can create resources in several teams. Data sources take `team_id` too, to look an object up in another team.

A provider block without `team_id` works when the token can access a single team. A token that can access several teams needs `team_id` on the provider block, or on every resource that takes one.

## A new team

Create the team first, then use its ID. Apply the team on its own, then add its resources in a later change:

```terraform
resource "betterstack_team" "search" {
  name = "Search"
}

resource "betterstack_heartbeat" "indexer" {
  team_id = betterstack_team.search.id
  name    = "Indexer"
  period  = 3600
  grace   = 600
}
```
