package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func TestExtractedMetricDataSourceKeepsGroupQueryScopeWhenResultBelongsToSiblingSource(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Errorf("lookup attempted %s", req.Method)
		}
		record := `{"id":"17","attributes":{"source_id":"sibling","name":"duration","sql_expression":"count()","type":"uint64","aggregations":["count"]}}`
		// The source is a path segment: /api/v4/sources/:source_id/extracted-metrics has no show, so an id is found in the list.
		if req.URL.Path == "/api/v4/sources/scope/extracted-metrics" {
			w.Write([]byte(`{"data":[` + record + `]}`))
		} else {
			t.Errorf("unexpected lookup path %s", req.URL.Path)
			http.NotFound(w, req)
		}
	}))
	defer srv.Close()
	d := &ExtractedMetricDataSource{client: testClient(srv.URL)}
	var schemaResponse datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)
	typ := schemaResponse.Schema.Type().TerraformType(ctx)
	config := valueFromAPI(typ, map[string]interface{}{"source_id": "scope", "name": "duration"})
	response := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResponse.Schema, Raw: config}}, &response)
	if response.Diagnostics.HasError() {
		t.Fatal(response.Diagnostics)
	}
	var result ExtractedMetricModel
	if diagnostics := response.State.Get(ctx, &result); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	if result.Id.ValueString() != "17" || result.SourceId.ValueString() != "scope" || result.SqlExpression.ValueString() != "count()" {
		t.Fatal("group-scoped lookup lost its configured scope or resolved the wrong expression")
	}
}

func TestExtractedMetricDataSourceIDCannotBypassConfiguredQueryScope(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/api/v4/sources/scope/extracted-metrics" {
			t.Errorf("identity lookup bypassed its group scope: %s %s", req.Method, req.URL)
		}
		w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()
	d := &ExtractedMetricDataSource{client: testClient(srv.URL)}
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

