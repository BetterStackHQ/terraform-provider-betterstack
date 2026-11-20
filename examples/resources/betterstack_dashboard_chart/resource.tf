resource "betterstack_dashboard_chart" "example" {
  dashboard_id = "12345"
  chart_type   = "line_chart"
  queries      = [{ query_type = "example" }]
}
