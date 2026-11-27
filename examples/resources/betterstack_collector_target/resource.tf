resource "betterstack_collector_target" "example" {
  collector_id = "1234"
  kind         = "postgres"
  host         = "db.example.com"
}
