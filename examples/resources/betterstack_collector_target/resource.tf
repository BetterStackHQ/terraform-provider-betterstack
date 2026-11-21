resource "betterstack_collector_target" "example" {
  collector_id = "1"
  kind         = "postgres"
  host         = "db.example.com"
}
