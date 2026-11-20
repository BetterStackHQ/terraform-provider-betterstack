package provider

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Opt in with BETTERSTACK_MIGRATION_FIXTURES=1, BETTERSTACK_LEGACY_UPTIME_DIR and
// BETTERSTACK_LEGACY_LOGTAIL_DIR pointing to reference clones containing the pinned release tags.
// Requires terraform >=1.8, git and Go. This creates REAL SDK-provider state in a private temporary
// directory against a literal isolated API, then uses Terraform's cross-provider move protocol.
// The only legacy source modification is the test entrypoint selecting the existing WithURL hook.
// No fixture state, credentials, Terraform invocation, or provider binary is uploaded anywhere.
func TestAccLegacyStateMigration(t *testing.T) {
	if os.Getenv("BETTERSTACK_MIGRATION_FIXTURES") != "1" {
		t.Skip("set BETTERSTACK_MIGRATION_FIXTURES=1 to exercise real Terraform-created legacy state")
	}
	terraform, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal("Terraform 1.8+ is required")
	}
	var version struct {
		Version string `json:"terraform_version"`
	}
	if err := json.Unmarshal(migrationCommand(t, "", nil, terraform, "version", "-json"), &version); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(version.Version, ".")
	if len(parts) < 2 {
		t.Fatal("cannot determine Terraform version")
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	if major < 1 || (major == 1 && minor < 8) {
		t.Fatal("Terraform 1.8+ is required for cross-provider moves")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, release := range []struct{ name, uptime, logtail string }{
		{"current", "v0.22.2", "v11.2.0"},
		{"previous", "v0.21.11", "v10.15.14"},
	} {
		t.Run(release.name, func(t *testing.T) {
			dir := t.TempDir()
			mirror := filepath.Join(dir, "mirror")
			buildLegacyMigrationProvider(t, os.Getenv("BETTERSTACK_LEGACY_UPTIME_DIR"), release.uptime, "better-uptime", mirror)
			buildLegacyMigrationProvider(t, os.Getenv("BETTERSTACK_LEGACY_LOGTAIL_DIR"), release.logtail, "logtail", mirror)
			target := migrationProviderPath(t, mirror, "betterstackhq", "betterstack", "0.1.0")
			migrationCommand(t, root, nil, "go", "build", "-o", target, ".")
			cliConfig := filepath.Join(dir, "terraform.rc")
			migrationWrite(t, cliConfig, fmt.Sprintf("provider_installation {\n filesystem_mirror {\n  path = %q\n }\n}\n", filepath.ToSlash(mirror)))
			api := &migrationFixtureAPI{objects: map[string]map[string]interface{}{}}
			server := httptest.NewServer(api)
			defer server.Close()
			env := []string{"TF_CLI_CONFIG_FILE=" + cliConfig, "TF_IN_AUTOMATION=1", "CHECKPOINT_DISABLE=1", "BETTERSTACK_MIGRATION_FIXTURE_ENDPOINT=" + server.URL}
			work := filepath.Join(dir, "state")
			if err := os.Mkdir(work, 0700); err != nil {
				t.Fatal(err)
			}
			config := filepath.Join(work, "main.tf")
			migrationWrite(t, config, fmt.Sprintf(legacyMigrationConfig, strings.TrimPrefix(release.uptime, "v"), strings.TrimPrefix(release.logtail, "v")))
			migrationCommand(t, work, env, terraform, "init", "-input=false", "-no-color")
			migrationCommand(t, work, env, terraform, "apply", "-auto-approve", "-input=false", "-no-color")
			before := migrationShowState(t, work, env, terraform)
			assertMigrationValue(t, before, "logtail_source.secret", "custom_bucket.0.secret_access_key", "fixture-bucket-secret")
			assertMigrationValue(t, before, "logtail_source.secret", "scrape_request_basic_auth_password", "fixture-scrape-secret")
			assertMigrationValue(t, before, "logtail_exploration.saved", "query.0.id", json.Number("707"))
			assertMigrationValue(t, before, "logtail_exploration.saved", "query.0.sql_query", "SELECT count() AS value")
			api.mu.Lock()
			api.readOnly = true
			api.mu.Unlock()
			migrationWrite(t, config, fmt.Sprintf(targetMigrationConfig, server.URL))
			migrationCommand(t, work, env, terraform, "init", "-input=false", "-no-color")
			migrationCommand(t, work, env, terraform, "plan", "-input=false", "-no-color", "-out=migration.plan")
			var plan struct {
				Changes []struct {
					Address  string `json:"address"`
					Previous string `json:"previous_address"`
					Change   struct {
						Actions []string `json:"actions"`
					} `json:"change"`
				} `json:"resource_changes"`
			}
			if err := json.Unmarshal(migrationCommand(t, work, env, terraform, "show", "-json", "migration.plan"), &plan); err != nil {
				t.Fatal(err)
			}
			moves := 0
			for _, change := range plan.Changes {
				if change.Previous != "" {
					moves++
				}
				if len(change.Change.Actions) != 1 || change.Change.Actions[0] != "no-op" {
					t.Fatalf("migration proposed a remote mutation for %s", change.Address)
				}
			}
			if moves != 5 {
				t.Fatalf("expected five state moves, got %d", moves)
			}
			migrationCommand(t, work, env, terraform, "apply", "-input=false", "-no-color", "migration.plan")
			after := migrationShowState(t, work, env, terraform)
			for old, current := range map[string]string{"betteruptime_monitor_group.same": "betterstack_monitor_group.same", "logtail_source_group.same": "betterstack_source_group.same", "betteruptime_datadog_integration.same": "betterstack_datadog_integration.same", "logtail_source.secret": "betterstack_source.secret", "logtail_exploration.saved": "betterstack_exploration.saved"} {
				assertMigrationValue(t, after, current, "id", before[old]["id"])
				if _, exists := after[old]; exists {
					t.Fatalf("both old and new addresses own %s", old)
				}
			}
			assertMigrationValue(t, after, "betterstack_datadog_integration.same", "webhook_url", before["betteruptime_datadog_integration.same"]["webhook_url"])
			assertMigrationValue(t, after, "betterstack_source.secret", "source_group_id", before["logtail_source_group.same"]["id"])
			assertMigrationValue(t, after, "betterstack_source.secret", "custom_bucket.secret_access_key", "fixture-bucket-secret")
			assertMigrationValue(t, after, "betterstack_source.secret", "scrape_request_basic_auth_password", "fixture-scrape-secret")
			assertMigrationValue(t, after, "betterstack_source.secret", "data_region", "germany")
			assertMigrationValue(t, after, "betterstack_exploration.saved", "chart.name", "Fixture exploration")
			assertMigrationValue(t, after, "betterstack_exploration.saved", "chart.settings.unit", "short")
			assertMigrationValue(t, after, "betterstack_exploration.saved", "chart.settings.decimal_places", json.Number("0"))
			assertMigrationValue(t, after, "betterstack_exploration.saved", "queries", []interface{}{map[string]interface{}{
				"name": "count", "query_type": "sql_expression", "sql_query": "SELECT count() AS value", "source_variable": "source",
			}})
			assertMigrationValue(t, after, "betterstack_exploration.saved", "queries_result.0.id", json.Number("707"))
			assertMigrationValue(t, after, "betterstack_exploration.saved", "queries_result.0.sql_query", before["logtail_exploration.saved"]["query"].([]interface{})[0].(map[string]interface{})["sql_query"])
			// With -detailed-exitcode an ordinary exec success is proof of a no-op (2 means changes).
			migrationCommand(t, work, env, terraform, "plan", "-detailed-exitcode", "-input=false", "-no-color")
			api.mu.Lock()
			defer api.mu.Unlock()
			if api.refusedWrites != 0 {
				t.Fatalf("migration attempted %d remote writes", api.refusedWrites)
			}
		})
	}
}

func migrationCommand(t *testing.T, dir string, env []string, binary string, args ...string) []byte {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Dir = dir
	command.Env = append(os.Environ(), env...)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.ReplaceAll(strings.ReplaceAll(string(output), "fixture-bucket-secret", "[redacted]"), "fixture-scrape-secret", "[redacted]")
		t.Fatalf("%s %s failed: %v\n%s", filepath.Base(binary), strings.Join(args, " "), err, message)
	}
	return output
}

func migrationWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}

func migrationProviderPath(t *testing.T, mirror, namespace, provider, version string) string {
	t.Helper()
	dir := filepath.Join(mirror, "registry.terraform.io", namespace, provider, version, runtime.GOOS+"_"+runtime.GOARCH)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	name := "terraform-provider-" + provider + "_v" + version
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

func buildLegacyMigrationProvider(t *testing.T, repository, tag, name, mirror string) {
	t.Helper()
	if repository == "" {
		t.Fatal("set both BETTERSTACK_LEGACY_UPTIME_DIR and BETTERSTACK_LEGACY_LOGTAIL_DIR to pinned reference clones")
	}
	source := t.TempDir()
	archive := migrationCommand(t, "", nil, "git", "-C", repository, "archive", "--format=tar", tag)
	reader := tar.NewReader(bytes.NewReader(archive))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if !filepath.IsLocal(header.Name) {
			t.Fatal("archive contains a non-local path")
		}
		path := filepath.Join(source, header.Name)
		switch header.Typeflag {
		case tar.TypeXHeader, tar.TypeXGlobalHeader:
			continue
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0700); err != nil {
				t.Fatal(err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, body, os.FileMode(header.Mode)); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatal("unexpected archive entry type")
		}
	}
	moduleFile, err := os.ReadFile(filepath.Join(source, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	moduleFields := strings.Fields(strings.SplitN(string(moduleFile), "\n", 2)[0])
	if len(moduleFields) != 2 || moduleFields[0] != "module" {
		t.Fatal("pinned legacy go.mod must declare its module first")
	}
	migrationWrite(t, filepath.Join(source, "main.go"), fmt.Sprintf(`package main
import (
 "io"
 "log"
 "os"
 %q
 "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
 "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)
func main() {
 log.SetOutput(io.Discard)
 plugin.Serve(&plugin.ServeOpts{ProviderFunc:func() *schema.Provider {
  return provider.New(provider.WithVersion(%q), provider.WithURL(os.Getenv("BETTERSTACK_MIGRATION_FIXTURE_ENDPOINT")))
 }})
}
`, moduleFields[1]+"/internal/provider", tag))
	binary := migrationProviderPath(t, mirror, "betterstackhq", name, strings.TrimPrefix(tag, "v"))
	migrationCommand(t, source, nil, "go", "build", "-o", binary, ".")
}

func migrationShowState(t *testing.T, dir string, env []string, terraform string) map[string]map[string]interface{} {
	t.Helper()
	var state struct {
		Values struct {
			Root struct {
				Resources []struct {
					Address string                 `json:"address"`
					Values  map[string]interface{} `json:"values"`
				} `json:"resources"`
			} `json:"root_module"`
		} `json:"values"`
	}
	if err := decodeNumbers(migrationCommand(t, dir, env, terraform, "show", "-json"), &state); err != nil {
		t.Fatal(err)
	}
	out := map[string]map[string]interface{}{}
	for _, resource := range state.Values.Root.Resources {
		out[resource.Address] = resource.Values
	}
	return out
}

func assertMigrationValue(t *testing.T, state map[string]map[string]interface{}, address, path string, want interface{}) {
	t.Helper()
	got := digPath(state[address], strings.Split(path, "."))
	actual, _ := json.Marshal(got)
	expected, _ := json.Marshal(want)
	if string(actual) != string(expected) {
		t.Fatalf("migration did not preserve %s.%s", address, path)
	}
}

// These routes and response projections are literal contracts from released SDK tests and the
// path API's serializers, deliberately independent of generated bindings and mover metadata.
type migrationFixtureAPI struct {
	mu            sync.Mutex
	objects       map[string]map[string]interface{}
	readOnly      bool
	refusedWrites int
}

func (api *migrationFixtureAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Authorization") != "Bearer fixture-token" {
		http.Error(w, "fixture authentication required", 401)
		return
	}
	if api.readOnly && r.Method != http.MethodGet {
		api.refusedWrites++
		http.Error(w, "remote writes forbidden during migration", 405)
		return
	}
	var key, legacyPath, targetPath, kind string
	for _, candidate := range []struct{ key, legacy, target, kind string }{
		// The unified provider reads the same path API the legacy providers wrote, at each resource's
		// newest version: only sources moved, to /api/v4.
		{"101", "/api/v2/monitor-groups", "/api/v2/monitor-groups", "monitor_group"},
		{"202", "/api/v1/source-groups", "/api/v1/source-groups", "source_group"},
		{"303", "/api/v2/datadog-integrations", "/api/v2/datadog-integrations", "datadog_integration"},
		{"404", "/api/v2/sources", "/api/v4/sources", "source"},
		{"404", "/api/v1/sources", "/api/v4/sources", "source"},
		{"505", "/api/v2/explorations", "/api/v2/explorations", "exploration"},
	} {
		if r.URL.Path == candidate.legacy || r.URL.Path == candidate.legacy+"/"+candidate.key || r.URL.Path == candidate.target+"/"+candidate.key {
			key, legacyPath, targetPath, kind = candidate.key, candidate.legacy, candidate.target, candidate.kind
			break
		}
	}
	if key == "" {
		http.Error(w, "unexpected fixture route", 404)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == legacyPath {
		var attributes map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		if err := decoder.Decode(&attributes); err != nil {
			http.Error(w, "invalid fixture JSON", 400)
			return
		}
		attributes["created_at"], attributes["updated_at"] = "2026-09-22T00:00:00Z", "2026-09-22T00:00:00Z"
		if kind == "datadog_integration" {
			attributes["webhook_url"] = "https://fixture.invalid/webhook/303"
		}
		if kind == "exploration" {
			attributes["chart"].(map[string]interface{})["name"] = attributes["name"]
			attributes["queries"].([]interface{})[0].(map[string]interface{})["id"] = json.Number("707")
			attributes["exploration_group_id"] = nil
			attributes["variables"] = []interface{}{}
		}
		if kind == "source" {
			attributes["team_id"], attributes["token"], attributes["table_name"], attributes["ingesting_host"] = json.Number("123"), "fixture-ingestion-token", "fixture_source", "in.fixture.invalid"
			if bucket, ok := attributes["custom_bucket"].(map[string]interface{}); ok {
				bucket["name"] = "fixture-bucket"
				delete(bucket, "secret_access_key")
			}
		}
		api.objects[key] = attributes
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"id": key, "type": kind, "attributes": attributes}})
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "unexpected fixture mutation", 405)
		return
	}
	stored := api.objects[key]
	if stored == nil {
		http.Error(w, "fixture not created", 404)
		return
	}
	attributes := make(map[string]interface{}, len(stored))
	for name, value := range stored {
		attributes[name] = value
	}
	// v3 is the source's newest version: its team_id is the t-prefixed Better Stack id.
	if kind == "source" && r.URL.Path == targetPath+"/"+key {
		attributes["team_id"] = "t123"
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"id": key, "type": kind, "attributes": attributes}})
}

