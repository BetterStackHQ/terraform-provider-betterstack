# Check a website every 30 seconds
resource "betterstack_monitor" "website" {
  url                 = "https://uptime.betterstack.com"
  monitor_type        = "status"
  pronounceable_name  = "Uptime homepage"
  check_frequency     = 30
  confirmation_period = 60
  regions             = ["us", "eu"]
  ssl_expiration      = 14
  domain_expiration   = 14
  email               = true
  push                = true
  team_wait           = 180
}

# Look for a keyword, alert through an escalation policy and pause at night
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

resource "betterstack_monitor" "pricing" {
  url                  = "https://betterstack.com/pricing"
  monitor_type         = "keyword"
  required_keyword     = "Pricing"
  pronounceable_name   = "Pricing page"
  check_frequency      = 60
  recovery_period      = 300
  request_headers      = [{ name = "Accept-Language", value = "en" }]
  escalation_policy_id = betterstack_escalation_policy.on_call.id
  maintenance_from     = "02:00:00"
  maintenance_to       = "05:00:00"
  maintenance_timezone = "UTC"
}
