package provider

import (
	"context"
	"encoding/json"
	"math/big"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStateFromPlanResolvesTopLevelUnknowns(t *testing.T) {
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id":     tftypes.String,
		"url":    tftypes.String,
		"port":   tftypes.Number,
		"paused": tftypes.Bool,
	}}
	plan := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":     tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"url":    tftypes.NewValue(tftypes.String, "https://example.com"),
		"port":   tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		"paused": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
	})
	got := stateFromPlan(plan, nullValue(objType), map[string]interface{}{
		"id":   "42",
		"port": json.Number("8080"),
		// paused absent -> null
	})
	if !got.IsFullyKnown() {
		t.Fatalf("state still contains unknowns: %v", got)
	}
	var attrs map[string]tftypes.Value
	_ = got.As(&attrs)
	var id string
	_ = attrs["id"].As(&id)
	if id != "42" {
		t.Errorf("id = %q", id)
	}
	var port big.Float
	_ = attrs["port"].As(&port)
	if v, _ := port.Int64(); v != 8080 {
		t.Errorf("port = %v", port.String())
	}
	if !attrs["paused"].IsNull() {
		t.Errorf("paused = %v, want null (absent from response)", attrs["paused"])
	}
	var url string
	_ = attrs["url"].As(&url)
	if url != "https://example.com" {
		t.Errorf("known planned url must be untouched, got %q", url)
	}
}

// A computed sub-attribute inside a user-configured list element is unknown at plan time;
// the walker must resolve it from the matching response element.
func TestStateFromPlanResolvesNestedUnknowns(t *testing.T) {
	stepType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"type":        tftypes.String,
		"severity_id": tftypes.Number,
		"wait_before": tftypes.Number,
	}}
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"steps": tftypes.List{ElementType: stepType},
	}}
	plan := tftypes.NewValue(objType, map[string]tftypes.Value{
		"steps": tftypes.NewValue(tftypes.List{ElementType: stepType}, []tftypes.Value{
			tftypes.NewValue(stepType, map[string]tftypes.Value{
				"type":        tftypes.NewValue(tftypes.String, "escalation"),
				"severity_id": tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
				"wait_before": tftypes.NewValue(tftypes.Number, big.NewFloat(180)),
			}),
		}),
	})
	got := stateFromPlan(plan, nullValue(objType), map[string]interface{}{
		"steps": []interface{}{map[string]interface{}{"type": "escalation", "severity_id": json.Number("3")}},
	})
	if !got.IsFullyKnown() {
		t.Fatalf("nested unknown not resolved: %v", got)
	}
	var attrs map[string]tftypes.Value
	_ = got.As(&attrs)
	var steps []tftypes.Value
	_ = attrs["steps"].As(&steps)
	var step map[string]tftypes.Value
	_ = steps[0].As(&step)
	var severity big.Float
	_ = step["severity_id"].As(&severity)
	if v, _ := severity.Int64(); v != 3 {
		t.Errorf("severity_id = %s, want 3", severity.String())
	}
	var wait big.Float
	_ = step["wait_before"].As(&wait)
	if v, _ := wait.Int64(); v != 180 {
		t.Errorf("known wait_before must be untouched, got %s", wait.String())
	}
}

func TestConfiguredFromPlanRecursesThroughKnownObjectsListsAndSets(t *testing.T) {
	elementType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String, "resolved": tftypes.String,
	}}
	listType := tftypes.List{ElementType: elementType}
	setType := tftypes.Set{ElementType: elementType}
	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"object": elementType, "list": listType, "set": setType,
	}}
	element := func(resolved interface{}) tftypes.Value {
		return tftypes.NewValue(elementType, map[string]tftypes.Value{
			"name":     tftypes.NewValue(tftypes.String, "configured"),
			"resolved": tftypes.NewValue(tftypes.String, resolved),
		})
	}
	root := func(resolved interface{}) tftypes.Value {
		return tftypes.NewValue(rootType, map[string]tftypes.Value{
			"object": element(resolved),
			"list":   tftypes.NewValue(listType, []tftypes.Value{element(resolved)}),
			"set":    tftypes.NewValue(setType, []tftypes.Value{element(resolved)}),
		})
	}

	got := configuredFromPlan(root(tftypes.UnknownValue), root(tftypes.UnknownValue), root("prior-state"))
	if !got.IsFullyKnown() {
		t.Fatalf("nested configured unknowns were not resolved: %v", got)
	}
	assertNestedResolvedValues(t, got, "prior-state")
}

