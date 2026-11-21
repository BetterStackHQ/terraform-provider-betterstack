resource "betterstack_on_call_rotation" "example" {
  id                 = "default"
  users              = ["alice@example.com", "bob@example.com"]
  rotation_length    = 1
  rotation_interval  = "week"
  start_rotations_at = "2026-01-05T09:00:00Z"
}
