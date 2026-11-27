# Connect a GCP project through the Better Stack GCP module
resource "betterstack_source" "gcp" {
  name     = "GCP production"
  platform = "gcp"
}

module "better_stack" {
  source         = "github.com/BetterStackHQ/gcp//terraform"
  project_id     = "my-gcp-project"
  source_token   = betterstack_source.gcp.token
  ingesting_host = betterstack_source.gcp.ingesting_host
}

resource "betterstack_source_gcp_project" "gcp" {
  source_id          = betterstack_source.gcp.id
  gcp_project_id     = module.better_stack.project_id
  gcp_project_number = module.better_stack.project_number
}

# Reuse a GCP project another source connected
resource "betterstack_source" "gcp_staging" {
  name     = "GCP staging"
  platform = "gcp"
}

resource "betterstack_source_gcp_project" "gcp_staging" {
  source_id      = betterstack_source.gcp_staging.id
  gcp_account_id = "1234"
}