func TestStateFromPlanUsesPriorStateForNestedUnknownsMissingFromResponse(t *testing.T) {
	elementType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String, "resolved": tftypes.String,
	}}
	listType := tftypes.List{ElementType: elementType}
	setType := tftypes.Set{ElementType: elementType}
	rootType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"object": elementType, "list": listType, "set": setType,
	}}
	element := func(resolved interface{}) tftypes.Value {
		return tftypes.NewValue(elementType, map[string]tftypes.Value{
			"name":     tftypes.NewValue(tftypes.String, "configured"),
			"resolved": tftypes.NewValue(tftypes.String, resolved),
		})
	}
	root := func(resolved interface{}) tftypes.Value {
		return tftypes.NewValue(rootType, map[string]tftypes.Value{
			"object": element(resolved),
			"list":   tftypes.NewValue(listType, []tftypes.Value{element(resolved)}),
			"set":    tftypes.NewValue(setType, []tftypes.Value{element(resolved)}),
		})
	}

	got := stateFromPlan(root(tftypes.UnknownValue), root("prior-state"), map[string]interface{}{})
	if !got.IsFullyKnown() {
		t.Fatalf("nested planned unknowns were not resolved from state: %v", got)
	}
	assertNestedResolvedValues(t, got, "prior-state")
}

func assertNestedResolvedValues(t *testing.T, root tftypes.Value, want string) {
	t.Helper()
	var attrs map[string]tftypes.Value
	if err := root.As(&attrs); err != nil {
		t.Fatal(err)
	}
	values := []tftypes.Value{attrs["object"]}
	for _, name := range []string{"list", "set"} {
		var elems []tftypes.Value
		if err := attrs[name].As(&elems); err != nil || len(elems) != 1 {
			t.Fatalf("%s = %v, err=%v", name, attrs[name], err)
		}
		values = append(values, elems[0])
	}
	for i, value := range values {
		var nested map[string]tftypes.Value
		if err := value.As(&nested); err != nil {
			t.Fatal(err)
		}
		var got string
		if err := nested["resolved"].As(&got); err != nil || got != want {
			t.Fatalf("container %d resolved = %q, err=%v; want %q", i, got, err, want)
		}
	}
}

func TestValueFromAPICoercesLooseTypes(t *testing.T) {
	if v := valueFromAPI(tftypes.String, json.Number("987")); v.IsNull() {
		t.Error("stringified ids: number -> string must coerce")
	} else {
		var s string
		_ = v.As(&s)
		if s != "987" {
			t.Errorf("coerced string = %q", s)
		}
	}
	if v := valueFromAPI(tftypes.Number, "1.5"); v.IsNull() {
		t.Error("numeric strings must coerce to numbers")
	}
	if v := valueFromAPI(tftypes.Bool, "yes"); !v.IsNull() {
		t.Error("junk for bool must become null, not panic")
	}
}

func TestDynamicFromAPIInference(t *testing.T) {
	v := dynamicFromAPI(map[string]interface{}{
		"policy_id": json.Number("7"),
		"tags":      []interface{}{"a", json.Number("2")},
	})
	if !v.Type().Is(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"policy_id": tftypes.Number,
		"tags":      tftypes.Tuple{ElementTypes: []tftypes.Type{tftypes.String, tftypes.Number}},
	}}) {
		t.Errorf("inferred type = %v", v.Type())
	}
}

