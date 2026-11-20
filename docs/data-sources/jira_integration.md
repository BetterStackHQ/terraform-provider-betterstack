---
page_title: "betterstack_jira_integration Data Source - terraform-provider-betterstack"
subcategory: "Incidents"
description: |-
  A Jira integration that creates and syncs issues for incidents. You connect Jira through OAuth in the Better Stack UI, not through the API.
---

# betterstack_jira_integration (Data Source)

A Jira integration that creates and syncs issues for incidents. You connect Jira through OAuth in the Better Stack UI, not through the API.

## Example Usage

```terraform
data "betterstack_jira_integration" "example" {
  id = "12345"
}
```

## Schema

### Required

- `id` (String) The id of the jira_integration to look up.

### Read-Only

- `name` (String) Integration name.
- `team_name` (String) Name of the team that owns this integration.
- `automatic_issue_creation` (Boolean) Whether to create a Jira issue for every incident automatically.
- `jira_project_key` (String) Key of the Jira project that issues are created in.
- `jira_issue_type_id` (String) ID of the Jira issue type for created issues.
- `jira_fields` (Dynamic) Extra Jira fields to set on created issues, keyed by Jira field id.

## Endpoints

- `GET /api/v2/jira-integrations`
- `GET /api/v2/jira-integrations/:id`
- `PATCH /api/v2/jira-integrations/:id`
