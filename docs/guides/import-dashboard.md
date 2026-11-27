---
page_title: "Import a dashboard built in the web app"
description: |-
  Bring a dashboard you built in the web app, with its charts, sections and alerts, under Terraform without recreating it.
---

# Import a dashboard built in the web app

Build a dashboard in the web app, then bring it under Terraform as it is: the same dashboard, charts, sections and alerts, with their ids, links and alert history. Terraform 1.14 or later finds every chart, section and alert with `terraform query` and writes their configuration for you.

## 1. Import the dashboard

Open the dashboard in the web app. Its id is the number in the page address, for example `1234` in `https://telemetry.betterstack.com/team/t123/dashboards/1234`. Add an `import` block for it:

```terraform
import {
  to = betterstack_dashboard.service
  id = "1234"
}
```

## 2. List its charts, sections and alerts

Create a file ending in `.tfquery.hcl`, for example `dashboard.tfquery.hcl`, next to your configuration:

```terraform
list "betterstack_dashboard_chart" "charts" {
  provider = betterstack
  config {
    dashboard_id = "1234"
  }
}

list "betterstack_dashboard_section" "sections" {
  provider = betterstack
  config {
    dashboard_id = "1234"
  }
}
```

Run `terraform query` to see what it finds. Each line shows a chart's or section's name and the id it is imported by.

Alerts belong to a chart. To import them too, add a block per chart with alerts, with the chart's id from the list above:

```terraform
list "betterstack_alert" "latency_alerts" {
  provider = betterstack
  config {
    dashboard_id = "1234"
    chart_id     = "5678"
  }
}
```

## 3. Generate the configuration

```shell
terraform plan -generate-config-out=dashboard.tf
terraform query -generate-config-out=charts.tf
```

The first command writes the dashboard's resource block from its `import` block. The second writes a resource block and an `import` block for every chart, section and alert the list found.

Review the files and replace each `dashboard_id = "1234"` with `dashboard_id = betterstack_dashboard.service.id`, so the charts and sections follow the dashboard. Rename the resources if you like: the `import` blocks point at them by name.

## 4. Apply

Run `terraform plan`. It should only import: no changes, nothing added or destroyed. Then run `terraform apply`. You can remove the `import` blocks afterwards.

From now on, change the dashboard in Terraform. A change made in the web app shows as a difference in the next plan.
