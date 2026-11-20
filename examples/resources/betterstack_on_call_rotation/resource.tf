resource "betterstack_on_call_rotation" "example" {
  id                 = "example"
  users              = ["example"]
  rotation_length    = 1
  rotation_interval  = "hour"
  start_rotations_at = "2026-01-01T00:00:00Z"
  end_rotations_at   = "2026-01-01T00:00:00Z"
}
