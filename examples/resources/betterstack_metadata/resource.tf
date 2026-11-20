resource "betterstack_metadata" "example" {
  owner_type = "Monitor"
  owner_id   = "2"
  key        = "Owner"
  values     = [{ type = "example" }]
}
