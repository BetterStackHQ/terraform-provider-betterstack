package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Real legacy-created raw-state proof for the offline helper. Neither the helper nor the
// migration preparation may reassert credentials with PATCH. All such requests fail the test.
func TestAccOfflineStateHandoff(t *testing.T) {
	if os.Getenv("BETTERSTACK_MIGRATION_FIXTURES") != "1" {
		t.Skip("set BETTERSTACK_MIGRATION_FIXTURES=1 for real Terraform state handoff proof")
	}
	terraform, err := exec.LookPath("terraform")
	if err != nil {
		t.Fatal(err)
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	mirror := filepath.Join(dir, "mirror")
	buildLegacyMigrationProvider(t, os.Getenv("BETTERSTACK_LEGACY_LOGTAIL_DIR"), "v11.2.0", "logtail", mirror)
	buildLegacyMigrationProvider(t, os.Getenv("BETTERSTACK_LEGACY_UPTIME_DIR"), "v0.22.2", "better-uptime", mirror)
	migrationCommand(t, root, nil, "go", "build", "-o", migrationProviderPath(t, mirror, "betterstackhq", "betterstack", "0.1.0"), ".")
	cliConfig := filepath.Join(dir, "terraform.rc")
	migrationWrite(t, cliConfig, fmt.Sprintf("provider_installation {\n filesystem_mirror {\n  path = %q\n }\n}\n", filepath.ToSlash(mirror)))
	api := &handoffFixtureAPI{sources: map[string]map[string]interface{}{}}
	server := httptest.NewServer(api)
	defer server.Close()
	env := []string{"TF_CLI_CONFIG_FILE=" + cliConfig, "TF_IN_AUTOMATION=1", "CHECKPOINT_DISABLE=1", "BETTERSTACK_MIGRATION_FIXTURE_ENDPOINT=" + server.URL}
	work := filepath.Join(dir, "state")
	if err := os.Mkdir(work, 0700); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(work, "main.tf")
	migrationWrite(t, config, legacyHandoffConfig)
	migrationCommand(t, work, env, terraform, "init", "-input=false", "-no-color")
	migrationCommand(t, work, env, terraform, "apply", "-auto-approve", "-input=false", "-no-color")
	backup := filepath.Join(dir, "original-pull.json")
	backupBytes := migrationCommand(t, work, env, terraform, "state", "pull")
	migrationWrite(t, backup, string(backupBytes))
	before := migrationShowState(t, work, env, terraform)
	assertMigrationValue(t, before, "logtail_collector.main", "databases.0.password", "fixture-database-secret")
	assertMigrationValue(t, before, "betteruptime_outgoing_webhook.receiver", "custom_webhook_template_attributes.0.auth_password", "fixture-webhook-secret")
	api.mu.Lock()
	api.readOnly = true
	api.mu.Unlock()
	// Forget only the guided legacy owners. Their Delete is not called.
	migrationCommand(t, work, env, terraform, "state", "rm", "logtail_source_aws_account.link", "logtail_source_gcp_project.link", "betteruptime_outgoing_webhook.receiver")
	migrationWrite(t, config, fmt.Sprintf(targetHandoffConfig, server.URL))
	migrationCommand(t, work, env, terraform, "init", "-input=false", "-no-color")
	// Refresh-only applies address moves and reads, not configured credential differences.
	migrationCommand(t, work, env, terraform, "plan", "-refresh-only", "-input=false", "-no-color", "-out=owners.plan")
	migrationCommand(t, work, env, terraform, "apply", "-input=false", "-no-color", "owners.plan")
	// CLI import reads the existing child. An ordinary import-block apply could also PATCH.
	migrationCommand(t, work, env, terraform, "import", "-input=false", "-no-color", "betterstack_collector_target.database", "9007199254740993")
	migrationCommand(t, work, env, terraform, "import", "-input=false", "-no-color", "betterstack_outgoing_webhook_integration.receiver", "507")
	imported := migrationShowState(t, work, env, terraform)
	assertMigrationValue(t, imported, "betterstack_outgoing_webhook_integration.receiver", "custom_webhook_template.auth_password", nil)
	current := filepath.Join(dir, "current-pull.json")
	currentBytes := migrationCommand(t, work, env, terraform, "state", "pull")
	migrationWrite(t, current, string(currentBytes))
	script := filepath.Join(root, "scripts", "transfer-state.py")
	for index, handoff := range []struct{ source, target, child string }{
		{"logtail_source_aws_account.link", "betterstack_source.aws", ""},
		{"logtail_source_gcp_project.link", "betterstack_source.gcp", ""},
		{"logtail_collector.main", "betterstack_collector_target.database", "9007199254740993"},
		{"betteruptime_outgoing_webhook.receiver", "betterstack_outgoing_webhook_integration.receiver", ""},
	} {
		raw, err := os.ReadFile(current)
		if err != nil {
			t.Fatal(err)
		}
		var snapshot struct {
			Serial int64 `json:"serial"`
		}
		if err := json.Unmarshal(raw, &snapshot); err != nil {
			t.Fatal(err)
		}
		candidate := filepath.Join(dir, fmt.Sprintf("candidate-%d.json", index))
		args := []string{script, "--backup", backup, "--current", current, "--source", handoff.source, "--target", handoff.target, "--expected-serial", strconv.FormatInt(snapshot.Serial, 10), "--output", candidate}
		if handoff.child != "" {
			args = append(args, "--child-id", handoff.child)
		}
		output := migrationCommand(t, work, nil, python, args...)
		if strings.Contains(string(output), "fixture-aws-secret") || strings.Contains(string(output), "fixture-database-secret") || strings.Contains(string(output), "fixture-webhook-secret") {
			t.Fatal("offline helper printed secrets")
		}
		if info, err := os.Stat(candidate); err != nil || info.Mode().Perm() != 0600 {
			t.Fatal("candidate is not private")
		}
		current = candidate
	}
	unchanged, err := os.ReadFile(backup)
	if err != nil || string(unchanged) != string(backupBytes) {
		t.Fatal("original state backup changed")
	}
	unchanged, err = os.ReadFile(filepath.Join(dir, "current-pull.json"))
	if err != nil || string(unchanged) != string(currentBytes) {
		t.Fatal("current input state changed")
	}
	// The operator performs this explicit supported backend installation; the helper never does.
	migrationCommand(t, work, env, terraform, "state", "push", current)
	migrationCommand(t, work, env, terraform, "plan", "-detailed-exitcode", "-input=false", "-no-color")
	after := migrationShowState(t, work, env, terraform)
	assertMigrationValue(t, after, "betterstack_source.aws", "aws_external_id", "fixture-aws-secret")
	assertMigrationValue(t, after, "betterstack_source.aws", "aws_auto_subscribe_log_groups", false)
	assertMigrationValue(t, after, "betterstack_source.gcp", "gcp_project_number", "9007199254740993")
	assertMigrationValue(t, after, "betterstack_collector_target.database", "password", "fixture-database-secret")
	assertMigrationValue(t, after, "betterstack_collector_target.database", "id", "9007199254740993")
	assertMigrationValue(t, after, "betterstack_collector.main", "secret", "fixture-enrollment-secret")
	assertMigrationValue(t, after, "betterstack_outgoing_webhook_integration.receiver", "custom_webhook_template.auth_password", "fixture-webhook-secret")
	assertMigrationValue(t, after, "betterstack_outgoing_webhook_integration.receiver", "id", before["betteruptime_outgoing_webhook.receiver"]["id"])
	for _, address := range []string{"betterstack_source.aws", "betterstack_source.gcp", "betterstack_collector.main"} {
		assertMigrationValue(t, after, address, "data_region", "eu-nbg-2")
	}
	api.mu.Lock()
	defer api.mu.Unlock()
	if api.refusedWrites != 0 {
		t.Fatalf("migration attempted %d remote writes", api.refusedWrites)
	}
}

type handoffFixtureAPI struct {
	mu            sync.Mutex
	sources       map[string]map[string]interface{}
	collector     map[string]interface{}
	database      map[string]interface{}
	webhook       map[string]interface{}
	readOnly      bool
	refusedWrites int
}

func (api *handoffFixtureAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Authorization") != "Bearer fixture-token" {
		http.Error(w, "fixture authentication required", 401)
		return
	}
	if api.readOnly && r.Method != http.MethodGet {
		api.refusedWrites++
		http.Error(w, "remote writes forbidden during handoff", 405)
		return
	}
	respond := func(id, kind string, attributes map[string]interface{}) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"id": id, "type": kind, "attributes": attributes}})
	}
	readBody := func() map[string]interface{} {
		var body map[string]interface{}
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		if err := decoder.Decode(&body); err != nil {
			http.Error(w, "invalid fixture request", 400)
			return nil
		}
		return body
	}
	if r.Method == http.MethodPost && r.URL.Path == "/api/v2/sources" {
		body := readBody()
		if body == nil {
			return
		}
		id := "501"
		if body["platform"] == "gcp" {
			id = "502"
		}
		body["token"], body["table_name"], body["ingesting_host"], body["team_id"] = "fixture-token-"+id, "fixture_"+id, "in.fixture.invalid", json.Number("123")
		// Both published serializers resolve an omitted region to the source's real storage backend.
		if region, _ := body["data_region"].(string); region == "" {
			body["data_region"] = "eu-nbg-2"
		}
		body["created_at"], body["updated_at"] = "2026-09-22T00:00:00Z", "2026-09-22T00:00:00Z"
		api.sources[id] = body
		w.WriteHeader(http.StatusCreated)
		respond(id, "source", body)
		return
	}
	if r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/v2/sources/") {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/sources/")
		body := readBody()
		if body == nil {
			return
		}
		if api.sources[id] == nil {
			http.Error(w, "unknown source", 404)
			return
		}
		// Link credentials are never returned; only the effective subscription default is readable.
		if value, exists := body["aws_auto_sub_log_groups"]; exists {
			api.sources[id]["aws_auto_sub_log_groups"] = value
		}
		respond(id, "source", api.sources[id])
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/api/v1/collectors" {
		body := readBody()
		if body == nil {
			return
		}
		rows, ok := body["databases"].([]interface{})
		if !ok || len(rows) != 1 {
			http.Error(w, "expected one legacy database", 400)
			return
		}
		database, ok := rows[0].(map[string]interface{})
		if !ok {
			http.Error(w, "invalid database", 400)
			return
		}
		delete(database, "password")
		database["id"] = json.Number("9007199254740993")
		api.database = database
		delete(body, "databases")
		body["secret"], body["status"], body["team_id"], body["source_id"] = "fixture-enrollment-secret", "active", json.Number("123"), json.Number("505")
		body["hosts_count"], body["hosts_up_count"], body["databases_count"] = json.Number("0"), json.Number("0"), json.Number("1")
		if region, _ := body["data_region"].(string); region == "" {
			body["data_region"] = "eu-nbg-2"
		}
		body["created_at"], body["updated_at"] = "2026-09-22T00:00:00Z", "2026-09-22T00:00:00Z"
		api.collector = body
		w.WriteHeader(http.StatusCreated)
		respond("503", "collector", body)
		return
	}
	if r.Method == http.MethodPost && r.URL.Path == "/api/v2/outgoing-webhooks" {
		body := readBody()
		if body == nil {
			return
		}
		body["custom_webhook_template_attributes"].(map[string]interface{})["id"] = "508"
		body["created_at"], body["updated_at"] = "2026-09-22T00:00:00Z", "2026-09-22T00:00:00Z"
		api.webhook = body
		w.WriteHeader(http.StatusCreated)
		respond("507", "outgoing_webhook", body)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "unexpected fixture write", 405)
		return
	}
	if r.URL.Path == "/api/v1/collectors/503/databases" {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{map[string]interface{}{"id": json.Number("9007199254740993"), "attributes": api.database}}})
		return
	}
	if r.URL.Path == "/api/v1/collectors/503" {
		respond("503", "collector", api.collector)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/v2/sources/") {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/sources/")
		respond(id, "source", api.sources[id])
		return
	}
	if r.URL.Path == "/api/v2/outgoing-webhooks/507" {
		respond("507", "outgoing_webhook", api.webhook)
		return
	}
	if r.Header.Get("Better-Stack-Version") == "" {
		http.Error(w, "version header missing", 400)
		return
	}
	copyMap := func(source map[string]interface{}) map[string]interface{} {
		out := map[string]interface{}{}
		for key, value := range source {
			out[key] = value
		}
		return out
	}
	if r.URL.Path == "/api/outgoing-webhooks/507" {
		body := copyMap(api.webhook)
		template := copyMap(body["custom_webhook_template_attributes"].(map[string]interface{}))
		delete(template, "id")
		delete(template, "auth_password")
		delete(body, "custom_webhook_template_attributes")
		delete(body, "team_name")
		body["custom_webhook_template"], body["team_id"] = template, "t123"
		respond("507", "outgoing_webhook", body)
		return
	}
	database := copyMap(api.database)
	database["id"], database["collector_id"], database["kind"], database["enabled"], database["status"] = "9007199254740993", "503", database["service_type"], true, "connected"
	delete(database, "service_type")
	if r.URL.Path == "/api/collector-targets/9007199254740993" {
		respond("9007199254740993", "collector_target", database)
		return
	}
	if r.URL.Path == "/api/collectors/503" {
		body := copyMap(api.collector)
		delete(body, "secret")
		delete(body, "team_name")
		body["team_id"], body["source_id"], body["databases"] = "t123", "505", []interface{}{database}
		for from, to := range map[string]string{"logs_retention": "logs_retention_days", "metrics_retention": "metrics_retention_days"} {
			if value, exists := body[from]; exists {
				body[to] = value
				delete(body, from)
			}
		}
		respond("503", "collector", body)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/sources/") {
		id := strings.TrimPrefix(r.URL.Path, "/api/sources/")
		body := copyMap(api.sources[id])
		delete(body, "team_name")
		body["team_id"] = "t123"
		for from, to := range map[string]string{"logs_retention": "logs_retention_days", "metrics_retention": "metrics_retention_days", "aws_auto_sub_log_groups": "aws_auto_subscribe_log_groups"} {
			if value, exists := body[from]; exists {
				body[to] = value
				delete(body, from)
			}
		}
		respond(id, "source", body)
		return
	}
	http.Error(w, "unexpected fixture route", 404)
}

