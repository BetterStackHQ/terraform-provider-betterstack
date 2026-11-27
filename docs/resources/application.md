---
page_title: "betterstack_application Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An error-tracking application collects the exceptions, releases and web events of one service.
---

# betterstack_application (Resource)

An error-tracking application collects the exceptions, releases and web events of one service.

~> Destroying an application deletes its data for good, so `deletion_protection` is on by default: Terraform refuses to destroy or replace it until you set `deletion_protection = false` and apply.

-> Replaces `logtail_errors_application`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

### Create an application

```terraform
resource "betterstack_application" "example" {
  name        = "new-web-app"
  platform    = "ruby_errors"
  data_region = "us"
}
```

### With source-code mapping

```terraform
resource "betterstack_application" "with_source_code_mapping" {
  name                     = "new-web-app"
  platform                 = "ruby_errors"
  data_region              = "us"
  github_repository_name   = "acme/web"
  code_mapping_stack_root  = "/usr/src/app/"
  code_mapping_source_root = "apps/backend/"
}
```

### In an application group

```terraform
resource "betterstack_application_group" "example" {
  name = "Production apps"
}

resource "betterstack_application" "in_application_group" {
  name                 = "new-web-app"
  platform             = "ruby_errors"
  data_region          = "us"
  application_group_id = betterstack_application_group.example.id
}
```

## Schema

### Required