const legacyMigrationConfig = `
terraform {
  required_providers {
    betteruptime = { source = "BetterStackHQ/better-uptime", version = "=%s" }
    logtail = { source = "BetterStackHQ/logtail", version = "=%s" }
  }
}
provider "betteruptime" { api_token = "fixture-token" }
provider "logtail" { api_token = "fixture-token" }
resource "betteruptime_monitor_group" "same" {
  name = "Fixture monitors"
  sort_index = 7
  paused = false
}
resource "logtail_source_group" "same" {
  name = "Fixture sources"
  sort_index = 8
}
resource "betteruptime_datadog_integration" "same" {
  name = "Fixture Datadog"
  policy_id = 0
  call = false
  sms = false
  email = true
  push = false
  critical_alert = false
  team_wait = 120
  recovery_period = 180
  alerting_rule = "alert"
  paused = false
}
resource "logtail_source" "secret" {
  name = "Fixture secret source"
  platform = "prometheus_scrape"
  data_region = "germany"
  source_group_id = logtail_source_group.same.id
  ingesting_paused = false
  logs_retention = 14
  metrics_retention = 60
  scrape_urls = ["https://metrics.fixture.invalid"]
  scrape_frequency_secs = 30
  scrape_request_basic_auth_user = "fixture"
  scrape_request_basic_auth_password = "fixture-scrape-secret"
  skip_ssl_verify = false
  custom_bucket {
    endpoint = "https://fixture-bucket.s3.fixture.invalid"
    access_key_id = "fixture-access-key"
    secret_access_key = "fixture-bucket-secret"
    keep_data_after_retention = true
  }
}
resource "logtail_exploration" "saved" {
  name = "Fixture exploration"
  date_range_from = "now-1h"
  date_range_to = "now"
  chart {
    chart_type = "line_chart"
    description = "Fixture query count"
    settings = jsonencode({ unit = "short", decimal_places = 0 })
  }
  query {
    name = "count"
    query_type = "sql_expression"
    sql_query = "SELECT count() AS value"
    source_variable = "source"
  }
}
`