func TestStringifyScalarKeepsDigits(t *testing.T) {
	if got := stringifyScalar(json.Number("1750000000000001")); got != "1750000000000001" {
		t.Errorf("large id corrupted: %q", got)
	}
	if got := stringifyScalar(true); got != "true" {
		t.Errorf("bool = %q", got)
	}
}

func TestCompactJSONMatchesJsonencode(t *testing.T) {
	got, ok := compactJSON(map[string]interface{}{
		"sql":   "env = 'prod' AND x < 3",
		"count": json.Number("2"),
	})
	if !ok {
		t.Fatal("compactJSON failed")
	}
	// Sorted keys, no HTML escaping, exact numbers — Terraform's jsonencode() shape.
	want := `{"count":2,"sql":"env = 'prod' AND x < 3"}`
	if got != want {
		t.Errorf("compactJSON = %s, want %s", got, want)
	}
}

func TestTftypesToAPIRoundTrip(t *testing.T) {
	v := tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"n": tftypes.Number,
		"s": tftypes.String,
	}}, map[string]tftypes.Value{
		"n": tftypes.NewValue(tftypes.Number, big.NewFloat(42)),
		"s": tftypes.NewValue(tftypes.String, "x"),
	})
	api, err := tftypesToAPI(v)
	if err != nil {
		t.Fatalf("tftypesToAPI: %v", err)
	}
	m := api.(map[string]interface{})
	if m["n"] != json.Number("42") || m["s"] != "x" {
		t.Errorf("api = %#v", m)
	}
}

func TestADatetimeRefreshKeepsTheSameInstantAsWritten(t *testing.T) {
	written := types.StringValue("2026-11-23T00:00:00Z")
	if got := datetimeValue(written, "2026-11-23T00:00:00.000Z"); !got.Equal(written) {
		t.Fatalf("the same instant written another way must keep the configured form, got %s", got)
	}
	if got := datetimeValue(written, "2026-11-23T01:00:00+01:00"); !got.Equal(written) {
		t.Fatalf("an offset naming the same instant must keep the configured form, got %s", got)
	}
	if got := datetimeValue(written, "2026-11-24T00:00:00.000Z"); got.ValueString() != "2026-11-24T00:00:00Z" {
		t.Fatalf("a different instant is drift and must be refreshed, got %s", got)
	}
	if got := datetimeValue(types.StringNull(), "2026-11-23T00:00:00.000Z"); got.ValueString() != "2026-11-23T00:00:00Z" {
		t.Fatalf("with no prior value, as after an import, the instant is stored in its shortest form, got %s", got)
	}
	if got := datetimeValue(types.StringNull(), "2026-11-23T01:00:00.250+01:00"); got.ValueString() != "2026-11-23T01:00:00.25+01:00" {
		t.Fatalf("the shortest form keeps the API's offset and a nonzero fraction, got %s", got)
	}
	if got := datetimeValue(types.StringNull(), "next tuesday"); got.ValueString() != "next tuesday" {
		t.Fatalf("a value that isn't RFC 3339 is stored as the API sent it, got %s", got)
	}
	if got := datetimeValue(types.StringValue("not a time"), "2026-11-23T00:00:00.000Z"); got.ValueString() != "2026-11-23T00:00:00Z" {
		t.Fatalf("an unparsable prior value is replaced, got %s", got)
	}
}

// A list of parents, each with a nested list of rows, as an item key sees it.
var (
	matchRowType    = map[string]attr.Type{"id": types.StringType, "kind": types.StringType, "ref": types.Int64Type, "label": types.StringType, "note": types.StringType}
	matchParentType = map[string]attr.Type{"id": types.StringType, "name": types.StringType, "rows": types.ListType{ElemType: types.ObjectType{AttrTypes: matchRowType}}}
	matchSpec       = itemMatch{Keys: [][]string{{"name"}}, Children: map[string]itemMatch{"rows": {Keys: [][]string{{"kind", "ref"}, {"kind", "label"}}}}}
)

