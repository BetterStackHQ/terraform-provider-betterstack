resource "betterstack_collector_target" "example" {
  collector_id = "12345"
  kind         = "postgres"
  host         = "db.example.com"
}
