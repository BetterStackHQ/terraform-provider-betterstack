# Expect a nightly job once a day
resource "betterstack_heartbeat" "nightly_backup" {
  name   = "Nightly database backup"
  period = 86400
  grace  = 7200
  email  = true
  push   = true
}

# Alert through an escalation policy, except during weekend maintenance
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

resource "betterstack_heartbeat" "queue_worker" {
  name                 = "Queue worker"
  period               = 300
  grace                = 300
  escalation_policy_id = betterstack_escalation_policy.on_call.id
  maintenance_days     = ["sat", "sun"]
  maintenance_from     = "00:00:00"
  maintenance_to       = "06:00:00"
  maintenance_timezone = "London"
}