func matchRow(id attr.Value, ref attr.Value, label, note string) attr.Value {
	noteValue := attr.Value(types.StringUnknown())
	if note != "" {
		noteValue = types.StringValue(note)
	}
	return types.ObjectValueMust(matchRowType, map[string]attr.Value{"id": id, "kind": types.StringValue("Item"), "ref": ref, "label": types.StringValue(label), "note": noteValue})
}

func matchParent(id attr.Value, name string, rows ...attr.Value) attr.Value {
	return types.ObjectValueMust(matchParentType, map[string]attr.Value{"id": id, "name": types.StringValue(name), "rows": types.ListValueMust(types.ObjectType{AttrTypes: matchRowType}, rows)})
}

func matchList(parents ...attr.Value) types.List {
	return types.ListValueMust(types.ObjectType{AttrTypes: matchParentType}, parents)
}

func rowIDs(t *testing.T, v attr.Value) [][]string {
	t.Helper()
	var out [][]string
	for _, parent := range listObjects(v) {
		ids := []string{parent.Attributes()["id"].String()}
		for _, row := range listObjects(parent.Attributes()["rows"]) {
			ids = append(ids, row.Attributes()["id"].String()+"/"+row.Attributes()["note"].String())
		}
		out = append(out, ids)
	}
	return out
}

// A reorder and a move keep every id and unconfigured value; a row without a known first key falls
// back to the next; an explicit id is never taken by another row first; a new row stays unknown.
func TestMatchPlannedListKeepsIdsThroughReordersAndMoves(t *testing.T) {
	ctx := context.Background()
	s, unknown, none := types.StringValue, types.StringUnknown(), types.Int64Null()
	prior := matchList(
		matchParent(s("1"), "A", matchRow(s("11"), types.Int64Value(7), "seven", "kept"), matchRow(s("12"), none, "manual", "by label")),
		matchParent(s("2"), "B", matchRow(s("21"), types.Int64Value(8), "eight", "x"), matchRow(s("22"), types.Int64Value(8), "eight again", "y")),
	)
	planned := matchList(
		matchParent(unknown, "B", matchRow(unknown, types.Int64Value(8), "eight", ""), matchRow(unknown, types.Int64Value(7), "seven", "")),
		matchParent(unknown, "A", matchRow(unknown, none, "manual", ""), matchRow(s("21"), types.Int64Value(8), "eight", ""), matchRow(unknown, types.Int64Value(9), "new", "")),
	)
	var diags diag.Diagnostics
	got := rowIDs(t, matchPlannedList(ctx, planned, prior, matchSpec, &diags))
	want := [][]string{
		{`"2"`, `"22"/"y"`, `"11"/"kept"`},
		{`"1"`, `"12"/"by label"`, `"21"/"x"`, `<unknown>/<unknown>`},
	}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("parent %d: got %v, want %v", i, got[i], want[i])
		}
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("parent %d: got %v, want %v", i, got[i], want[i])
			}
		}
	}

	// The write carries the matched ids, and nothing else the configuration leaves out.
	config := matchList(
		matchParent(types.StringNull(), "B", matchRow(types.StringNull(), types.Int64Value(8), "eight", "")),
	)
	plannedWrite := matchList(matchParent(s("2"), "B", matchRow(s("22"), types.Int64Value(8), "eight", "y")))
	request := withPlannedIDs(ctx, config, plannedWrite, matchSpec, &diags)
	row := listObjects(listObjects(request)[0].Attributes()["rows"])[0].Attributes()
	if row["id"].String() != `"22"` || !row["note"].IsUnknown() || diags.HasError() {
		t.Fatalf("request row = %v, diagnostics %v", row, diags)
	}

	// Two parents with the same key are matched in order, each to its own prior item.
	twins := matchList(matchParent(s("1"), "A"), matchParent(s("3"), "A"))
	twinsPlanned := matchList(matchParent(unknown, "A"), matchParent(unknown, "A"))
	if got := rowIDs(t, matchPlannedList(ctx, twinsPlanned, twins, matchSpec, &diags)); got[0][0] != `"1"` || got[1][0] != `"3"` {
		t.Fatalf("parents with the same name: got %v", got)
	}

	// A request it can't pair with the plan fails, instead of going out without ids.
	withPlannedIDs(ctx, config, matchList(matchParent(s("2"), "B"), matchParent(s("1"), "A")), matchSpec, &diags)
	if !diags.HasError() {
		t.Fatal("an unpaired request was sent without an error")
	}
}

