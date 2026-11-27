terraform {
  required_providers {
    betterstack = {
      source = "BetterStackHQ/betterstack"
    }
  }
}

provider "betterstack" {
  # Falls back to the BETTERSTACK_API_TOKEN environment variable.
  api_token = var.betterstack_api_token
  # team_id = "t123"
}

variable "betterstack_api_token" {
  type      = string
  sensitive = true
  default   = null
}
