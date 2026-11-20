package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProviderProtocolSchemaHasNoImplementationDiagnostics(t *testing.T) {
	for name, factory := range testProviderFactories {
		server, err := factory()
		if err != nil {
			t.Fatal(err)
		}
		response, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
		if err != nil {
			t.Fatal(err)
		}
		for _, diagnostic := range response.Diagnostics {
			if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
				t.Errorf("%s: %s: %s", name, diagnostic.Summary, diagnostic.Detail)
			}
		}
	}
}

func TestMetricDataSourceKeepsGroupQueryScopeWhenResultBelongsToSiblingSource(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Errorf("lookup attempted %s", req.Method)
		}
		record := `{"id":"17","attributes":{"source_id":"sibling","name":"duration","sql_expression":"count()","type":"uint64","aggregations":["count"]}}`
		// The source is a path segment: /api/v2/sources/:source_id/metrics has no show, so an id is found in the list.
		if req.URL.Path == "/api/v2/sources/scope/metrics" {
			w.Write([]byte(`{"data":[` + record + `]}`))
		} else {
			t.Errorf("unexpected lookup path %s", req.URL.Path)
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()
	d := &MetricExpressionDataSource{client: testClient(srv.URL)}
	var schemaResponse datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)
	typ := schemaResponse.Schema.Type().TerraformType(ctx)
	config := valueFromAPI(typ, map[string]interface{}{"source_id": "scope", "name": "duration"})
	response := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResponse.Schema, Raw: config}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var result MetricExpressionModel
	if diagnostics := response.State.Get(ctx, &result); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if result.Id.ValueString() != "17" || result.SourceId.ValueString() != "scope" || result.SqlExpression.ValueString() != "count()" {
		t.Fatal("group-scoped lookup lost its configured scope or resolved the wrong expression")
	}
}

func TestMetricDataSourceIDCannotBypassConfiguredQueryScope(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/api/v2/sources/scope/metrics" {
			t.Errorf("identity lookup bypassed its group scope: %s %s", req.Method, req.URL)
		}
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	d := &MetricExpressionDataSource{client: testClient(srv.URL)}
	var schemaResponse datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)
	typ := schemaResponse.Schema.Type().TerraformType(ctx)
	config := valueFromAPI(typ, map[string]interface{}{"id": "17", "source_id": "scope"})
	response := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResponse.Schema, Raw: config}}, &response)
	if !response.Diagnostics.HasError() {
		t.Fatal("accepted an identity outside its configured query scope")
	}
}