// A read list drops the API's own items unless configured, and follows the configured order.
func TestKeyedAPIReadDropsBuiltInsAndKeepsTheConfiguredOrder(t *testing.T) {
	item := func(name string) interface{} { return map[string]interface{}{"name": name} }
	raw := []interface{}{item("end_time"), item("env"), item("source"), item("start_time"), item("svc"), item("time")}
	names := func(items []interface{}) []string {
		out := []string{}
		for _, i := range items {
			out = append(out, i.(map[string]interface{})["name"].(string))
		}
		return out
	}
	builtIns := []string{"time", "start_time", "end_time", "source"}
	got := names(keyedAPIRead(raw, "name", []string{"svc", "source", "env"}, true, builtIns))
	want := []string{"svc", "source", "env"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
	if got := names(keyedAPIRead(raw, "name", nil, true, builtIns)); len(got) != 2 || got[0] != "env" || got[1] != "svc" {
		t.Fatalf("an import keeps only the user's items in API order, got %v", got)
	}
	twice := append([]interface{}{item("env")}, raw...)
	if got := names(keyedAPIRead(twice, "name", []string{"env"}, true, builtIns)); len(got) != 3 || got[0] != "env" || got[1] != "env" || got[2] != "svc" {
		t.Fatalf("a second item with a configured name must stay, got %v", got)
	}
	if got := names(keyedAPIRead(raw, "name", []string{"svc"}, false, builtIns)); len(got) != 2 || got[0] != "env" || got[1] != "svc" {
		t.Fatalf("with an unknown planned name the read keeps the API order, got %v", got)
	}
}

func TestSameJSONIgnoresSpacingAndKeyOrder(t *testing.T) {
	if !sameJSON(`{ "b": 1, "a": [1, 2] }`, `{"a":[1,2],"b":1}`) {
		t.Fatal("the same value spelled differently is the same JSON")
	}
	if sameJSON(`{"a":1}`, `{"a":2}`) || sameJSON(`not json`, `not json`) {
		t.Fatal("different values, or text that isn't JSON, are not the same JSON")
	}
}

// A camelCase wire key answers to its snake_case Terraform name.
// A renamed top-level item keeps the id of the prior item at its position, which nothing else took; an
// added one stays unknown, and a prior item another item took by name is never taken by position.
func TestMatchPlannedListKeepsARenamedItemsIdByPosition(t *testing.T) {
	ctx := context.Background()
	s, unknown := types.StringValue, types.StringUnknown()
	prior := matchList(matchParent(s("1"), "A"), matchParent(s("2"), "B"))
	var diags diag.Diagnostics
	renamed := matchList(matchParent(unknown, "A"), matchParent(unknown, "Bee"))
	if got := rowIDs(t, matchPlannedList(ctx, renamed, prior, matchSpec, &diags)); got[0][0] != `"1"` || got[1][0] != `"2"` {
		t.Fatalf("rename: got %v", got)
	}
	added := matchList(matchParent(unknown, "New"), matchParent(unknown, "A"), matchParent(unknown, "B"))
	if got := rowIDs(t, matchPlannedList(ctx, added, prior, matchSpec, &diags)); got[0][0] != `<unknown>` || got[1][0] != `"1"` || got[2][0] != `"2"` {
		t.Fatalf("an added item took a prior id: got %v", got)
	}
}

func TestAPIFieldMatchesTheSnakeCaseName(t *testing.T) {
	m := map[string]interface{}{"canonicalSourceIds": []interface{}{"a"}, "selected_label": "x"}
	if v, ok := apiField(m, "canonical_source_ids"); !ok || v == nil {
		t.Fatal("canonicalSourceIds did not answer to canonical_source_ids")
	}
	if v, ok := apiField(m, "selected_label"); !ok || v != "x" {
		t.Fatal("a snake_case key must match itself")
	}
	if _, ok := apiField(m, "service_ids"); ok {
		t.Fatal("a missing key matched")
	}
}

// An object configured with no field set plans every field empty, and only then.
func TestEmptyConfiguredObjectPlansEveryFieldEmpty(t *testing.T) {
	ctx := context.Background()
	types_ := map[string]attr.Type{"policy_id": types.StringType, "team_id": types.StringType}
	empty := types.ObjectValueMust(types_, map[string]attr.Value{"policy_id": types.StringNull(), "team_id": types.StringNull()})
	prior := types.ObjectValueMust(types_, map[string]attr.Value{"policy_id": types.StringValue("1"), "team_id": types.StringNull()})
	resp := &planmodifier.ObjectResponse{PlanValue: prior}
	emptyConfiguredObject{}.PlanModifyObject(ctx, planmodifier.ObjectRequest{ConfigValue: empty, PlanValue: prior}, resp)
	if !resp.PlanValue.Equal(empty) {
		t.Fatalf("an empty object kept the current fields: %v", resp.PlanValue)
	}
	team := types.ObjectValueMust(types_, map[string]attr.Value{"policy_id": types.StringNull(), "team_id": types.StringValue("2")})
	resp = &planmodifier.ObjectResponse{PlanValue: team}
	emptyConfiguredObject{}.PlanModifyObject(ctx, planmodifier.ObjectRequest{ConfigValue: team, PlanValue: team}, resp)
	if !resp.PlanValue.Equal(team) {
		t.Fatalf("a configured field changed the plan: %v", resp.PlanValue)
	}
}

// Resolving a plan against prior state without an attribute takes it from the API response alone.
func TestWithoutAttributesNullsOnlyTheNamedOnes(t *testing.T) {
	typ := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"a": tftypes.String, "b": tftypes.String}}
	v := tftypes.NewValue(typ, map[string]tftypes.Value{"a": tftypes.NewValue(tftypes.String, "x"), "b": tftypes.NewValue(tftypes.String, "y")})
	var attrs map[string]tftypes.Value
	if err := withoutAttributes(v, "a").As(&attrs); err != nil {
		t.Fatal(err)
	}
	if !attrs["a"].IsNull() || attrs["b"].IsNull() {
		t.Fatalf("got %v", attrs)
	}
}

