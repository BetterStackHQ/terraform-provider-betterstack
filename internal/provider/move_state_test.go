package provider

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func moveResponse(s schema.Schema) *resource.MoveStateResponse {
	return &resource.MoveStateResponse{TargetState: tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}}
}

func applyLegacyMove(s schema.Schema, metadata, state, address string, version int64) *resource.MoveStateResponse {
	resp := moveResponse(s)
	legacyStateMover(metadata).StateMover(context.Background(), resource.MoveStateRequest{
		SourceTypeName: "logtail_source", SourceProviderAddress: address, SourceSchemaVersion: version,
		SourceRawState: &tfprotov6.RawState{JSON: []byte(state)},
	}, resp)
	return resp
}

// These are converter regression inputs, not a claim of Terraform-created fixture state.
// TestAccLegacyStateMigration below exercises released SDK providers and real Terraform state.
const conversionMetadata = `{
  "source_provider":"registry.terraform.io/betterstackhq/logtail", "source_type":"logtail_source",
  "kind":"resource", "handling":"move", "identity":"preserve", "schema_versions":[0],
  "renames":{"source_group_id":"group_id"}, "constants":{"service":"logs"},
  "transforms":[
    {"op":"string_id","path":"group_id"},
    {"op":"singleton","path":"custom_bucket"},
    {"op":"json_decode","path":"settings"},
    {"op":"seconds","path":"period","factor":60},
    {"op":"sentinel_null","path":"recovery_period","value":-1},
    {"op":"rename","path":"headers[].old_name","to":"headers[].name"}
  ],
  "ignored_fields":["team_name"],
  "source_attributes":[
    {"name":"id","type":"string"}, {"name":"source_group_id","type":"int","required":true},
    {"name":"custom_bucket","type":"list(object)","max_items":1},
    {"name":"custom_bucket.endpoint","type":"string","required":true},
    {"name":"custom_bucket.secret_access_key","type":"string","required":true,"sensitive":true},
    {"name":"settings","type":"string"}, {"name":"period","type":"int"},
    {"name":"recovery_period","type":"int"}, {"name":"team_name","type":"string"},
    {"name":"headers","type":"list(object)"}, {"name":"headers.old_name","type":"string"},
    {"name":"service","type":"string"}
  ]
}`

const conversionState = `{
  "id":"123", "source_group_id":9007199254740993, "team_name":"Example",
  "custom_bucket":[{"endpoint":"https://bucket.invalid","secret_access_key":"state-secret"}],
  "settings":"{\"enabled\":true,\"labels\":[\"a\",3],\"nested\":{\"password\":\"nested-secret\"}}",
  "period":5, "recovery_period":-1, "headers":[{"old_name":"X-Test"}]
}`

func conversionSchema() schema.Schema {
	return schema.Schema{Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true},
		"group_id": schema.StringAttribute{Required: true},
		"service":  schema.StringAttribute{Required: true},
		"custom_bucket": schema.SingleNestedAttribute{Optional: true, Attributes: map[string]schema.Attribute{
			"endpoint":          schema.StringAttribute{Required: true},
			"secret_access_key": schema.StringAttribute{Required: true, Sensitive: true},
		}},
		"settings":        schema.DynamicAttribute{Optional: true},
		"period":          schema.Int64Attribute{Optional: true},
		"recovery_period": schema.Int64Attribute{Optional: true},
		"headers": schema.ListNestedAttribute{Optional: true, NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{Required: true},
		}}},
	}}
}