- `name` (String) Application name.
- `platform` (String, Forces replacement) The SDK or framework the application reports errors from. See the [95 valid values](#values--platform).

### Optional

- `application_group_id` (String) The ID of the application group to put this application in, from [List application groups](https://betterstack.com/docs/errors/api/applications-groups/list/). Remove the argument to move it out of its group.
- `ingesting_paused` (Boolean) Whether ingestion is paused for this application. Defaults to `false`.
- `errors_retention` (Number, Computed) Days to keep errors. Accepts one of the retention periods your plan offers, or the value already set. Defaults to your plan's default retention, or to a longer one the team's other applications already use.
- `team_id` (String) The team that owns the application, such as `t123`. Defaults to the provider's `team_id`. Set it to create this object in another team. Changing it replaces the resource, because an object can't move to another team.
- `web_events_retention` (Number, Computed) Days to keep web events. Accepts one of the retention periods your plan offers, or the value already set. A connected application's web events retention is set on its connected source.
- `replays_retention` (Number, Computed) Days to keep session replays. Accepts one of the retention periods your plan offers, or the value already set. A change applies to every application connected to the same source. Only an application that records session replays accepts it.
- `data_region` (String, Forces replacement) The data region. It's `europe`, `us`, `singapore` or a dedicated cluster's name from [List data regions](https://betterstack.com/docs/errors/api/data-regions/). An application created with `correlate_with_source_id` uses that source's region, and a different region here is refused. Omit it to use the organization's default.
- `correlate_with_source_id` (String, Forces replacement) The telemetry source, from [List sources](https://betterstack.com/docs/logs/api/list-all-existing-sources/), whose logs and traces are correlated with these errors.
- `github_repository_name` (String) The linked GitHub repository for source-code mapping, as `owner/name`, such as `acme/web`. Connect it to your team with the [GitHub integration](https://betterstack.com/docs/errors/integrations/github/) first. Set it to an empty string to disconnect it. Can't be set together with `gitlab_repository_name`.
- `gitlab_repository_name` (String) The linked GitLab repository for source-code mapping, as `group/name`, such as `acme/web`, from the team's connected repositories. Set it to an empty string to disconnect it. Can't be set together with `github_repository_name`.
- `code_mapping_stack_root` (String) The absolute path to the root of the stack trace on your server, such as `/usr/src/app/`. Stack frames under it are mapped to files in the linked repository.
- `code_mapping_source_root` (String) The path from your repository's root to the deployed source code, such as `apps/backend/`. A stack frame's path under `code_mapping_stack_root` is looked up under this path in the linked repository.
- `custom_bucket` (Attributes) Your own S3-compatible bucket to archive data in, or `null` when the application uses Better Stack storage. Changing only `access_key_id` and `secret_access_key` rotates the bucket's key pair. Changing `name`, `endpoint` or `keep_data_after_retention`, or adding or removing it, replaces the resource. (see [below for nested schema](#nestedatt--custom_bucket))
- `vrl_transformation_exceptions` (String, Computed) The VRL program that groups incoming exceptions into errors. A new application starts with its platform's default grouping, which groups exceptions by error name and call stack on most platforms. Read more about [exception grouping](https://betterstack.com/docs/errors/using-the-product/exception-grouping/). Set it to an empty string to remove it.
- `vrl_transformation_replays` (String, Computed) The VRL program applied to session replays during ingestion. Set it to an empty string to remove it.
- `vrl_transformation_web_events` (String, Computed) The VRL program applied to web events, such as page views and web vitals, during ingestion. A new application starts with the default web events program. Set it to an empty string to remove it.
- `vrl_transformation_logs` (String, Computed) The VRL program applied to logs collected alongside errors, during ingestion. Read more about [VRL transformations](https://betterstack.com/docs/logs/using-logtail/transforming-ingested-data/logs-vrl/). Set it to an empty string to remove it.
- `vrl_transformation_spans` (String, Computed) The VRL program applied to spans collected alongside errors, during ingestion. Set it to an empty string to remove it.
- `spike_protection_baseline_multiplier` (Number) The baseline multiplier of spike protection, from 0 to 100. Lower values start sampling sooner, and higher values allow bigger bursts before sampling starts. It's `null` while spike protection is off. Remove the argument to turn spike protection off.
- `ai_prompt_template` (String, Computed) The AI prompt used to fix this application's errors. It holds the default prompt when the application has none of its own. A blank prompt is refused.
- `js_tag_configuration` (Attributes, Computed) Settings of the JavaScript tag, as on the application's **Frontend** page. It's `null` for an application without a JavaScript tag. (see [below for nested schema](#nestedatt--js_tag_configuration))
- `deletion_protection` (Boolean) Whether destroying or replacing this application is refused. Defaults to `true`. Set it to `false` and apply before you delete or replace it on purpose. It's kept in state and never sent to Better Stack.

### Read-Only

- `id` (String) The ID of this application.
- `js_tag_own_domain_status` (String) Whether the CNAME record of your own subdomain points to `proxy.betterstack.net`. It's `null` when `js_tag_configuration` has no `own_domain`.
  - `checking`: Better Stack hasn't checked the CNAME record since the subdomain last changed.
  - `missing`: The subdomain has no CNAME record.
  - `invalid`: The CNAME record doesn't point to `proxy.betterstack.net`.
  - `valid`: The CNAME record points to `proxy.betterstack.net`, so the tag is served from the subdomain.
- `source_types` (List of String) Row types the application holds, such as `exceptions`, `logs`, `metrics`, `replays`, `spans` or `web_events`, sorted by name.
- `parent_source_id` (String) The source this application is connected to, whose collections `query_collections` names. `null` when it isn't connected to a source.
- `query_collections` (Attributes) ClickHouse names to query the application's data by. (see [below for nested schema](#nestedatt--query_collections))
- `token` (String, Sensitive) Ingestion token. It's the DSN to configure your SDK with. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `js_tag_token` (String, Sensitive) Token of the browser tag, separate from `token`. It's the `?t=` value the `/b.js` browser tag is loaded with. `null` for a platform without a browser tag. To print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `table_name` (String) The application's table name, which identifies its data in SQL queries.
- `ingesting_host` (String) Host to send data to.
- `created_at` (String (RFC 3339)) When the application was created.
- `updated_at` (String (RFC 3339)) When the application was last updated.
- `team_name` (String) Name of the team that owns the application.

<a id="values--platform"></a>
### Valid values for `platform`

`aiohttp_errors`, `android_errors`, `angular_errors`, `asgi_errors`, `asp_dot_net_core_errors`, `asp_dot_net_errors`, `astro_errors`, `aws_lambda_dot_net_errors`, `aws_lambda_node_errors`, `aws_lambda_python_errors`, `azure_functions_node_errors`, `bottle_errors`, `bun_errors`, `celery_errors`, `chalice_errors`, `cli_errors`, `cloudflare_pages_errors`, `connect_errors`, `dart_errors`, `deno_errors`, `django_errors`, `dot_net_errors`, `dot_net_http_errors`, `dot_net_maui_errors`, `echo_errors`, `electron_errors`, `elixir_errors`, `ember_errors`, `express_errors`, `falcon_errors`, `fastapi_errors`, `fasthttp_errors`, `fastify_errors`, `fiber_errors`, `flask_errors`, `flutter_errors`, `gatsby_errors`, `gin_errors`, `go_errors`, `godot_errors`, `google_cloud_function_dot_net_errors`, `google_cloud_function_node_errors`, `google_cloud_function_python_errors`, `hapi_errors`, `ios_errors`, `iris_errors`, `java_errors`, `javascript_errors`, `koa_errors`, `kotlin_errors`, `laravel_errors`, `log4j_errors`, `logback_errors`, `macos_errors`, `minidump_errors`, `native_errors`, `negroni_errors`, `nest_js_errors`, `next_js_errors`, `node_errors`, `nuxt_errors`, `php_errors`, `powershell_errors`, `pyramid_errors`, `python_errors`, `qt_errors`, `quart_errors`, `rack_middleware_errors`, `rails_errors`, `react_errors`, `react_native_errors`, `react_router_framework_errors`, `remix_errors`, `rq_errors`, `ruby_errors`, `rust_errors`, `sanic_errors`, `serverless_python_errors`, `solid_errors`, `solidstart_errors`, `spring_boot_errors`, `spring_errors`, `starlette_errors`, `svelte_errors`, `sveltekit_errors`, `symfony_errors`, `tanstack_start_react_errors`, `tornado_errors`, `tryton_errors`, `unity_errors`, `unreal_engine_errors`, `vue_errors`, `windows_forms_errors`, `wpf_errors`, `wsgi_errors`

<a id="nestedatt--custom_bucket"></a>
### Nested Schema for `custom_bucket`

Optional:

- `endpoint` (String) The S3-compatible endpoint URL of the bucket. It can't be changed after the bucket is set.
- `access_key_id` (String) The access key ID Better Stack uses to write to the bucket.
- `secret_access_key` (String) The secret access key for `access_key_id`. It's write-only.
- `keep_data_after_retention` (Boolean) Whether data stays in the bucket after the retention period. It can't be changed after the bucket is set.

Read-Only:

- `name` (String) The bucket's name, which Better Stack reads from `endpoint`.

<a id="nestedatt--js_tag_configuration"></a>
### Nested Schema for `js_tag_configuration`

Optional:

- `collect` (Attributes) The features the JavaScript tag collects, each on or off. (see [below for nested schema](#nestedatt--js_tag_configuration--collect))
- `sampling` (Attributes) Sampling rates of the JavaScript tag, as percentages. (see [below for nested schema](#nestedatt--js_tag_configuration--sampling))
- `exclude_selectors` (List of String) CSS selectors of elements excluded from session replays and autocapture.
- `mask_selectors` (List of String) CSS selectors of elements whose text session replays mask.
- `shared_session_domains` (List of String) Your other domains to share sessions with.
- `accept_incoming_sessions` (Boolean) Continue sessions that arrive from your other domains.
- `decorate_forms` (Boolean) Also share sessions through forms that submit to your other domains.
- `custom_config` (String) JavaScript code the tag runs when it loads.
- `session_timeout_minutes` (Number) Session timeout in minutes, from 1 to 60.
- `events_batch_interval_seconds` (Number) Seconds between batches of events, at least 2.
- `own_domain` (String) Your own subdomain to serve the tag from. Its CNAME record must point to `proxy.betterstack.net`.

<a id="nestedatt--js_tag_configuration--collect"></a>
### Nested Schema for `js_tag_configuration.collect`

Optional:

- `errors` (Boolean) Collect frontend errors.
- `web_vitals` (Boolean) Record web vitals.
- `console` (Boolean) Collect `console.log` calls.
- `memory` (Boolean) Record JavaScript heap memory, in Chrome only.
- `analytics` (Boolean) Record website analytics.
- `anonymous_events` (Boolean) Capture anonymous user events.
- `replays` (Boolean) Record session replays.
- `identified_events` (Boolean) Capture identified user events.
- `fingerprint` (Boolean) Collect the browser fingerprint.

<a id="nestedatt--js_tag_configuration--sampling"></a>
### Nested Schema for `js_tag_configuration.sampling`

Optional:

- `errors` (Number) Percentage of errors sent, from 0 to 100.
- `replays` (Number) Percentage of sessions recorded as session replays, from 0 to 100.
- `web_events` (Number) Percentage of sessions that send web events, from 0 to 100.

<a id="nestedatt--query_collections"></a>
### Nested Schema for `query_collections`

Optional:

- `database` (String) The ClickHouse database that holds the application's data.
- `table` (String) The physical table, as `database.table`.
- `recent` (Map of String) The collection of the last 30 minutes for each row type the application holds, keyed by row type, such as `exceptions` or `metrics`.
- `s3` (String) The collection that holds older rows of every type.
- `app_filter` (String) The condition that keeps only the application's rows in its parent's raw collections. `null` unless the application is connected to a source.
- `metrics_app_filter` (String) The condition that keeps only the application's rows in its parent's metrics collection. `null` unless the application is connected to a source.

<a id="values--platform"></a>
### Valid values for `platform`

`aiohttp_errors`, `android_errors`, `angular_errors`, `asgi_errors`, `asp_dot_net_core_errors`, `asp_dot_net_errors`, `astro_errors`, `aws_lambda_dot_net_errors`, `aws_lambda_node_errors`, `aws_lambda_python_errors`, `azure_functions_node_errors`, `bottle_errors`, `bun_errors`, `celery_errors`, `chalice_errors`, `cli_errors`, `cloudflare_pages_errors`, `connect_errors`, `dart_errors`, `deno_errors`, `django_errors`, `dot_net_errors`, `dot_net_http_errors`, `dot_net_maui_errors`, `echo_errors`, `electron_errors`, `elixir_errors`, `ember_errors`, `express_errors`, `falcon_errors`, `fastapi_errors`, `fasthttp_errors`, `fastify_errors`, `fiber_errors`, `flask_errors`, `flutter_errors`, `gatsby_errors`, `gin_errors`, `go_errors`, `godot_errors`, `google_cloud_function_dot_net_errors`, `google_cloud_function_node_errors`, `google_cloud_function_python_errors`, `hapi_errors`, `ios_errors`, `iris_errors`, `java_errors`, `javascript_errors`, `koa_errors`, `kotlin_errors`, `laravel_errors`, `log4j_errors`, `logback_errors`, `macos_errors`, `minidump_errors`, `native_errors`, `negroni_errors`, `nest_js_errors`, `next_js_errors`, `node_errors`, `nuxt_errors`, `php_errors`, `powershell_errors`, `pyramid_errors`, `python_errors`, `qt_errors`, `quart_errors`, `rack_middleware_errors`, `rails_errors`, `react_errors`, `react_native_errors`, `react_router_framework_errors`, `remix_errors`, `rq_errors`, `ruby_errors`, `rust_errors`, `sanic_errors`, `serverless_python_errors`, `solid_errors`, `solidstart_errors`, `spring_boot_errors`, `spring_errors`, `starlette_errors`, `svelte_errors`, `sveltekit_errors`, `symfony_errors`, `tanstack_start_react_errors`, `tornado_errors`, `tryton_errors`, `unity_errors`, `unreal_engine_errors`, `vue_errors`, `windows_forms_errors`, `wpf_errors`, `wsgi_errors`

## Import

Import an existing application by its ID. Get the ID from the [`betterstack_application` data source](../data-sources/application.md), which finds it by `name`.

With Terraform 1.5 or later, use an `import` block:

```terraform
import {
  to = betterstack_application.example
  id = "12345"
}
```

Or run `terraform import`:

```shell
terraform import betterstack_application.example 12345
```