const legacyHandoffConfig = `
terraform {
  required_providers {
    logtail = { source = "BetterStackHQ/logtail", version = "=11.2.0" }
    betteruptime = { source = "BetterStackHQ/better-uptime", version = "=0.22.2" }
  }
}
provider "logtail" { api_token = "fixture-token" }
provider "betteruptime" { api_token = "fixture-token" }
resource "logtail_source" "aws" {
  name = "Fixture AWS"
  platform = "aws"
  logs_retention = 30
  metrics_retention = 30
  ingesting_paused = false
}
resource "logtail_source_aws_account" "link" {
  source_id = logtail_source.aws.id
  aws_role_arn = "arn:aws:iam::123456789012:role/fixture"
  aws_external_id = "fixture-aws-secret"
  auto_subscribe_log_groups = false
}
resource "logtail_source" "gcp" {
  name = "Fixture GCP"
  platform = "gcp"
  logs_retention = 30
  metrics_retention = 30
  ingesting_paused = false
}
resource "logtail_source_gcp_project" "link" {
  source_id = logtail_source.gcp.id
  gcp_project_id = "fixture-project"
  gcp_project_number = "9007199254740993"
}
resource "logtail_collector" "main" {
  name = "Fixture collector"
  platform = "docker"
  ingesting_paused = false
  logs_retention = 30
  metrics_retention = 30
  databases {
    service_type = "postgres"
    host = "db.fixture.invalid"
    port = 5432
    username = "fixture"
    password = "fixture-database-secret"
    ssl_mode = "require"
  }
}
resource "betteruptime_outgoing_webhook" "receiver" {
  name = "Fixture webhook"
  url = "https://receiver.fixture.invalid"
  trigger_type = "incident_change"
  notify_alongside_primary_responder = false
  on_incident_started = true
  on_incident_acknowledged = false
  on_incident_resolved = false
  on_incident_reopened = false
  on_incident_comment = false
  custom_webhook_template_attributes {
    http_method = "post"
    auth_username = "fixture"
    auth_password = "fixture-webhook-secret"
    body_template = "{\"kind\":\"fixture\"}"
    headers_template {
      name = "X-Fixture"
      value = "preserved"
    }
  }
}
`