func TestDynamicQueryReadKeepsConfiguredShapeAndComputedResult(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Errorf("refresh attempted %s", req.Method)
		}
		w.Write([]byte(`{"data":{"id":"3","attributes":{"name":"example","queries":[{"id":"9","query_type":"sql_expression","sql_query":"SELECT 1","name":"server default"}]}}}`))
	}))
	defer srv.Close()
	r := &ExplorationResource{client: testClient(srv.URL)}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	typ := schemaResponse.Schema.Type().TerraformType(ctx)
	query := map[string]interface{}{"query_type": "sql_expression", "sql_query": "SELECT 1"}
	prior := valueFromAPI(typ, map[string]interface{}{"id": "3", "name": "example", "queries": []interface{}{query}})
	response := resource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	r.Read(ctx, resource.ReadRequest{State: tfsdk.State{Schema: schemaResponse.Schema, Raw: prior}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	result, err := tftypesToAPI(response.State.Raw)
	if err != nil {
		t.Fatal(err)
	}
	state := result.(map[string]interface{})
	if !reflect.DeepEqual(state["queries"], []interface{}{query}) {
		t.Fatalf("refresh changed the configured dynamic shape: %#v", state["queries"])
	}
	computed := state["queries_result"].([]interface{})[0].(map[string]interface{})
	if computed["id"] != "9" || computed["name"] != "server default" {
		t.Fatalf("computed query result lost server fields: %#v", computed)
	}
	var model ExplorationModel
	if diagnostics := response.State.Get(ctx, &model); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	model.Queries = model.QueriesResult
	body := r.body(ctx, model)
	if _, exists := body["queries_result"]; exists {
		t.Fatal("computed result was sent as an API input")
	}
	sent := body["queries"].([]interface{})[0].(map[string]interface{})
	if _, exists := sent["id"]; exists {
		t.Fatal("nested computed identity was sent as an API input")
	}
	if sent["sql_query"] != "SELECT 1" {
		t.Fatal("dynamic query input was lost")
	}
}

// These fixtures are literal production contracts, not values copied from the binding registry. They keep
// the generated binding and the server-side expectation independent.
func TestRepresentativeBindingsMatchLiteralContracts(t *testing.T) {
	tests := []struct {
		name       string
		binding    *apiBinding
		createPath string
		readPath   string
		envelope   envelope
		selector   string
	}{
		{
			name: "uptime monitor", binding: monitorBinding,
			createPath: "/api/v2/monitors", readPath: "/api/v2/monitors/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
			},
		},
		{
			name: "telemetry exploration", binding: explorationBinding,
			createPath: "/api/v2/explorations", readPath: "/api/v2/explorations/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
			},
			selector: "better_stack_team_id",
		},
		{
			name: "errors application", binding: applicationBinding,
			createPath: "/api/v4/applications", readPath: "/api/v4/applications/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
			},
			selector: "better_stack_team_id",
		},
		{
			name: "flat status page resource", binding: statusPageResourceBinding,
			createPath: "/api/v3/status-pages/{status_page_id}/resources",
			readPath:   "/api/v3/status-pages/{status_page_id}/resources/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
			},
		},
		{
			name: "incidents on-call rotation", binding: onCallRotationBinding,
			createPath: "/api/v2/on-calls/{id}/rotation", readPath: "/api/v2/on-calls/{id}/rotation",
			// A bare singleton: the rotation object is the body, addressed by its calendar's id.
			envelope: envelope{ItemPath: []string{}, AttributesPath: []string{}, IDPath: []string{"id"}, RequestPath: []string{}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.binding.Create == nil || tc.binding.Create.Path != tc.createPath {
				t.Fatalf("create path: got %#v want %q", tc.binding.Create, tc.createPath)
			}
			if tc.binding.Read == nil || tc.binding.Read.Path != tc.readPath {
				t.Fatalf("read path: got %#v want %q", tc.binding.Read, tc.readPath)
			}
			if !reflect.DeepEqual(tc.binding.Envelope, tc.envelope) {
				t.Fatalf("envelope: got %#v want %#v", tc.binding.Envelope, tc.envelope)
			}
			want := tc.selector
			if want == "" {
				want = "team_id"
			}
			if tc.binding.Selector == nil || tc.binding.Selector.In != "query" || tc.binding.Selector.Name != want {
				t.Fatalf("selector: got %#v, want query %s", tc.binding.Selector, want)
			}
		})
	}
}

func TestNestedFreeFormValuesAndWriteOnlySecretsSurviveRefresh(t *testing.T) {
	ctx := context.Background()
	settings := map[string]interface{}{"enabled": true, "threshold": json.Number("2.5"),
		"axes": []interface{}{map[string]interface{}{"name": "duration", "visible": false}}}
	chart := flattenExplorationChart(ctx, types.ObjectNull(explorationChartAttrTypes()),
		map[string]interface{}{"chart_type": "line_chart", "settings": settings, "name": "computed-name"})
	body := expandExplorationChart(ctx, chart)
	if !reflect.DeepEqual(body["settings"], settings) {
		t.Fatalf("native nested values were changed: %#v", body)
	}
	if _, sent := body["name"]; sent {
		t.Fatal("readonly nested output leaked into write body")
	}
	bucket := flattenSourceCustomBucket(ctx, types.ObjectNull(sourceCustomBucketAttrTypes()),
		map[string]interface{}{"name": "archive", "endpoint": "https://s3.example", "access_key_id": "key", "keep_data_after_retention": true})
	var prior SourceCustomBucketModel
	bucket.As(ctx, &prior, basetypes.ObjectAsOptions{})
	prior.SecretAccessKey = types.StringValue("saved-secret")
	bucket, diags := types.ObjectValueFrom(ctx, sourceCustomBucketAttrTypes(), prior)
	if diags.HasError() {
		t.Fatal(diags)
	}
	refreshed := flattenSourceCustomBucket(ctx, bucket, map[string]interface{}{"name": "archive", "secret_access_key": nil})
	if got := expandSourceCustomBucket(ctx, refreshed)["secret_access_key"]; got != "saved-secret" {
		t.Fatalf("refresh erased the write-only credential: %v", got)
	}
}

