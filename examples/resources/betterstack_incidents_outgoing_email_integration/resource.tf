resource "betterstack_incidents_outgoing_email_integration" "example" {
  recipients = [{ email = "test@example.com" }]
}
