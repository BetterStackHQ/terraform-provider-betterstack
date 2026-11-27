resource "betterstack_incidents_outgoing_webhook_integration" "example" {
  url          = "https://example.com/webhooks/better-stack"
  trigger_type = "incident_change"
}