func TestClientAgainstLiteralErrorsApplicationFixture(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v4/applications" {
			http.Error(w, "unexpected literal route", http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("better_stack_team_id") != "t123" {
			http.Error(w, "missing literal team selector", http.StatusBadRequest)
			return
		}
		if r.Header.Get("Better-Stack-Version") != "" {
			http.Error(w, "the path API takes no version header", http.StatusBadRequest)
			return
		}
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["name"] != "frontend" || body["data"] != nil {
			http.Error(w, "expected a flat application body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"7","type":"application","attributes":{"name":"frontend"}}}`))
	}))
	defer srv.Close()
	c := NewClient("token", srv.URL, map[string]string{"better_stack_team_id": "t123"}, "test")
	out, err := c.Create(context.Background(), applicationBinding, nil, map[string]interface{}{"name": "frontend"})
	if err != nil {
		t.Fatal(err)
	}
	if out["id"] != "7" || out["name"] != "frontend" {
		t.Fatalf("literal response envelope was not decoded: %#v", out)
	}
}

func TestExplorationBodyUsesConfigPresenceAndPlannedUnknowns(t *testing.T) {
	r := &ExplorationResource{}
	config := ExplorationModel{
		Name:               types.StringValue("resolved-name"),
		ExplorationGroupId: types.Int64Null(),
	}
	body := r.body(context.Background(), config)
	if body["name"] != "resolved-name" {
		t.Fatalf("configured unknown must use its planned value: %#v", body)
	}
	if _, present := body["exploration_group_id"]; present {
		t.Fatalf("omitted nullable field must stay omitted: %#v", body)
	}

	plan := config
	state := config
	state.ExplorationGroupId = types.Int64Value(17)
	body = r.updateBody(context.Background(), config, plan, state)
	if value, present := body["exploration_group_id"]; !present || value != nil {
		t.Fatalf("removing a prior value must be present as JSON null: %#v", body)
	}
}

func TestConfiguredUnknownsUsePlanWithoutReplacingOmittedNulls(t *testing.T) {
	typeOf := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String, "exploration_group_id": tftypes.String,
	}}
	config := tftypes.NewValue(typeOf, map[string]tftypes.Value{
		"name":                 tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"exploration_group_id": tftypes.NewValue(tftypes.String, nil),
	})
	plan := tftypes.NewValue(typeOf, map[string]tftypes.Value{
		"name":                 tftypes.NewValue(tftypes.String, "resolved-name"),
		"exploration_group_id": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
	})
	var configured map[string]tftypes.Value
	_ = configuredFromPlan(config, plan, nullValue(typeOf)).As(&configured)
	var name string
	if err := configured["name"].As(&name); err != nil || name != "resolved-name" {
		t.Fatalf("configured unknown did not use plan: name=%q err=%v", name, err)
	}
	if !configured["exploration_group_id"].IsNull() || !configured["exploration_group_id"].Type().Is(tftypes.String) {
		t.Fatalf("omitted null was replaced: %#v", configured["exploration_group_id"])
	}
}
