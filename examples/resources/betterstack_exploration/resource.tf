resource "betterstack_exploration" "example" {
  name    = "Error Rate Analysis"
  chart   = { chart_type = "line_chart" }
  queries = [{ query_type = "sql_expression" }]
}
