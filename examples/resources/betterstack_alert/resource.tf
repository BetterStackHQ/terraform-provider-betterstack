resource "betterstack_alert" "example" {
  name         = "High Error Rate"
  alert_type   = "threshold"
  dashboard_id = "123"
  chart_id     = "456"
}
