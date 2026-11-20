resource "betterstack_heartbeat" "example" {
  name   = "Testing heartbeat"
  period = 10800
  grace  = 300
}
