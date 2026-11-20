---
page_title: "betterstack_application Resource - terraform-provider-betterstack"
subcategory: "Errors"
description: |-
  An error-tracking application: a stream of exceptions, releases, and web events from one service.
---

# betterstack_application (Resource)

An error-tracking application: a stream of exceptions, releases, and web events from one service.

~> While `deletion_protection` is `true`, destroying this resource or changing an attribute that replaces it fails with an error, in a plan or in a `pulumi destroy`: deleting it permanently deletes everything stored in it. To delete it on purpose, set `deletion_protection = false` and apply that first. To stop managing it without deleting it, use a `removed` block with `lifecycle { destroy = false }`.

-> Replaces `logtail_errors_application`. Existing state moves over with a `moved {}` block; see the [migration guide](../guides/migration.md).

## Example Usage

```terraform
resource "betterstack_application" "example" {
  name     = "example"
  platform = "react_errors"
}
```

## Schema

### Required

- `name` (String) Application name.
- `platform` (String) SDK/framework platform. Fixed at creation. Set to a value the errors platform catalog doesn't know and it is still stored and honored end-to-end; the model itself only rejects blank or "collector". Valid values: `react_errors`, `vue_errors`, `angular_errors`, `react_native_errors`, `ember_errors`, `solid_errors`, `svelte_errors`, `sveltekit_errors`, `next_js_errors`, `nuxt_errors`, `remix_errors`, `gatsby_errors`, `astro_errors`, `solidstart_errors`, `javascript_errors`, `dart_errors`, `tanstack_start_react_errors`, `react_router_framework_errors`, `unity_errors`, `android_errors`, `ios_errors`, `flutter_errors`, `asp_dot_net_errors`, `asp_dot_net_core_errors`, `dot_net_errors`, `dot_net_http_errors`, `django_errors`, `fastapi_errors`, `flask_errors`, `falcon_errors`, `bottle_errors`, `asgi_errors`, `sanic_errors`, `starlette_errors`, `tornado_errors`, `chalice_errors`, `wsgi_errors`, `python_errors`, `express_errors`, `fastify_errors`, `nest_js_errors`, `koa_errors`, `hapi_errors`, `node_errors`, `bun_errors`, `deno_errors`, `laravel_errors`, `symfony_errors`, `php_errors`, `spring_errors`, `spring_boot_errors`, `java_errors`, `rails_errors`, `rack_middleware_errors`, `ruby_errors`, `echo_errors`, `gin_errors`, `fiber_errors`, `go_errors`, `negroni_errors`, `rust_errors`, `elixir_errors`, `powershell_errors`, `rq_errors`, `celery_errors`, `quart_errors`, `pyramid_errors`, `connect_errors`, `iris_errors`, `tryton_errors`, `log4j_errors`, `logback_errors`, `aiohttp_errors`, `fasthttp_errors`, `dot_net_maui_errors`, `unreal_engine_errors`, `qt_errors`, `electron_errors`, `macos_errors`, `windows_forms_errors`, `wpf_errors`, `godot_errors`, `minidump_errors`, `aws_lambda_dot_net_errors`, `aws_lambda_node_errors`, `aws_lambda_python_errors`, `azure_functions_node_errors`, `google_cloud_function_dot_net_errors`, `google_cloud_function_node_errors`, `google_cloud_function_python_errors`, `serverless_python_errors`, `cloudflare_pages_errors`, `kotlin_errors`, `native_errors`, `cli_errors`. Can only be set when creating the resource; changing it afterwards replaces the resource.

### Optional

Optional attributes are also computed: left unset, they take the server-side default.

