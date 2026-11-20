resource "betterstack_chart_alert" "example" {
  dashboard_id = 1
  chart_id     = 1
  name         = "High Error Rate"
  alert_type   = "threshold"
}