func TestAMalformedTeamOrUserIDSaysHowToGetOne(t *testing.T) {
	cases := map[string]string{
		teamIDPattern:     "must be a team ID such as t123 (get it via List teams)",
		userIDPattern:     "must be a user ID such as u123, or an e-mail (get the user ID via List team members)",
		userIDOnlyPattern: "must be a user ID such as u123 (get it via List team members)",
		"^[a-z]+$":        "must match ^[a-z]+$",
	}
	for pattern, want := range cases {
		if got := idFormMessage(pattern); got != want {
			t.Errorf("idFormMessage(%q) = %q, want %q", pattern, got, want)
		}
	}
	// The patterns are the ones the schema emits, so a value is refused or kept exactly as the API does.
	for value, pattern := range map[string]string{"t552385": teamIDPattern, "u913": userIDPattern, "ana@example.com": userIDPattern} {
		if !regexp.MustCompile(pattern).MatchString(value) {
			t.Errorf("%q must match %s", value, pattern)
		}
	}
	for value, pattern := range map[string]string{"552385": teamIDPattern, "913": userIDPattern} {
		if regexp.MustCompile(pattern).MatchString(value) {
			t.Errorf("%q must not match %s", value, pattern)
		}
	}
}

func TestASpellingTheAPIReadsBackAsAnotherKeepsThePriorValue(t *testing.T) {
	methods := map[string]string{"GET": "get", "POST": "post"}
	if got := aliasedValue(methods, types.StringValue("GET"), "get"); got.ValueString() != "GET" {
		t.Fatalf("the spelling as written must stay: %v", got)
	}
	if got := aliasedValue(methods, types.StringValue("GET"), "post"); got.ValueString() != "post" {
		t.Fatalf("a changed value must refresh: %v", got)
	}
	if got := aliasedValue(methods, types.StringNull(), "get"); got.ValueString() != "get" {
		t.Fatalf("an import must read the API's value: %v", got)
	}
}

