resource "betterstack_metric_expression" "example" {
  source_id      = "12345"
  name           = "level"
  sql_expression = "JSONExtract(raw, 'level', 'Nullable(String)')"
}