func TestQueryReadFillsServerFieldsAndWritesOnlyInputs(t *testing.T) {
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
	var model ExplorationModel
	if diagnostics := response.State.Get(ctx, &model); diagnostics.HasError() {
		t.Fatal(diagnostics)
	}
	body := r.body(ctx, model)
	sent := body["queries"].([]map[string]interface{})[0]
	if _, exists := sent["id"]; exists {
		t.Fatal("the read-only query id was sent as an API input")
	}
	if sent["sql_query"] != "SELECT 1" || sent["name"] != "server default" {
		t.Fatalf("query inputs were lost: %#v", sent)
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
			createPath: "/api/v3/monitors", readPath: "/api/v3/monitors/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: nil, // monitors link records without sideloading them: no included list
			},
		},
		{
			name: "telemetry exploration", binding: explorationBinding,
			createPath: "/api/v3/explorations", readPath: "/api/v3/explorations/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
			},
			selector: "team_id",
		},
		{
			name: "errors application", binding: applicationBinding,
			createPath: "/api/v4/applications", readPath: "/api/v4/applications/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
			},
			selector: "team_id",
		},
		{
			name: "flat status page", binding: statusPageBinding,
			createPath: "/api/v5/status-pages",
			readPath:   "/api/v5/status-pages/{id}",
			envelope: envelope{
				ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
				IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
				RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
			},
		},
		{
			name: "incidents on-call rotation", binding: onCallRotationBinding,
			createPath: "/api/v3/on-call-schedules/{id}/rotation", readPath: "/api/v3/on-call-schedules/{id}/rotation",
			// A bare singleton: the rotation object is the body, addressed by its schedule's id.
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

func compactOrFail(t *testing.T, v interface{}) string {
	text, ok := compactJSON(v)
	if !ok {
		t.Fatalf("not JSON: %#v", v)
	}
	return text
}

// A webhook body template is text that need not be valid JSON: the on-call template puts
// $EMAILS_JSON_ARRAY in unquoted. It reaches the request exactly as written, and a template that is
// valid JSON is sent as the same string, never decoded into an object.
func TestOutgoingWebhookBodyTemplateIsSentAsWritten(t *testing.T) {
	ctx := context.Background()
	templates := []string{
		"{\n  \"data\":{\n    \"message\": \"$MESSAGE\",\n    \"emails\": $EMAILS_JSON_ARRAY,\n    \"calendar_id\": $ON_CALL_CALENDAR_ID\n  }\n}\n",
		"{\"incident_id\": \"$INCIDENT_ID\",  \"cause\": \"$CAUSE\"}",
	}
	for _, template := range templates {
		model := IncidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttributesModel{
			Id: types.StringNull(), HttpMethod: types.StringValue("post"),
			HeadersTemplate: types.ListNull(types.ObjectType{AttrTypes: incidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttributesHeadersTemplateAttrTypes()}),
			BodyTemplate:    types.StringValue(template), AuthUsername: types.StringNull(), AuthPassword: types.StringNull(),
		}
		obj, diags := types.ObjectValueFrom(ctx, incidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttributesAttrTypes(), model)
		if diags.HasError() {
			t.Fatal(diags)
		}
		body := expandIncidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttributes(ctx, obj)
		got, ok := body["body_template"].(string)
		if !ok || got != template {
			t.Fatalf("body_template wasn't sent as written: %#v", body["body_template"])
		}
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		var wire map[string]interface{}
		if err := json.Unmarshal(encoded, &wire); err != nil || wire["body_template"] != template {
			t.Fatalf("the request body doesn't carry the template unchanged: %s", encoded)
		}
	}
}

func TestNestedSettingsAndWriteOnlySecretsSurviveRefresh(t *testing.T) {
	ctx := context.Background()
	settings := map[string]interface{}{"unit": "ms", "decimal_places": json.Number("2"), "custom_unit": false,
		"y_axis_annotations": []interface{}{map[string]interface{}{"value": json.Number("2.5"), "label": "Slow", "color": "#e0687d"}}}
	chart := flattenExplorationChart(ctx, types.ObjectNull(explorationChartAttrTypes()),
		map[string]interface{}{"chart_type": "line_chart", "settings": settings, "name": "computed-name"})
	body := expandExplorationChart(ctx, chart)
	if !sameJSON(compactOrFail(t, body["settings"]), compactOrFail(t, settings)) {
		t.Fatalf("nested settings were changed: %#v", body)
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
		if r.URL.Query().Get("team_id") != "t123" {
			http.Error(w, "missing literal team selector", http.StatusBadRequest)
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
	c := NewClient("token", srv.URL, map[string]string{"team_id": "t123"}, "test")
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
		ExplorationGroupId: types.StringNull(),
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
	state.ExplorationGroupId = types.StringValue("17")
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

// A when: condition on fields inside an object sibling: absent (or a null object) applies, a set
// field doesn't, and an unknown one leaves the plan alone until a known mismatch decides.
func TestWhenReadsAFieldInsideAnObjectSibling(t *testing.T) {
	obj := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"policy_id": tftypes.String, "policy_name": tftypes.String}}
	root := func(target tftypes.Value) tftypes.Value {
		return tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"escalation_target": obj}}, map[string]tftypes.Value{"escalation_target": target})
	}
	target := func(id, name tftypes.Value) tftypes.Value {
		return tftypes.NewValue(obj, map[string]tftypes.Value{"policy_id": id, "policy_name": name})
	}
	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	null := tftypes.NewValue(tftypes.String, nil)
	unknown := tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	conds := []siblingCond{{Path: "escalation_target.policy_id", Absent: true}, {Path: "escalation_target.policy_name", Absent: true}}
	for _, c := range []struct {
		name         string
		raw          tftypes.Value
		known, match bool
	}{
		{"a team", root(target(null, null)), true, true},
		{"no target", root(tftypes.NewValue(obj, nil)), true, true},
		{"a policy", root(target(str("7"), str("Critical"))), true, false},
		{"a policy by name", root(target(unknown, str("Critical"))), true, false},
		{"a target not yet known", root(tftypes.NewValue(obj, tftypes.UnknownValue)), false, false},
		{"a policy id not yet known", root(target(unknown, null)), false, false},
	} {
		if known, match := siblingMatches(c.raw, conds); known != c.known || match != c.match {
			t.Errorf("%s: known=%v match=%v, want %v %v", c.name, known, match, c.known, c.match)
		}
	}
	unknownTarget := root(tftypes.NewValue(obj, tftypes.UnknownValue))
	noTarget := root(tftypes.NewValue(obj, nil))
	team, policy := root(target(null, null)), root(target(str("7"), str("Critical")))
	for _, c := range []struct {
		name                string
		plan, config, state tftypes.Value
		known, match        bool
	}{
		{"a target configured as a team, planned unknown", root(target(unknown, unknown)), team, team, true, true},
		{"a target switched back to a team, the plan still holding the policy", policy, team, policy, true, true},
		{"an alert created without a target", unknownTarget, noTarget, tftypes.NewValue(noTarget.Type(), nil), true, true},
		{"an unconfigured target kept from a team state, planned unknown", unknownTarget, noTarget, team, true, true},
		{"an unconfigured target kept from a policy state", unknownTarget, noTarget, policy, true, false},
	} {
		if known, match := siblingMatchesPlan(c.plan, c.config, c.state, conds); known != c.known || match != c.match {
			t.Errorf("%s: known=%v match=%v, want %v %v", c.name, known, match, c.known, c.match)
		}
	}
	if !keptOut(team, types.BoolNull(), noTarget, conds) || keptOut(team, types.BoolValue(false), noTarget, conds) || keptOut(team, types.BoolNull(), team, conds) {
		t.Errorf("keptOut: a flag the API left out stays out only while the target isn't configured")
	}
	if got := appliesWhenText(path.Root("call"), conds); !strings.Contains(got, "escalation_target.policy_id is null and escalation_target.policy_name is null") {
		t.Errorf("text = %q", got)
	}
}