func TestAListReadBackWithMoreItemsKeepsThePriorValueWhileItHoldsThemAll(t *testing.T) {
	ids := func(v ...string) types.List {
		elems := make([]attr.Value, 0, len(v))
		for _, s := range v {
			elems = append(elems, types.StringValue(s))
		}
		return types.ListValueMust(types.StringType, elems)
	}
	if got := supersetRead(ids("2"), ids("1", "2", "3")); !got.Equal(ids("2")) {
		t.Fatalf("the listed ids must stay while the read holds them: %v", got)
	}
	if got := supersetRead(ids("2", "4"), ids("1", "2")); !got.Equal(ids("2")) {
		t.Fatalf("a listed id the read lacks must be planned again, without the extra ids: %v", got)
	}
	if got := supersetRead(ids("4"), ids()); !got.Equal(ids()) {
		t.Fatalf("an empty read must refresh: %v", got)
	}
	if got := supersetRead(ids(), ids("1")); !got.Equal(ids("1")) {
		t.Fatalf("an empty list holds no items, so a read with some must refresh: %v", got)
	}
	if got := supersetRead(types.ListNull(types.StringType), ids("1")); !got.Equal(ids("1")) {
		t.Fatalf("an import must read the API's list: %v", got)
	}
}

func TestAVRLProgramReadBackWithoutItsClosingLineKeepsThePriorValue(t *testing.T) {
	written := types.StringValue(".service = \"cart\"\n.")
	if got := vrlValue(written, ".service = \"cart\""); !got.Equal(written) {
		t.Fatalf("the program as written must stay: %v", got)
	}
	if got := vrlValue(types.StringValue(".a = 1"), ".a = 1\n."); got.ValueString() != ".a = 1" {
		t.Fatalf("the program as written must stay: %v", got)
	}
	if got := vrlValue(written, ".service = \"checkout\""); got.ValueString() != ".service = \"checkout\"" {
		t.Fatalf("a changed program must refresh: %v", got)
	}
	if got := vrlValue(types.StringNull(), ".a = 1"); got.ValueString() != ".a = 1" {
		t.Fatalf("an import must read the API's program: %v", got)
	}
}

func TestATimeZoneReadBackUnderItsOtherNameKeepsThePriorValue(t *testing.T) {
	written := types.StringValue("Europe/Prague")
	if got := timeZoneValue(written, "Prague"); !got.Equal(written) {
		t.Fatalf("the IANA ID as written must stay: %v", got)
	}
	if got := timeZoneValue(types.StringValue("Prague"), "Europe/Prague"); got.ValueString() != "Prague" {
		t.Fatalf("the Rails name as written must stay: %v", got)
	}
	if got := timeZoneValue(written, "Tokyo"); got.ValueString() != "Tokyo" {
		t.Fatalf("a changed zone must refresh: %v", got)
	}
	if got := timeZoneValue(types.StringNull(), "Prague"); got.ValueString() != "Prague" {
		t.Fatalf("an import must read the API's zone: %v", got)
	}
}
