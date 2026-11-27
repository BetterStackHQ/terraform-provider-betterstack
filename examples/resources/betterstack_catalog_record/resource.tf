resource "betterstack_catalog_record" "example" {
  catalog_relation_id = "12345"
  attributes          = [{ attribute = { id = "example" }, values = [{ type = "example" }] }]
}