- `application_group_id` (Number) Application group this application belongs to. Send null to ungroup; an unknown or out-of-team group id 404s rather than accepting it (unlike the legacy API, "0" no longer ungroups). Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `ingesting_paused` (Boolean) Pause ingestion for this application. Defaults to `false`.
- `errors_retention` (Number) Days to retain errors: one of the windows your plan offers, or the value already stored.
- `data_region` (String) The data region, as the source picker names it: `europe`, `us`, `singapore` or a dedicated cluster's name. It reads back as that name. The older ids `germany`, `us_east` and `us_west` are also accepted. An application created with `correlate_with_source_id` takes that source's region, and a different region here is refused. Set on create only: sending it on update is refused. Omit it to use the organization's default. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `correlate_with_source_id` (Number) Telemetry source whose logs and traces are correlated with these errors. Fixed at creation; sending it on update always 422s, even to the same value. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `github_repository_name` (String) Linked GitHub repository (owner/name) for source-code mapping, from the team's connected repositories. Mutually exclusive with gitlab_repository_name; an empty string disconnects it. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `gitlab_repository_name` (String) Linked GitLab repository (group/name) for source-code mapping, from the team's connected repositories. Mutually exclusive with github_repository_name; an empty string disconnects it. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `code_mapping_stack_root` (String) Stack-trace root path used to map an exception back to a source file. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `code_mapping_source_root` (String) Repository source root used to map an exception back to a source file. Removing this argument after configuring it sends JSON null on update; bridged SDKs use the same removal semantics.
- `custom_bucket` (Attributes, see [nested schema](#nestedatt--custom_bucket)) Custom S3-compatible archival bucket, set on create when your plan includes custom buckets. secret_access_key is write-only and never returned. An update may send only access_key_id and secret_access_key together, which rotates the key pair after probing it against the bucket. Can only be set when creating the resource; changing it afterwards replaces the resource.
- `vrl_transformation_exceptions` (String) Server-side VRL program applied to exception events. Reads back the seeded platform default when no program has been set, never null.
- `vrl_transformation_replays` (String) Server-side VRL program applied to session-replay events.
- `vrl_transformation_web_events` (String) Server-side VRL program applied to web-vitals/web events.
- `vrl_transformation_logs` (String) Server-side VRL program applied to log events collected alongside errors.
- `vrl_transformation_spans` (String) Server-side VRL program applied to trace spans collected alongside errors.
- `deletion_protection` (Boolean) While true (the default), destroying or replacing this application fails, in a plan or in a pulumi destroy. Kept in Terraform or Pulumi state only: the API never sees it. Set it to false and apply that before deleting or replacing it on purpose.

### Read-Only

- `id` (String) The id of this application.
- `token` (String, Sensitive) Ingestion token (the SDK's DSN). Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `js_tag_token` (String, Sensitive) Browser-tag token, distinct from token: the ?t= value the /b.js browser tag is loaded with. Null for a platform with no browser tag. Terraform hides this value in plans and output. Referencing it from another resource works as is; to print it, declare an output with `sensitive = true` and run `terraform output -raw <name>`.
- `table_name` (String) Underlying table name.
- `ingesting_host` (String) Host to send data to.
- `team_id` (String) Team that owns the application, as its t-prefixed Better Stack id (t123).
- `team_name` (String) Name of the team that owns the application.
- `created_at` (String (RFC 3339)) When the application was created.
- `updated_at` (String (RFC 3339)) When the application was last updated.

<a id="nestedatt--custom_bucket"></a>
### Nested Schema for `custom_bucket`

- `name` (String)
- `endpoint` (String)
- `access_key_id` (String)
- `secret_access_key` (String)
- `keep_data_after_retention` (Boolean)

## Endpoints

- `GET /api/v4/applications`
- `GET /api/v4/applications/:id`
- `POST /api/v4/applications`
- `PATCH /api/v4/applications/:id`
- `DELETE /api/v4/applications/:id`

-> correlate_with_source_id names a telemetry source (GET /api/v4/sources on telemetry.betterstack.com) whose logs and traces are correlated with these errors.

## Import

```shell
terraform import betterstack_application.example 12345
```
