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
  # better_stack_team_id = "my-team"
  # team_id = "my-team"
  # team_name = "my-team"
}
