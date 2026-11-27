# Page the on-call person, then the whole team after 10 minutes
data "betterstack_severity" "high" {
  name = "High severity"
}

resource "betterstack_escalation_policy" "on_call" {
  name = "On-call, then the whole team"
  steps = [
    { type = "escalation", wait_before = 0, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "current_on_call" }] },
    { type = "escalation", wait_before = 600, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "entire_team" }] },
  ]
}

# Post to Slack right away, page the on-call person after 3 minutes and repeat
resource "betterstack_escalation_policy" "slack_first" {
  name         = "Slack, then on-call"
  repeat_count = 3
  repeat_delay = 300
  steps = [
    { type = "escalation", wait_before = 0, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "all_slack_integrations" }] },
    { type = "escalation", wait_before = 180, severity_id = data.betterstack_severity.high.id, step_members = [{ type = "current_on_call" }] },
  ]
}