func TestLegacyMovePreservesValues(t *testing.T) {
	resp := applyLegacyMove(conversionSchema(), conversionMetadata, conversionState, "mirror.example/betterstackhq/logtail", 0)
	if resp.Diagnostics.HasError() {
		t.Fatalf("move: %v", resp.Diagnostics)
	}
	got, err := tftypesToAPI(resp.TargetState.Raw)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"id":"123","group_id":"9007199254740993","service":"logs","custom_bucket":{"endpoint":"https://bucket.invalid","secret_access_key":"state-secret"},"settings":{"enabled":true,"labels":["a",3],"nested":{"password":"nested-secret"}},"period":300,"recovery_period":null,"headers":[{"name":"X-Test"}]}`
	var expected interface{}
	if err := decodeNumbers([]byte(want), &expected); err != nil {
		t.Fatal(err)
	}
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(expected)
	if string(gotJSON) != string(wantJSON) {
		t.Fatal("moved state did not preserve the expected identity, secrets, nested values and timing")
	}
	second := applyLegacyMove(conversionSchema(), conversionMetadata, conversionState, "registry.terraform.io/betterstackhq/logtail", 0)
	if !resp.TargetState.Raw.Equal(second.TargetState.Raw) {
		t.Fatal("conversion depends on registry host or previous calls")
	}
}

func TestLegacyMoveProjectsComputedDynamicFieldsBeforeRefresh(t *testing.T) {
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true},
		"queries":        schema.DynamicAttribute{Required: true, Sensitive: true},
		"queries_result": schema.DynamicAttribute{Computed: true, Sensitive: true},
	}}
	metadata := `{
		"source_provider":"registry.terraform.io/betterstackhq/logtail","source_type":"logtail_source",
		"kind":"resource","handling":"move","identity":"preserve","schema_versions":[0],
		"transforms":[{"op":"string_id","path":"queries[].id"}],
		"response_projections":[{"source":"queries","target":"queries_result","readonly_fields":["queries[].id"]}],
		"source_attributes":[{"name":"id","type":"string"},{"name":"queries","type":"list(object)"},
			{"name":"queries.id","type":"int"},{"name":"queries.sql_query","type":"string"},{"name":"queries.secret","type":"string","sensitive":true}]
	}`
	state := `{"id":"123","queries":[{"id":707,"sql_query":"SELECT 1","secret":"retained-secret"}]}`
	resp := applyLegacyMove(s, metadata, state, "registry.terraform.io/betterstackhq/logtail", 0)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	value, err := tftypesToAPI(resp.TargetState.Raw)
	if err != nil {
		t.Fatal(err)
	}
	actual, _ := json.Marshal(value)
	expected := `{"id":"123","queries":[{"secret":"retained-secret","sql_query":"SELECT 1"}],"queries_result":[{"id":"707","secret":"retained-secret","sql_query":"SELECT 1"}]}`
	if string(actual) != expected {
		t.Fatal("move did not preserve complete computed results separately from configured Dynamic inputs")
	}
	unsafe := strings.Replace(metadata, `"readonly_fields":["queries[].id"]`, `"readonly_fields":["id"]`, 1)
	refused := applyLegacyMove(s, unsafe, state, "registry.terraform.io/betterstackhq/logtail", 0)
	if !refused.Diagnostics.HasError() || !refused.TargetState.Raw.IsNull() {
		t.Fatal("a field projection was allowed to remove an unrelated object identity")
	}
}

func TestSentinelOmissionRetainsEveryNonmatchingValue(t *testing.T) {
	state := map[string]interface{}{
		"empty": "", "nonempty": "SELECT secret", "whitespace": " ", "null": nil, "number": json.Number("0"),
	}
	for key := range state {
		if err := applyMigrationTransform(state, migrationTransform{Op: "sentinel_omit", Path: key, Value: ""}); err != nil {
			t.Fatal(err)
		}
	}
	if _, exists := state["empty"]; exists {
		t.Fatal("matching empty sentinel survived")
	}
	if state["nonempty"] != "SELECT secret" || state["whitespace"] != " " || state["number"] != json.Number("0") {
		t.Fatal("sentinel omission changed a nonmatching value")
	}
	if value, exists := state["null"]; !exists || value != nil {
		t.Fatal("empty-string omission removed null")
	}
	if err := applyMigrationTransform(state, migrationTransform{Op: "sentinel_omit", Path: "null", Value: nil}); err != nil {
		t.Fatal(err)
	}
	if _, exists := state["null"]; exists {
		t.Fatal("explicit null sentinel survived")
	}
	array := map[string]interface{}{"rows": []interface{}{""}}
	if err := applyMigrationTransform(array, migrationTransform{Op: "sentinel_omit", Path: "rows[]", Value: ""}); err == nil {
		t.Fatal("field omission accepted a cardinality-changing array-element path")
	}
	if array["rows"].([]interface{})[0] != "" {
		t.Fatal("refused array omission changed state")
	}
}

func TestLegacyMoveRejectsLossyConversions(t *testing.T) {
	cases := []struct {
		name, metadata, state string
		version               int64
	}{
		{"future schema", conversionMetadata, conversionState, 1},
		{"absent schema catalog", strings.Replace(conversionMetadata, "source_attributes", "unrecognized_attributes", 1), conversionState, 0},
		{"missing identity", conversionMetadata, strings.Replace(conversionState, `"id":"123"`, `"id":""`, 1), 0},
		{"missing parent", conversionMetadata, strings.Replace(conversionState, `"source_group_id":9007199254740993`, `"source_group_id":null`, 1), 0},
		{"missing secret", conversionMetadata, strings.Replace(conversionState, `"secret_access_key":"state-secret"`, `"secret_access_key":null`, 1), 0},
		{"missing computed-only secret", strings.Replace(conversionMetadata, `"required":true,"sensitive":true`, `"computed":true,"sensitive":true`, 1), strings.Replace(conversionState, `"secret_access_key":"state-secret"`, `"secret_access_key":null`, 1), 0},
		{"unknown source attribute", conversionMetadata, strings.Replace(conversionState, `"id":"123"`, `"id":"123","future":"never-log-this-secret"`, 1), 0},
		{"unknown nested field", conversionMetadata, strings.Replace(conversionState, `"endpoint":"https://bucket.invalid"`, `"endpoint":"https://bucket.invalid","future":"never-log-this-secret"`, 1), 0},
		{"incompatible nested field", conversionMetadata, strings.Replace(conversionState, `"endpoint":"https://bucket.invalid"`, `"endpoint":{"secret":"never-log-this-secret"}`, 1), 0},
		{"multiple singleton blocks", conversionMetadata, strings.Replace(conversionState, `"custom_bucket":[`, `"custom_bucket":[{},`, 1), 0},
		{"wrong list shape", conversionMetadata, strings.Replace(conversionState, `"headers":[{"old_name":"X-Test"}]`, `"headers":{"old_name":"X-Test"}`, 1), 0},
		{"no implicit singleton", strings.Replace(conversionMetadata, `{"op":"singleton","path":"custom_bucket"},`, "", 1), conversionState, 0},
		{"no implicit ID coercion", strings.Replace(conversionMetadata, `{"op":"string_id","path":"group_id"},`, "", 1), conversionState, 0},
		{"malformed embedded JSON", conversionMetadata, strings.Replace(conversionState, `"settings":"{`, `"settings":"broken{`, 1), 0},
		{"constant conflict", conversionMetadata, strings.Replace(conversionState, `"id":"123"`, `"id":"123","service":"other"`, 1), 0},
		{"unexpected target field", strings.Replace(conversionMetadata, `"group_id"`, `"missing_field"`, -1), conversionState, 0},
		{"invalid JSON", conversionMetadata, `{`, 0},
		{"trailing JSON", conversionMetadata, conversionState + `{}`, 0},
		{"duplicate keys", conversionMetadata, strings.Replace(conversionState, `"id":"123"`, `"id":"123","id":"456"`, 1), 0},
		{"target integer overflow", conversionMetadata, strings.Replace(conversionState, `"period":5`, `"period":9223372036854775807`, 1), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := applyLegacyMove(conversionSchema(), tc.metadata, tc.state, "registry.terraform.io/betterstackhq/logtail", tc.version)
			if !resp.Diagnostics.HasError() || !resp.TargetState.Raw.IsNull() {
				t.Fatal("unsafe state was accepted or partially written")
			}
			for _, d := range resp.Diagnostics {
				if strings.Contains(d.Detail(), "never-log-this-secret") || strings.Contains(d.Detail(), "state-secret") || strings.Contains(d.Detail(), "nested-secret") {
					t.Fatal("state secret leaked into a diagnostic")
				}
			}
		})
	}
}

func TestLegacyMoveRejectsForeignProviders(t *testing.T) {
	for _, address := range []string{"", "logtail", "betterstackhq/logtail", "registry.terraform.io/other/logtail", "registry.terraform.io/betterstackhq/logtail-extra", "registry.terraform.io/evilbetterstackhq/logtail", "registry.terraform.io/betterstacKhq/logtail", "registry.terraform.io/betterstackhq/logtail/extra", "https://registry.terraform.io/betterstackhq/logtail", "/betterstackhq/logtail"} {
		resp := applyLegacyMove(conversionSchema(), conversionMetadata, conversionState, address, 0)
		if !resp.TargetState.Raw.IsNull() || resp.Diagnostics.HasError() {
			t.Fatalf("foreign address %q was adopted", address)
		}
	}
	for _, metadata := range []string{strings.Replace(conversionMetadata, `"handling":"move"`, `"handling":"guided"`, 1), strings.Replace(conversionMetadata, `"kind":"resource"`, `"kind":"data_source"`, 1), strings.Replace(conversionMetadata, `"identity":"preserve"`, `"identity":"lookup"`, 1), strings.Replace(conversionMetadata, `"source_type":"logtail_source"`, `"source_type":"logtail_source_other"`, 1)} {
		resp := applyLegacyMove(conversionSchema(), metadata, conversionState, "registry.terraform.io/betterstackhq/logtail", 0)
		if !resp.TargetState.Raw.IsNull() || resp.Diagnostics.HasError() {
			t.Fatal("guided, lookup, data-source or foreign type mapping claimed mover support")
		}
	}
}

func TestLegacyMoveRejectsRenameCollisions(t *testing.T) {
	root := map[string]interface{}{"old": "first-secret", "new": "second-secret"}
	if err := renameMigrationPath(root, "old", "new", ""); err == nil {
		t.Fatal("rename silently overwrote a value")
	}
	if err := applyMigrationTransform(root, migrationTransform{Op: "rename", Path: "old", To: "new", OnConflict: "target"}); err != nil {
		t.Fatal(err)
	}
	if root["new"] != "second-secret" {
		t.Fatal("declared legacy precedence changed")
	}
	if _, exists := root["old"]; exists {
		t.Fatal("obsolete alias survived explicit precedence conversion")
	}
}

func TestLegacyMoveSingletonAndPrecision(t *testing.T) {
	for _, tc := range []struct {
		name, state string
		null        bool
	}{
		{"absent optional block", strings.Replace(conversionState, `[{"endpoint":"https://bucket.invalid","secret_access_key":"state-secret"}]`, `[]`, 1), true},
		{"present block", conversionState, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := applyLegacyMove(conversionSchema(), conversionMetadata, tc.state, "registry.terraform.io/betterstackhq/logtail", 0)
			if resp.Diagnostics.HasError() {
				t.Fatalf("move: %v", resp.Diagnostics)
			}
			var attrs map[string]tftypes.Value
			if err := resp.TargetState.Raw.As(&attrs); err != nil {
				t.Fatal(err)
			}
			if attrs["custom_bucket"].IsNull() != tc.null {
				t.Fatal("optional singleton presence changed")
			}
		})
	}
	// IDs must not be rounded, even when the source represents an integer with exponent notation.
	state := strings.Replace(conversionState, "9007199254740993", "9007199254740993e0", 1)
	resp := applyLegacyMove(conversionSchema(), conversionMetadata, state, "registry.terraform.io/betterstackhq/logtail", 0)
	if resp.Diagnostics.HasError() {
		t.Fatalf("move: %v", resp.Diagnostics)
	}
	var attrs map[string]tftypes.Value
	if err := resp.TargetState.Raw.As(&attrs); err != nil {
		t.Fatal(err)
	}
	var id string
	if err := attrs["group_id"].As(&id); err != nil || id != "9007199254740993" {
		t.Fatal("integer ID lost precision")
	}
}

func TestLegacyMoveRejectsTargetShapeMismatch(t *testing.T) {
	// A declared source map is not permission to silently discard target object members.
	s := schema.Schema{Attributes: map[string]schema.Attribute{"settings": schema.MapAttribute{Optional: true, ElementType: types.StringType}}}
	value := map[string]interface{}{"settings": map[string]interface{}{"secret": map[string]interface{}{"nested": true}}}
	if _, err := strictMigrationValue(s.Type().TerraformType(context.Background()), value, "state"); err == nil {
		t.Fatal("object coerced to string")
	}
}

func TestLegacyConditionalTimingAndHistoricalAlias(t *testing.T) {
	var timing migrationTransform
	if err := decodeNumbers([]byte(`{"op":"seconds","path":"request_timeout_seconds","factor":0.001,"when":{"path":"monitor_type","in":["ping","tcp","udp","smtp","pop","imap","dns"]}}`), &timing); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ kind, want string }{{"ping", "0.5"}, {"status", "500"}} {
		state := map[string]interface{}{"monitor_type": tc.kind, "request_timeout_seconds": json.Number("500")}
		if err := applyMigrationTransform(state, timing); err != nil {
			t.Fatal(err)
		}
		if state["request_timeout_seconds"] != json.Number(tc.want) {
			t.Fatal("monitor timing changed effective units")
		}
	}
	for _, tc := range []struct {
		old, current, want string
		conflict           bool
	}{
		{".foo = 1", "", ".foo = 1", false},
		{"", ".foo = 2", ".foo = 2", false},
		{".foo = 1", ".foo = 2", "", true},
	} {
		state := map[string]interface{}{"vrl_transformation": tc.old, "vrl_transformation_logs": tc.current}
		for _, path := range []string{"vrl_transformation", "vrl_transformation_logs"} {
			if err := applyMigrationTransform(state, migrationTransform{Op: "sentinel_null", Path: path, Value: ""}); err != nil {
				t.Fatal(err)
			}
		}
		err := applyMigrationTransform(state, migrationTransform{Op: "rename", Path: "vrl_transformation", To: "vrl_transformation_logs"})
		if tc.conflict {
			if err == nil {
				t.Fatal("conflicting historical programs were merged")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if state["vrl_transformation_logs"] != tc.want {
			t.Fatal("historical program was lost")
		}
	}
}

func TestLegacyIgnoredFieldsAreExplicitAndNested(t *testing.T) {
	state := map[string]interface{}{"headers": []interface{}{map[string]interface{}{"id": "legacy-row", "name": "Authorization", "value": "fixture-secret"}}}
	err := editMigrationPath(state, []string{"headers[]", "id"}, false, func(parent map[string]interface{}, key string) error { delete(parent, key); return nil })
	if err != nil {
		t.Fatal(err)
	}
	header := state["headers"].([]interface{})[0].(map[string]interface{})
	if _, exists := header["id"]; exists {
		t.Fatal("declared representation field survived")
	}
	if header["name"] != "Authorization" || header["value"] != "fixture-secret" {
		t.Fatal("header semantics or secret changed")
	}
}

func TestLegacyWrapTurnsEachValueIntoAnObjectInOrder(t *testing.T) {
	transform := migrationTransform{Op: "wrap", Path: "routes[].channel_ids", To: "routes[].channels", Key: "id"}
	state := map[string]interface{}{"routes": []interface{}{
		map[string]interface{}{"channel_ids": []interface{}{"b", "a"}, "applications": []interface{}{}},
		map[string]interface{}{"channel_ids": nil},
	}}
	if err := applyMigrationTransform(state, transform); err != nil {
		t.Fatal(err)
	}
	first := state["routes"].([]interface{})[0].(map[string]interface{})
	channels := first["channels"].([]interface{})
	if len(channels) != 2 || channels[0].(map[string]interface{})["id"] != "b" || channels[1].(map[string]interface{})["id"] != "a" {
		t.Fatal("wrapped values lost or reordered")
	}
	if _, exists := first["channel_ids"]; exists {
		t.Fatal("source field survived wrap")
	}
	for _, rows := range []interface{}{
		[]interface{}{map[string]interface{}{"id": "x"}},
		[]interface{}{[]interface{}{"x"}},
		[]interface{}{nil},
		"not a list",
	} {
		if err := applyMigrationTransform(map[string]interface{}{"routes": []interface{}{map[string]interface{}{"channel_ids": rows}}}, transform); err == nil {
			t.Fatal("a value that is not plain was wrapped")
		}
	}
}

func TestLegacyIndexByPreservesValuesAndRejectsAmbiguousKeys(t *testing.T) {
	transform := migrationTransform{Op: "index_by", Path: "service_option", To: "services_options", Key: "name"}
	state := map[string]interface{}{"service_option": []interface{}{
		map[string]interface{}{"name": "api", "logs": true, "nested": map[string]interface{}{"secret": "fixture-secret"}},
		map[string]interface{}{"name": "worker", "logs": false},
	}}
	if err := applyMigrationTransform(state, transform); err != nil {
		t.Fatal(err)
	}
	index := state["services_options"].(map[string]interface{})
	if index["api"].(map[string]interface{})["logs"] != true || index["worker"].(map[string]interface{})["logs"] != false || index["api"].(map[string]interface{})["nested"].(map[string]interface{})["secret"] != "fixture-secret" {
		t.Fatal("indexed options lost values")
	}
	if _, exists := index["api"].(map[string]interface{})["name"]; exists {
		t.Fatal("index key was duplicated into the value")
	}
	if _, exists := state["service_option"]; exists {
		t.Fatal("source field survived index conversion")
	}
	for _, rows := range []interface{}{
		[]interface{}{map[string]interface{}{"name": "same"}, map[string]interface{}{"name": "same"}},
		[]interface{}{map[string]interface{}{"enabled": true}},
		[]interface{}{map[string]interface{}{"name": json.Number("1")}},
		[]interface{}{map[string]interface{}{"name": ""}},
		[]interface{}{"not an object"},
		map[string]interface{}{"name": "not a list"},
	} {
		if err := applyMigrationTransform(map[string]interface{}{"service_option": rows}, transform); err == nil {
			t.Fatal("ambiguous indexed options accepted")
		}
	}
}

func TestLegacyMigrationManifestMatchesRegisteredMovers(t *testing.T) {
	raw, err := os.ReadFile("../../migration-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	type mapping struct {
		SourceProvider string `json:"source_provider"`
		SourceType     string `json:"source_type"`
		TargetType     string `json:"target_type"`
		Handling       string `json:"handling"`
		Kind           string `json:"kind"`
	}
	var manifest struct {
		Movers   []mapping `json:"emitted_movers"`
		Mappings []mapping `json:"mappings"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	registered := map[string]resource.Resource{}
	for _, factory := range New("test")().Resources(ctx) {
		r := factory()
		var metadata resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "betterstack"}, &metadata)
		registered[metadata.TypeName] = r
	}
	run := func(t *testing.T, m mapping) *resource.MoveStateResponse {
		t.Helper()
		r, exists := registered[m.TargetType]
		if !exists {
			t.Fatalf("manifest target %s is not registered", m.TargetType)
		}
		var s resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &s)
		resp := moveResponse(s.Schema)
		if mover, ok := r.(resource.ResourceWithMoveState); ok {
			for _, handler := range mover.MoveState(ctx) {
				handler.StateMover(ctx, resource.MoveStateRequest{SourceTypeName: m.SourceType, SourceProviderAddress: m.SourceProvider, SourceSchemaVersion: 999999, SourceRawState: &tfprotov6.RawState{JSON: []byte(`{"id":"123"}`)}}, resp)
			}
		}
		return resp
	}
	for _, m := range manifest.Movers {
		t.Run(m.SourceType, func(t *testing.T) {
			resp := run(t, m)
			if !resp.Diagnostics.HasError() || !resp.TargetState.Raw.IsNull() {
				t.Fatal("declared mover missing from built registration, or accepted an unknown source schema")
			}
		})
	}
	for _, m := range manifest.Mappings {
		if m.Kind != "resource" || m.Handling != "guided" {
			continue
		}
		if _, exists := registered[m.TargetType]; !exists {
			continue
		}
		resp := run(t, m)
		if resp.Diagnostics.HasError() || !resp.TargetState.Raw.IsNull() {
			t.Fatalf("guided source %s claims state-mover support", m.SourceType)
		}
	}
}

