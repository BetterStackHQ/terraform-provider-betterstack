resource "betterstack_incidents_outgoing_pagerduty_integration" "example" {
  name     = "PagerDuty alerts"
  key      = "routing-key-placeholder"
  severity = "critical"
}
