# Serve a status page from your own domain
resource "betterstack_status_page" "public" {
  company_name      = "Example, Inc."
  company_url       = "https://example.org"
  subdomain         = "example-inc-status"
  custom_domain     = "status.example.org"
  timezone          = "UTC"
  design            = "v2"
  theme             = "system"
  subscribable      = true
  automatic_reports = true
}

# Show monitors in sections
resource "betterstack_monitor" "website" {
  url                = "https://uptime.betterstack.com"
  pronounceable_name = "Website"
}

resource "betterstack_monitor" "api" {
  url                   = "https://uptime.betterstack.com/api/v2/monitors"
  monitor_type          = "expected_status_code"
  expected_status_codes = [401]
  pronounceable_name    = "API"
}

resource "betterstack_status_page" "services" {
  company_name = "Example, Inc."
  subdomain    = "example-inc-services"
  timezone     = "UTC"
  sections = [
    { name = "Website", resources = [{ resource_type = "Monitor", resource_id = betterstack_monitor.website.id, public_name = "Website", widget_type = "history" }] },
    { name = "API", resources = [{ resource_type = "Monitor", resource_id = betterstack_monitor.api.id, public_name = "Public API", widget_type = "response_times" }] },
  ]
}