// These cases are shared with `betterstack terraform migrate`, which has its own port of this
// mover in the CLI (cli/src/terraform/convert.rs): both have to pass them.
func TestLegacyMoveMatchesCLIParityCases(t *testing.T) {
	raw, err := os.ReadFile("../../../../cli/testdata/terraform_move_parity.json")
	if os.IsNotExist(err) {
		t.Skip("CLI parity cases are not in this checkout")
	}
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		SourceProvider string          `json:"source_provider"`
		SourceType     string          `json:"source_type"`
		State          json.RawMessage `json:"state"`
		Converted      json.RawMessage `json:"converted"`
		// Defaults the provider mover gives an attribute the legacy state lacks. The CLI port leaves
		// it out, so migrate's configuration omits it and plans the same default.
		ProviderDefaults map[string]json.RawMessage `json:"provider_defaults"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, c := range cases {
		t.Run(c.SourceType, func(t *testing.T) {
			var moved *resource.MoveStateResponse
			for _, factory := range New("test")().Resources(ctx) {
				r := factory()
				mover, ok := r.(resource.ResourceWithMoveState)
				if !ok {
					continue
				}
				var s resource.SchemaResponse
				r.Schema(ctx, resource.SchemaRequest{}, &s)
				for _, handler := range mover.MoveState(ctx) {
					resp := moveResponse(s.Schema)
					handler.StateMover(ctx, resource.MoveStateRequest{SourceTypeName: c.SourceType, SourceProviderAddress: c.SourceProvider, SourceRawState: &tfprotov6.RawState{JSON: c.State}}, resp)
					if resp.Diagnostics.HasError() {
						t.Fatal(resp.Diagnostics)
					}
					if !resp.TargetState.Raw.IsNull() {
						moved = resp
					}
				}
			}
			if moved == nil {
				t.Fatal("no registered mover accepted the case")
			}
			decoder := json.NewDecoder(strings.NewReader(string(c.Converted)))
			decoder.UseNumber()
			var want interface{}
			if err := decoder.Decode(&want); err != nil {
				t.Fatal(err)
			}
			got := parityValue(moved.TargetState.Raw).(map[string]interface{})
			for name := range got {
				// Response projections are computed results, not configured inputs.
				if base := strings.TrimSuffix(name, "_result"); base != name && got[base] != nil {
					delete(got, name)
				}
			}
			want = parityDropNulls(want)
			for name, raw := range c.ProviderDefaults {
				decoder := json.NewDecoder(strings.NewReader(string(raw)))
				decoder.UseNumber()
				var value interface{}
				if err := decoder.Decode(&value); err != nil {
					t.Fatal(err)
				}
				if _, ok := want.(map[string]interface{})[name]; ok {
					t.Fatalf("%s: the CLI port converts it, so it is no provider-only default", name)
				}
				want.(map[string]interface{})[name] = value
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("provider mover and CLI port disagree:\n got %v\nwant %v", got, want)
			}
		})
	}
}

func parityValue(v tftypes.Value) interface{} {
	if v.IsNull() || !v.IsKnown() {
		return nil
	}
	switch {
	case v.Type().Is(tftypes.String):
		var s string
		_ = v.As(&s)
		return s
	case v.Type().Is(tftypes.Bool):
		var b bool
		_ = v.As(&b)
		return b
	case v.Type().Is(tftypes.Number):
		f := new(big.Float)
		_ = v.As(&f)
		return json.Number(f.Text('f', -1))
	}
	switch v.Type().(type) {
	case tftypes.Object, tftypes.Map:
		var attributes map[string]tftypes.Value
		_ = v.As(&attributes)
		out := map[string]interface{}{}
		for name, child := range attributes {
			if value := parityValue(child); value != nil {
				out[name] = value
			}
		}
		return out
	}
	var items []tftypes.Value
	_ = v.As(&items)
	out := make([]interface{}, len(items))
	for i, child := range items {
		out[i] = parityValue(child)
	}
	return out
}

func parityDropNulls(value interface{}) interface{} {
	switch value := value.(type) {
	case map[string]interface{}:
		for name, child := range value {
			if child == nil {
				delete(value, name)
			} else {
				value[name] = parityDropNulls(child)
			}
		}
	case []interface{}:
		for i, child := range value {
			value[i] = parityDropNulls(child)
		}
	}
	return value
}