const targetHandoffConfig = `
terraform {
  required_providers { betterstack = { source = "BetterStackHQ/betterstack", version = "=0.1.0" } }
}
provider "betterstack" {
  api_token = "fixture-token"
  team_id = "t123"
  endpoint = %q
}
moved {
  from = logtail_source.aws
  to = betterstack_source.aws
}
moved {
  from = logtail_source.gcp
  to = betterstack_source.gcp
}
moved {
  from = logtail_collector.main
  to = betterstack_collector.main
}
resource "betterstack_source" "aws" {
  name = "Fixture AWS"
  platform = "aws"
  logs_retention_days = 30
  metrics_retention_days = 30
  ingesting_paused = false
  # Retain old SDK empty strings rather than changing these nullable fields to null.
  code_mapping_stack_root = ""
  code_mapping_source_root = ""
  aws_role_arn = "arn:aws:iam::123456789012:role/fixture"
  aws_external_id = "fixture-aws-secret"
  aws_auto_subscribe_log_groups = false
}
resource "betterstack_source" "gcp" {
  name = "Fixture GCP"
  platform = "gcp"
  logs_retention_days = 30
  metrics_retention_days = 30
  ingesting_paused = false
  code_mapping_stack_root = ""
  code_mapping_source_root = ""
  gcp_project_id = "fixture-project"
  gcp_project_number = "9007199254740993"
}
resource "betterstack_collector" "main" {
  name = "Fixture collector"
  platform = "docker"
  ingesting_paused = false
  note = ""
  user_vector_config = ""
}
resource "betterstack_collector_target" "database" {
  collector_id = betterstack_collector.main.id
  kind = "postgres"
  enabled = true
  host = "db.fixture.invalid"
  port = 5432
  username = "fixture"
  password = "fixture-database-secret"
  ssl_mode = "require"
}
resource "betterstack_outgoing_webhook_integration" "receiver" {
  name = "Fixture webhook"
  url = "https://receiver.fixture.invalid"
  trigger_type = "incident_change"
  notify_alongside_primary_responder = false
  on_incident_started = true
  on_incident_acknowledged = false
  on_incident_resolved = false
  on_incident_reopened = false
  on_incident_comment = false
  custom_webhook_template = {
    http_method = "post"
    auth_username = "fixture"
    auth_password = "fixture-webhook-secret"
    body_template = "{\"kind\":\"fixture\"}"
    headers_template = [{ name = "X-Fixture", value = "preserved" }]
  }
}
`
