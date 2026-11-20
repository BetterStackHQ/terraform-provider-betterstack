data "betterstack_source_gcp_metric" "example" {
  source_id   = "12345"
  metric_type = "compute.googleapis.com/instance/cpu/utilization"
}
