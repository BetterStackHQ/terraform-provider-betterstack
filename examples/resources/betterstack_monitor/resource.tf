resource "betterstack_monitor" "example" {
  monitor_type = "status"
  url          = "https://uptime.betterstack.com"
}
