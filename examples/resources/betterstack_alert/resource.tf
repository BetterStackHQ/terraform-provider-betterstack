resource "betterstack_alert" "example" {
  dashboard_id = "123"
  chart_id     = "456"
  name         = "High Error Rate"
  alert_type   = "threshold"
}
