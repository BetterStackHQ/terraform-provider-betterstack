resource "betterstack_metric" "example" {
  source_id      = "123"
  name           = "level"
  sql_expression = "JSONExtract(raw, 'level', 'Nullable(String)')"
}
