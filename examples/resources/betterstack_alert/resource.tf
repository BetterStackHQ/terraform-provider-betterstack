# Start an incident when errors pass 100 in 5 minutes
resource "betterstack_source" "app" {
  name     = "Production app"
  platform = "ubuntu"
}

resource "betterstack_dashboard" "app" {
  name = "Production app"
}

resource "betterstack_dashboard_chart" "errors" {
  dashboard_id = betterstack_dashboard.app.id
  chart_type   = "line_chart"
  name         = "Errors"
  queries = [{
    query_type = "sql_expression"
    sql_query  = "SELECT {{time}} AS time, count(*) AS value FROM {{source}} WHERE dt BETWEEN {{start_time}} AND {{end_time}} AND JSONExtract(raw, 'level', 'Nullable(String)') = 'error' GROUP BY time"
  }]
}

resource "betterstack_alert" "error_spike" {
  dashboard_id         = betterstack_dashboard.app.id
  chart_id             = betterstack_dashboard_chart.errors.id
  name                 = "Error spike"
  alert_type           = "threshold"
  operator             = "higher_than"
  value                = 100
  source_variable      = "source:${betterstack_source.app.table_name}"
  aggregation_interval = 300
  check_period         = 60
  confirmation_period  = 300
  recovery_period      = 300
  email                = true
  push                 = true
}
