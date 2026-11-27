# Collect logs from Ubuntu servers and keep them for 30 days
resource "betterstack_source" "servers" {
  name              = "Production servers"
  platform          = "ubuntu"
  logs_retention    = 30
  metrics_retention = 90
}

# Receive OpenTelemetry data in a source group
resource "betterstack_source_group" "backend" {
  name = "Backend"
}

resource "betterstack_source" "otel" {
  name            = "Checkout service"
  platform        = "open_telemetry"
  source_group_id = betterstack_source_group.backend.id
}