// An imported webhook's write-only template is read from the template the API returns, and a value
// state already holds stays as it is.
func TestWriteOnlyTemplateIsReadOnlyWhenStateHasNone(t *testing.T) {
	ctx := context.Background()
	attrs := map[string]interface{}{"custom_webhook_template": map[string]interface{}{"id": "7", "http_method": "post", "body_template": "{}"}}
	var imported IncidentsOutgoingWebhookIntegrationModel
	imported.CustomWebhookTemplateAttributes = types.ObjectNull(incidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttributesAttrTypes())
	imported.CustomWebhookTemplate = types.ObjectNull(incidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttrTypes())
	populateIncidentsOutgoingWebhookIntegration(ctx, &imported, attrs)
	if imported.CustomWebhookTemplateAttributes.IsNull() {
		t.Fatal("an import left the write-only template empty")
	}
	if got := imported.CustomWebhookTemplateAttributes.Attributes()["http_method"]; got.String() != `"post"` {
		t.Fatalf("http_method = %s", got)
	}
	configured := imported
	configured.CustomWebhookTemplateAttributes = flattenIncidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttributes(ctx,
		types.ObjectNull(incidentsOutgoingWebhookIntegrationCustomWebhookTemplateAttributesAttrTypes()), map[string]interface{}{"http_method": "put"})
	populateIncidentsOutgoingWebhookIntegration(ctx, &configured, attrs)
	if got := configured.CustomWebhookTemplateAttributes.Attributes()["http_method"]; got.String() != `"put"` {
		t.Fatalf("a configured value was replaced: %s", got)
	}
}

// A catalog item_id takes a team only as t123, and any id for another kind of item.
func TestPatternWhenRefusesABareTeamNumber(t *testing.T) {
	ctx := context.Background()
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"type": tftypes.String, "item_id": tftypes.String}}
	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"value": objType}}
	sch := schema.Schema{Attributes: map[string]schema.Attribute{"value": schema.SingleNestedAttribute{Optional: true, Attributes: map[string]schema.Attribute{
		"type": schema.StringAttribute{Optional: true}, "item_id": schema.StringAttribute{Optional: true}}}}}
	v := patternWhenString{Sibling: "type", Patterns: map[string]string{"Team": "^t[1-9][0-9]*$"}}
	for _, c := range []struct {
		kind, id string
		ok       bool
	}{{"Team", "t552385", true}, {"Team", "552385", false}, {"Policy", "131543", true}, {"Schedule", "77", true}} {
		raw := tftypes.NewValue(rootType, map[string]tftypes.Value{"value": tftypes.NewValue(objType, map[string]tftypes.Value{
			"type": tftypes.NewValue(tftypes.String, c.kind), "item_id": tftypes.NewValue(tftypes.String, c.id)})})
		req := validator.StringRequest{Path: path.Root("value").AtName("item_id"), ConfigValue: types.StringValue(c.id), Config: tfsdk.Config{Raw: raw, Schema: sch}}
		resp := &validator.StringResponse{}
		v.ValidateString(ctx, req, resp)
		if resp.Diagnostics.HasError() == c.ok {
			t.Errorf("%s %s: errors %v, want ok=%v", c.kind, c.id, resp.Diagnostics, c.ok)
		}
	}
}