const targetMigrationConfig = `
terraform {
  required_providers { betterstack = { source = "BetterStackHQ/betterstack", version = "=0.1.0" } }
}
provider "betterstack" {
  api_token = "fixture-token"
  team_id = "t123"
  endpoint = %q
}
moved {
  from = betteruptime_monitor_group.same
  to = betterstack_monitor_group.same
}
moved {
  from = logtail_source_group.same
  to = betterstack_source_group.same
}
moved {
  from = betteruptime_datadog_integration.same
  to = betterstack_datadog_integration.same
}
moved {
  from = logtail_source.secret
  to = betterstack_source.secret
}
moved {
  from = logtail_exploration.saved
  to = betterstack_exploration.saved
}
resource "betterstack_monitor_group" "same" {
  name = "Fixture monitors"
  sort_index = 7
  paused = false
}
resource "betterstack_source_group" "same" {
  name = "Fixture sources"
  sort_index = 8
}
resource "betterstack_datadog_integration" "same" {
  name = "Fixture Datadog"
  call = false
  sms = false
  email = true
  push = false
  critical_alert = false
  team_wait = 120
  recovery_period = 180
  alerting_rule = "alert"
  paused = false
}
resource "betterstack_source" "secret" {
  name = "Fixture secret source"
  platform = "prometheus_scrape"
  data_region = "germany"
  source_group_id = tonumber(betterstack_source_group.same.id)
  ingesting_paused = false
  logs_retention = 14
  metrics_retention = 60
  scrape_urls = ["https://metrics.fixture.invalid"]
  scrape_frequency_secs = 30
  scrape_request_basic_auth_user = "fixture"
  scrape_request_basic_auth_password = "fixture-scrape-secret"
  skip_ssl_verify = false
  custom_bucket = {
    name = "fixture-bucket"
    endpoint = "https://fixture-bucket.s3.fixture.invalid"
    access_key_id = "fixture-access-key"
    secret_access_key = "fixture-bucket-secret"
    keep_data_after_retention = true
  }
}
resource "betterstack_exploration" "saved" {
  name = "Fixture exploration"
  date_range_from = "now-1h"
  date_range_to = "now"
  chart = {
    chart_type = "line_chart"
    description = "Fixture query count"
    settings = { unit = "short", decimal_places = 0 }
  }
  queries = [{
    name = "count"
    query_type = "sql_expression"
    sql_query = "SELECT count() AS value"
    source_variable = "source"
  }]
  variables = []
}
`
