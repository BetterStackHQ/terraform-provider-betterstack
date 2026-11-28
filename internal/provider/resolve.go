package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// configWithPlannedUnknowns keeps configured nulls as presence signals while replacing configured
// unknowns with plan values, falling back recursively to prior state when the plan is still unknown.
func configWithPlannedUnknowns(ctx context.Context, config tfsdk.Config, plan tfsdk.Plan, prior tftypes.Value, target interface{}) diag.Diagnostics {
	config.Raw = configuredFromPlan(config.Raw, plan.Raw, prior)
	return config.Get(ctx, target)
}

func configuredFromPlan(config, plan, prior tftypes.Value) tftypes.Value {
	return valueWithFallback(config, valueWithFallback(plan, prior))
}

func nullValue(t tftypes.Type) tftypes.Value {
	return tftypes.NewValue(t, nil)
}

// valueWithFallback replaces unknowns in value with the corresponding fallback values. Known
// containers recurse so unknown object attributes and list, set, tuple, or map elements resolve too.
func valueWithFallback(value, fallback tftypes.Value) tftypes.Value {
	if !value.IsKnown() {
		if fallback.IsKnown() {
			return fallback
		}
		return value
	}
	if value.IsNull() || !fallback.IsKnown() || fallback.IsNull() {
		return value
	}
	switch t := value.Type().(type) {
	case tftypes.Object:
		var values, fallbacks map[string]tftypes.Value
		if value.As(&values) != nil || fallback.As(&fallbacks) != nil {
			return value
		}
		out := make(map[string]tftypes.Value, len(values))
		for name, current := range values {
			fallback, ok := fallbacks[name]
			if !ok {
				fallback = nullValue(current.Type())
			}
			out[name] = valueWithFallback(current, fallback)
		}
		return tftypes.NewValue(t, out)
	case tftypes.List:
		return sequenceWithFallback(t, value, fallback)
	case tftypes.Set:
		return sequenceWithFallback(t, value, fallback)
	case tftypes.Tuple:
		var values, fallbacks []tftypes.Value
		if value.As(&values) != nil || fallback.As(&fallbacks) != nil {
			return value
		}
		out := make([]tftypes.Value, len(values))
		for i, current := range values {
			if i < len(fallbacks) {
				out[i] = valueWithFallback(current, fallbacks[i])
			} else {
				out[i] = current
			}
		}
		return tftypes.NewValue(t, out)
	case tftypes.Map:
		var values, fallbacks map[string]tftypes.Value
		if value.As(&values) != nil || fallback.As(&fallbacks) != nil {
			return value
		}
		out := make(map[string]tftypes.Value, len(values))
		for name, current := range values {
			fallback, ok := fallbacks[name]
			if !ok {
				fallback = nullValue(current.Type())
			}
			out[name] = valueWithFallback(current, fallback)
		}
		return tftypes.NewValue(t, out)
	default:
		return value
	}
}

func sequenceWithFallback(t tftypes.Type, value, fallback tftypes.Value) tftypes.Value {
	var values, fallbacks []tftypes.Value
	if value.As(&values) != nil || fallback.As(&fallbacks) != nil {
		return value
	}
	out := make([]tftypes.Value, len(values))
	for i, current := range values {
		if i < len(fallbacks) {
			out[i] = valueWithFallback(current, fallbacks[i])
		} else {
			out[i] = current
		}
	}
	return tftypes.NewValue(t, out)
}

// stateFromPlan produces the post-apply state: the planned value with every unknown resolved,
// from the API response where present, prior state where absent, and to null otherwise. Known values are kept
// verbatim so Terraform never sees an "inconsistent result after apply". This is what lets
// Optional+Computed attributes (unknown at plan time when unset) survive a real apply.
func stateFromPlan(plan, prior tftypes.Value, api map[string]interface{}) tftypes.Value {
	return resolveValue(plan.Type(), plan, prior, apiMapValue(api), api != nil)
}

func apiMapValue(api map[string]interface{}) interface{} {
	if api == nil {
		return nil
	}
	return map[string]interface{}(api)
}

// resolveValue walks a planned value: unknowns use the API response, then prior state, then null.
// Known values are kept, and containers recurse so nested unknowns (e.g. a
// computed sub-attribute inside a configured list element) resolve too.
func resolveValue(t tftypes.Type, v, prior tftypes.Value, api interface{}, apiPresent bool) tftypes.Value {
	if !v.IsKnown() {
		if apiPresent {
			return valueFromAPI(t, api)
		}
		if prior.IsKnown() {
			return prior
		}
		return nullValue(t)
	}
	if v.IsNull() {
		return v
	}
	switch tt := t.(type) {
	case tftypes.Object:
		var attrs map[string]tftypes.Value
		if err := v.As(&attrs); err != nil {
			return v
		}
		priorAttrs := map[string]tftypes.Value{}
		if prior.IsKnown() && !prior.IsNull() {
			_ = prior.As(&priorAttrs)
		}
		apiMap, _ := api.(map[string]interface{})
		out := make(map[string]tftypes.Value, len(tt.AttributeTypes))
		for name, at := range tt.AttributeTypes {
			sub, present := interface{}(nil), false
			if apiMap != nil {
				sub, present = apiField(apiMap, name)
			}
			priorValue, ok := priorAttrs[name]
			if !ok {
				priorValue = nullValue(at)
			}
			out[name] = resolveValue(at, attrs[name], priorValue, sub, present)
		}
		return tftypes.NewValue(t, out)
	case tftypes.List:
		var elems []tftypes.Value
		if err := v.As(&elems); err != nil {
			return v
		}
		var priorElems []tftypes.Value
		if prior.IsKnown() && !prior.IsNull() {
			_ = prior.As(&priorElems)
		}
		apiList, _ := api.([]interface{})
		out := make([]tftypes.Value, len(elems))
		for i, el := range elems {
			sub, present := interface{}(nil), false
			if i < len(apiList) {
				sub = apiList[i]
				present = true
			}
			priorValue := nullValue(tt.ElementType)
			if i < len(priorElems) {
				priorValue = priorElems[i]
			}
			out[i] = resolveValue(tt.ElementType, el, priorValue, sub, present)
		}
		return tftypes.NewValue(t, out)
	case tftypes.Set:
		var elems, priorElems []tftypes.Value
		if err := v.As(&elems); err != nil {
			return v
		}
		if prior.IsKnown() && !prior.IsNull() {
			_ = prior.As(&priorElems)
		}
		apiList, _ := api.([]interface{})
		out := make([]tftypes.Value, len(elems))
		for i, el := range elems {
			sub, present := interface{}(nil), false
			if i < len(apiList) {
				sub, present = apiList[i], true
			}
			priorValue := nullValue(tt.ElementType)
			if i < len(priorElems) {
				priorValue = priorElems[i]
			}
			out[i] = resolveValue(tt.ElementType, el, priorValue, sub, present)
		}
		return tftypes.NewValue(t, out)
	case tftypes.Map:
		var elems map[string]tftypes.Value
		if err := v.As(&elems); err != nil {
			return v
		}
		priorElems := map[string]tftypes.Value{}
		if prior.IsKnown() && !prior.IsNull() {
			_ = prior.As(&priorElems)
		}
		apiMap, _ := api.(map[string]interface{})
		out := make(map[string]tftypes.Value, len(elems))
		for k, el := range elems {
			sub, present := interface{}(nil), false
			if apiMap != nil {
				sub, present = apiMap[k]
			}
			priorValue, ok := priorElems[k]
			if !ok {
				priorValue = nullValue(tt.ElementType)
			}
			out[k] = resolveValue(tt.ElementType, el, priorValue, sub, present)
		}
		return tftypes.NewValue(t, out)
	case tftypes.Tuple:
		var elems []tftypes.Value
		if err := v.As(&elems); err != nil {
			return v
		}
		var priorElems []tftypes.Value
		if prior.IsKnown() && !prior.IsNull() {
			_ = prior.As(&priorElems)
		}
		apiList, _ := api.([]interface{})
		out := make([]tftypes.Value, len(elems))
		for i, el := range elems {
			sub, present := interface{}(nil), false
			if i < len(apiList) {
				sub = apiList[i]
				present = true
			}
			priorValue := nullValue(tt.ElementTypes[i])
			if i < len(priorElems) {
				priorValue = priorElems[i]
			}
			out[i] = resolveValue(tt.ElementTypes[i], el, priorValue, sub, present)
		}
		return tftypes.NewValue(t, out)
	default:
		return v
	}
}

// valueFromAPI builds a fully-known value of the requested type from a decoded API JSON value,
// coercing where the API is looser than the schema (stringified ids, numeric strings) and
// falling back to null on a shape mismatch. Numbers ride json.Number so precision survives.
func valueFromAPI(t tftypes.Type, api interface{}) tftypes.Value {
	if api == nil {
		return tftypes.NewValue(t, nil)
	}
	if t.Is(tftypes.DynamicPseudoType) {
		return dynamicFromAPI(api)
	}
	switch {
	case t.Is(tftypes.String):
		if s, ok := stringFromAPI(api); ok {
			return tftypes.NewValue(t, s)
		}
		return tftypes.NewValue(t, nil)
	case t.Is(tftypes.Number):
		if f, ok := numberFromAPI(api); ok {
			return tftypes.NewValue(t, f)
		}
		return tftypes.NewValue(t, nil)
	case t.Is(tftypes.Bool):
		if b, ok := api.(bool); ok {
			return tftypes.NewValue(t, b)
		}
		return tftypes.NewValue(t, nil)
	}
	switch tt := t.(type) {
	case tftypes.Object:
		apiMap, ok := api.(map[string]interface{})
		if !ok {
			return tftypes.NewValue(t, nil)
		}
		out := make(map[string]tftypes.Value, len(tt.AttributeTypes))
		for name, at := range tt.AttributeTypes {
			sub, _ := apiField(apiMap, name)
			out[name] = valueFromAPI(at, sub)
		}
		return tftypes.NewValue(t, out)
	case tftypes.List:
		apiList, ok := api.([]interface{})
		if !ok {
			return tftypes.NewValue(t, nil)
		}
		out := make([]tftypes.Value, 0, len(apiList))
		for _, e := range apiList {
			out = append(out, valueFromAPI(tt.ElementType, e))
		}
		return tftypes.NewValue(t, out)
	case tftypes.Set:
		apiList, ok := api.([]interface{})
		if !ok {
			return tftypes.NewValue(t, nil)
		}
		out := make([]tftypes.Value, 0, len(apiList))
		for _, e := range apiList {
			out = append(out, valueFromAPI(tt.ElementType, e))
		}
		return tftypes.NewValue(t, out)
	case tftypes.Map:
		apiMap, ok := api.(map[string]interface{})
		if !ok {
			return tftypes.NewValue(t, nil)
		}
		out := make(map[string]tftypes.Value, len(apiMap))
		for k, e := range apiMap {
			out[k] = valueFromAPI(tt.ElementType, e)
		}
		return tftypes.NewValue(t, out)
	}
	return tftypes.NewValue(t, nil)
}

// dynamicFromAPI infers a concrete Terraform value for a free-form API value (types.Dynamic
// attributes). Arrays become tuples so heterogeneous elements stay representable.
func dynamicFromAPI(api interface{}) tftypes.Value {
	switch v := api.(type) {
	case nil:
		return tftypes.NewValue(tftypes.DynamicPseudoType, nil)
	case string:
		return tftypes.NewValue(tftypes.String, v)
	case bool:
		return tftypes.NewValue(tftypes.Bool, v)
	case json.Number:
		if f, ok := numberFromAPI(v); ok {
			return tftypes.NewValue(tftypes.Number, f)
		}
		return tftypes.NewValue(tftypes.String, v.String())
	case float64:
		return tftypes.NewValue(tftypes.Number, new(big.Float).SetFloat64(v))
	case []interface{}:
		elems := make([]tftypes.Value, 0, len(v))
		types := make([]tftypes.Type, 0, len(v))
		for _, e := range v {
			ev := dynamicFromAPI(e)
			elems = append(elems, ev)
			types = append(types, ev.Type())
		}
		return tftypes.NewValue(tftypes.Tuple{ElementTypes: types}, elems)
	case map[string]interface{}:
		attrs := make(map[string]tftypes.Value, len(v))
		types := make(map[string]tftypes.Type, len(v))
		for k, e := range v {
			ev := dynamicFromAPI(e)
			attrs[k] = ev
			types[k] = ev.Type()
		}
		return tftypes.NewValue(tftypes.Object{AttributeTypes: types}, attrs)
	default:
		return tftypes.NewValue(tftypes.String, fmt.Sprintf("%v", v))
	}
}

// tftypesToAPI converts a known Terraform value back into a JSON-marshalable API value
// (the write path for types.Dynamic attributes).
func tftypesToAPI(v tftypes.Value) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}
	if !v.IsKnown() {
		return nil, errors.New("cannot serialize an unknown value")
	}
	t := v.Type()
	switch {
	case t.Is(tftypes.String):
		var s string
		if err := v.As(&s); err != nil {
			return nil, err
		}
		return s, nil
	case t.Is(tftypes.Bool):
		var b bool
		if err := v.As(&b); err != nil {
			return nil, err
		}
		return b, nil
	case t.Is(tftypes.Number):
		var f big.Float
		if err := v.As(&f); err != nil {
			return nil, err
		}
		return json.Number(f.Text('f', -1)), nil
	}
	switch tt := t.(type) {
	case tftypes.Object:
		var attrs map[string]tftypes.Value
		if err := v.As(&attrs); err != nil {
			return nil, err
		}
		out := make(map[string]interface{}, len(attrs))
		for k, av := range attrs {
			conv, err := tftypesToAPI(av)
			if err != nil {
				return nil, err
			}
			out[k] = conv
		}
		return out, nil
	case tftypes.Map:
		var elems map[string]tftypes.Value
		if err := v.As(&elems); err != nil {
			return nil, err
		}
		out := make(map[string]interface{}, len(elems))
		for k, ev := range elems {
			conv, err := tftypesToAPI(ev)
			if err != nil {
				return nil, err
			}
			out[k] = conv
		}
		return out, nil
	case tftypes.List, tftypes.Set, tftypes.Tuple:
		var elems []tftypes.Value
		if err := v.As(&elems); err != nil {
			return nil, err
		}
		_ = tt
		out := make([]interface{}, 0, len(elems))
		for _, ev := range elems {
			conv, err := tftypesToAPI(ev)
			if err != nil {
				return nil, err
			}
			out = append(out, conv)
		}
		return out, nil
	}
	return nil, fmt.Errorf("unsupported type %s", t)
}

func dynamicString(ctx context.Context, value types.Dynamic) string {
	if value.IsNull() || value.IsUnknown() || value.IsUnderlyingValueNull() || value.IsUnderlyingValueUnknown() {
		return ""
	}
	tfValue, err := value.UnderlyingValue().ToTerraformValue(ctx)
	if err != nil {
		return ""
	}
	api, err := tftypesToAPI(tfValue)
	if err != nil {
		return ""
	}
	return stringifyScalar(api)
}

// --- tolerant scalar coercions (the API stringifies some ids and numbers) ---------------------

// idToAPI sends an id the way the API takes it: Terraform holds every id as a string, and an API
// whose ids are numbers gets the number. An id that isn't a number (a Team's t123) goes as written.
func idToAPI(s string) interface{} {
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	return s
}

// A map whose names the API returns with null values (a monitor's environment variables): each
// returned name keeps the value state holds for it, so a read never plans a change to a configured
// value. A name state doesn't hold is null and a name the API no longer returns is gone, so a name
// added or removed outside Terraform still shows as drift.
func mapNamesKeepingValues(prior types.Map, raw map[string]interface{}) types.Map {
	known := map[string]attr.Value{}
	if !prior.IsNull() && !prior.IsUnknown() {
		known = prior.Elements()
	} else if len(raw) == 0 {
		return types.MapNull(types.StringType)
	}
	out := make(map[string]attr.Value, len(raw))
	for name, value := range raw {
		if before, ok := known[name]; ok {
			out[name] = before
		} else if s, ok := value.(string); ok {
			out[name] = types.StringValue(s)
		} else {
			out[name] = types.StringNull()
		}
	}
	return types.MapValueMust(types.StringType, out)
}

func stringFromAPI(v interface{}) (string, bool) {
	switch s := v.(type) {
	case string:
		return s, true
	case json.Number:
		return s.String(), true
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(s), true
	}
	return "", false
}

func int64FromAPI(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return i, true
		}
		if f, err := n.Float64(); err == nil {
			return int64(f), true
		}
	case float64:
		return int64(n), true
	case string:
		if i, err := strconv.ParseInt(n, 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

func float64FromAPI(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case json.Number:
		if f, err := n.Float64(); err == nil {
			return f, true
		}
	case float64:
		return n, true
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func boolFromAPI(v interface{}) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

func numberFromAPI(v interface{}) (*big.Float, bool) {
	switch n := v.(type) {
	case json.Number:
		if f, _, err := big.ParseFloat(n.String(), 10, 512, big.ToNearestEven); err == nil {
			return f, true
		}
	case float64:
		return new(big.Float).SetFloat64(n), true
	case string:
		if f, _, err := big.ParseFloat(n, 10, 512, big.ToNearestEven); err == nil {
			return f, true
		}
	}
	return nil, false
}

// compactJSON renders an API value as canonical compact JSON (sorted keys, no HTML escaping,
// exact numbers) — byte-compatible with Terraform's jsonencode(), so JSON-string attributes
// refresh without phantom diffs.
func compactJSON(v interface{}) (string, bool) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", false
	}
	return strings.TrimSuffix(buf.String(), "\n"), true
}

// jsonStringFromAPI renders an API value for a JSON-string attribute. A string that itself parses
// as JSON is compacted rather than re-quoted (APIs storing JSON documents as text columns return
// them as strings); other strings pass through, containers compact like jsonencode().
func jsonStringFromAPI(v interface{}) (string, bool) {
	if s, ok := v.(string); ok {
		if json.Valid([]byte(s)) {
			dec := json.NewDecoder(strings.NewReader(s))
			dec.UseNumber()
			var decoded interface{}
			if dec.Decode(&decoded) == nil {
				return compactJSON(decoded)
			}
		}
		return s, true
	}
	return compactJSON(v)
}

// apiField is an API object's value for a Terraform attribute name. A wire key in camelCase
// (canonicalSourceIds) is snake_case in Terraform (canonical_source_ids), so that spelling matches too.
func apiField(m map[string]interface{}, name string) (interface{}, bool) {
	if v, ok := m[name]; ok {
		return v, true
	}
	for key, v := range m {
		if strings.ContainsAny(key, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") && snakeCase(key) == name {
			return v, true
		}
	}
	return nil, false
}

func snakeCase(key string) string {
	var b strings.Builder
	for i, r := range key {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := rune(key[i-1])
				if (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9') {
					b.WriteByte('_')
				}
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// keyedAPIRead leaves out of a read list the items the API holds by itself (ignored), unless the
// prior value names them, and puts the rest in the prior value's order by key. So built-in items and
// the API's own order never show as a change. Items the prior value doesn't name follow in API order.
func keyedAPIRead(raw []interface{}, key string, prior []string, priorComplete bool, ignored []string) []interface{} {
	keep := map[string]bool{}
	for _, k := range prior {
		keep[k] = true
	}
	skip := map[string]bool{}
	for _, k := range ignored {
		if !keep[k] {
			skip[k] = true
		}
	}
	var kept []interface{}
	byKey := map[string][]int{}
	for _, item := range raw {
		object, ok := item.(map[string]interface{})
		k := ""
		if ok {
			k = stringifyScalar(object[key])
		}
		if ok && skip[k] {
			continue
		}
		if ok && keep[k] {
			byKey[k] = append(byKey[k], len(kept))
		}
		kept = append(kept, item)
	}
	// A planned item whose key isn't known yet can't be placed, so the read keeps the API's order.
	if !priorComplete {
		return kept
	}
	out := make([]interface{}, 0, len(kept))
	placed := make([]bool, len(kept))
	for _, k := range prior {
		if indexes := byKey[k]; len(indexes) > 0 {
			out = append(out, kept[indexes[0]])
			placed[indexes[0]] = true
			byKey[k] = indexes[1:]
		}
	}
	// Everything else, a second item with a configured key included, follows in the API's order.
	for i, item := range kept {
		if !placed[i] {
			out = append(out, item)
		}
	}
	return out
}

// sameJSON reports whether two JSON texts hold the same value, so a JSON-string attribute keeps
// the configured spelling (its spacing, its key order) when the API returns the value compacted.
func sameJSON(a, b string) bool {
	decode := func(text string) (interface{}, bool) {
		dec := json.NewDecoder(strings.NewReader(text))
		dec.UseNumber()
		var v interface{}
		return v, dec.Decode(&v) == nil
	}
	av, aok := decode(a)
	bv, bok := decode(b)
	return aok && bok && reflect.DeepEqual(av, bv)
}

// withoutAttributes is an object value with the named attributes null, so resolving a plan against it
// takes those attributes from the API response alone, never from the prior state.
func withoutAttributes(v tftypes.Value, names ...string) tftypes.Value {
	if !v.IsKnown() || v.IsNull() {
		return v
	}
	var attrs map[string]tftypes.Value
	if err := v.As(&attrs); err != nil {
		return v
	}
	ot, ok := v.Type().(tftypes.Object)
	if !ok {
		return v
	}
	for _, name := range names {
		if at, exists := ot.AttributeTypes[name]; exists {
			attrs[name] = nullValue(at)
		}
	}
	return tftypes.NewValue(v.Type(), attrs)
}

// emptyConfiguredObject plans an object configured with no field set as exactly that, every field
// null, instead of keeping the fields' current values: `{}` stands for the default the API reads back
// as a string (an alert's `current_team`), so it is a change the write sends as {}.
type emptyConfiguredObject struct{}

func (emptyConfiguredObject) Description(context.Context) string {
	return "An object configured with no field set plans every field empty."
}

func (m emptyConfiguredObject) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (emptyConfiguredObject) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	for _, value := range req.ConfigValue.Attributes() {
		if !value.IsNull() {
			return
		}
	}
	resp.PlanValue = req.ConfigValue
}

// stringifyScalar renders an API scalar exactly (json.Number keeps its digits, so ids never
// collapse into scientific notation); container values fall back to compact JSON.
func stringifyScalar(v interface{}) string {
	switch s := v.(type) {
	case nil:
		return ""
	case string:
		return s
	case json.Number:
		return s.String()
	case bool:
		return strconv.FormatBool(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(raw)
	}
}

// datetimeValue refreshes an RFC 3339 attribute from the API. When the API writes the instant the
// prior value already holds in another form ("2026-11-23T00:00:00Z" read back as
// "2026-11-23T00:00:00.000Z"), the prior value stays, so a refresh never plans a change, or a
// replacement, the user didn't make. It is the semantic equality HashiCorp's timetypes.RFC3339
// gives, kept to the refresh so every datetime stays a plain string on every surface.
//
// Otherwise the API's instant is stored in its shortest RFC 3339 form, keeping the API's offset:
// "2026-11-23T00:00:00.000Z" becomes "2026-11-23T00:00:00Z". An import has no prior value, and
// Terraform requires a configured attribute to plan exactly as written, so this is the form that
// lets a configuration written the usual way plan clean after an import.
func datetimeValue(prior types.String, v string) types.String {
	after, errAfter := time.Parse(time.RFC3339Nano, v)
	if !prior.IsNull() && !prior.IsUnknown() && prior.ValueString() != v {
		before, errBefore := time.Parse(time.RFC3339Nano, prior.ValueString())
		if errBefore == nil && errAfter == nil && before.Equal(after) {
			return prior
		}
	}
	if errAfter == nil {
		return types.StringValue(after.Format(time.RFC3339Nano))
	}
	return types.StringValue(v)
}

// vrlValue refreshes a VRL program from the API. A program's closing `.` line (the event it
// returns) changes nothing it does, and an app may store it with or without that line, so when the
// API's program differs from the prior value only by it, the prior value stays and a refresh never
// plans a change the user didn't make.
func vrlValue(prior types.String, v string) types.String {
	if !prior.IsNull() && !prior.IsUnknown() && prior.ValueString() != v && vrlBody(prior.ValueString()) == vrlBody(v) {
		return prior
	}
	return types.StringValue(v)
}

// timeZoneValue refreshes a time zone from the API. The app stores a Rails name ("Prague") for an
// IANA ID ("Europe/Prague") it was sent, and the reverse, so when the API's zone is the prior value
// under its other name, the prior value stays and a refresh never plans a change the user didn't make.
func timeZoneValue(prior types.String, v string) types.String {
	if !prior.IsNull() && !prior.IsUnknown() && prior.ValueString() != v && sameTimeZone(prior.ValueString(), v) {
		return prior
	}
	return types.StringValue(v)
}

// keepTimeZoneName plans the zone in state when the configuration names the same zone under its
// other name, as after an import, where the read returns the Rails name and the configuration may
// hold the IANA ID. Terraform accepts the prior value as the plan for an equivalent configuration,
// so the first plan shows no change the API wouldn't make.
type keepTimeZoneName struct{}

func (keepTimeZoneName) Description(context.Context) string {
	return "A time zone written under its other name plans no change."
}

func (m keepTimeZoneName) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }

func (keepTimeZoneName) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueString() != req.StateValue.ValueString() && sameTimeZone(req.ConfigValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

// sameTimeZone reports whether two spellings name one zone: a Rails name is read as the IANA ID
// Rails maps it to (ActiveSupport::TimeZone::MAPPING), and an IANA ID as itself.
func sameTimeZone(a, b string) bool {
	ianaOf := func(zone string) string {
		if iana, ok := railsTimeZones[zone]; ok {
			return iana
		}
		return zone
	}
	return ianaOf(a) == ianaOf(b)
}

// railsTimeZones is ActiveSupport::TimeZone::MAPPING (activesupport 8.1.3): each Rails time zone
// name and the IANA ID it stands for.
var railsTimeZones = map[string]string{
	"International Date Line West": "Etc/GMT+12",
	"Midway Island":                "Pacific/Midway",
	"American Samoa":               "Pacific/Pago_Pago",
	"Hawaii":                       "Pacific/Honolulu",
	"Alaska":                       "America/Juneau",
	"Pacific Time (US & Canada)":   "America/Los_Angeles",
	"Tijuana":                      "America/Tijuana",
	"Mountain Time (US & Canada)":  "America/Denver",
	"Arizona":                      "America/Phoenix",
	"Chihuahua":                    "America/Chihuahua",
	"Mazatlan":                     "America/Mazatlan",
	"Central Time (US & Canada)":   "America/Chicago",
	"Saskatchewan":                 "America/Regina",
	"Guadalajara":                  "America/Mexico_City",
	"Mexico City":                  "America/Mexico_City",
	"Monterrey":                    "America/Monterrey",
	"Central America":              "America/Guatemala",
	"Eastern Time (US & Canada)":   "America/New_York",
	"Indiana (East)":               "America/Indiana/Indianapolis",
	"Bogota":                       "America/Bogota",
	"Lima":                         "America/Lima",
	"Quito":                        "America/Lima",
	"Atlantic Time (Canada)":       "America/Halifax",
	"Caracas":                      "America/Caracas",
	"La Paz":                       "America/La_Paz",
	"Santiago":                     "America/Santiago",
	"Asuncion":                     "America/Asuncion",
	"Newfoundland":                 "America/St_Johns",
	"Brasilia":                     "America/Sao_Paulo",
	"Buenos Aires":                 "America/Argentina/Buenos_Aires",
	"Montevideo":                   "America/Montevideo",
	"Georgetown":                   "America/Guyana",
	"Puerto Rico":                  "America/Puerto_Rico",
	"Greenland":                    "America/Nuuk",
	"Mid-Atlantic":                 "Atlantic/South_Georgia",
	"Azores":                       "Atlantic/Azores",
	"Cape Verde Is.":               "Atlantic/Cape_Verde",
	"Dublin":                       "Europe/Dublin",
	"Edinburgh":                    "Europe/London",
	"Lisbon":                       "Europe/Lisbon",
	"London":                       "Europe/London",
	"Casablanca":                   "Africa/Casablanca",
	"Monrovia":                     "Africa/Monrovia",
	"UTC":                          "Etc/UTC",
	"Belgrade":                     "Europe/Belgrade",
	"Bratislava":                   "Europe/Bratislava",
	"Budapest":                     "Europe/Budapest",
	"Ljubljana":                    "Europe/Ljubljana",
	"Prague":                       "Europe/Prague",
	"Sarajevo":                     "Europe/Sarajevo",
	"Skopje":                       "Europe/Skopje",
	"Warsaw":                       "Europe/Warsaw",
	"Zagreb":                       "Europe/Zagreb",
	"Brussels":                     "Europe/Brussels",
	"Copenhagen":                   "Europe/Copenhagen",
	"Madrid":                       "Europe/Madrid",
	"Paris":                        "Europe/Paris",
	"Amsterdam":                    "Europe/Amsterdam",
	"Berlin":                       "Europe/Berlin",
	"Bern":                         "Europe/Zurich",
	"Zurich":                       "Europe/Zurich",
	"Rome":                         "Europe/Rome",
	"Stockholm":                    "Europe/Stockholm",
	"Vienna":                       "Europe/Vienna",
	"West Central Africa":          "Africa/Algiers",
	"Bucharest":                    "Europe/Bucharest",
	"Cairo":                        "Africa/Cairo",
	"Helsinki":                     "Europe/Helsinki",
	"Kyiv":                         "Europe/Kiev",
	"Riga":                         "Europe/Riga",
	"Sofia":                        "Europe/Sofia",
	"Tallinn":                      "Europe/Tallinn",
	"Vilnius":                      "Europe/Vilnius",
	"Athens":                       "Europe/Athens",
	"Istanbul":                     "Europe/Istanbul",
	"Minsk":                        "Europe/Minsk",
	"Jerusalem":                    "Asia/Jerusalem",
	"Harare":                       "Africa/Harare",
	"Pretoria":                     "Africa/Johannesburg",
	"Kaliningrad":                  "Europe/Kaliningrad",
	"Moscow":                       "Europe/Moscow",
	"St. Petersburg":               "Europe/Moscow",
	"Volgograd":                    "Europe/Volgograd",
	"Samara":                       "Europe/Samara",
	"Kuwait":                       "Asia/Kuwait",
	"Riyadh":                       "Asia/Riyadh",
	"Nairobi":                      "Africa/Nairobi",
	"Baghdad":                      "Asia/Baghdad",
	"Tehran":                       "Asia/Tehran",
	"Abu Dhabi":                    "Asia/Muscat",
	"Muscat":                       "Asia/Muscat",
	"Baku":                         "Asia/Baku",
	"Tbilisi":                      "Asia/Tbilisi",
	"Yerevan":                      "Asia/Yerevan",
	"Kabul":                        "Asia/Kabul",
	"Ekaterinburg":                 "Asia/Yekaterinburg",
	"Islamabad":                    "Asia/Karachi",
	"Karachi":                      "Asia/Karachi",
	"Tashkent":                     "Asia/Tashkent",
	"Chennai":                      "Asia/Kolkata",
	"Kolkata":                      "Asia/Kolkata",
	"Mumbai":                       "Asia/Kolkata",
	"New Delhi":                    "Asia/Kolkata",
	"Kathmandu":                    "Asia/Kathmandu",
	"Dhaka":                        "Asia/Dhaka",
	"Sri Jayawardenepura":          "Asia/Colombo",
	"Almaty":                       "Asia/Almaty",
	"Astana":                       "Asia/Almaty",
	"Novosibirsk":                  "Asia/Novosibirsk",
	"Rangoon":                      "Asia/Rangoon",
	"Bangkok":                      "Asia/Bangkok",
	"Hanoi":                        "Asia/Bangkok",
	"Jakarta":                      "Asia/Jakarta",
	"Krasnoyarsk":                  "Asia/Krasnoyarsk",
	"Beijing":                      "Asia/Shanghai",
	"Chongqing":                    "Asia/Chongqing",
	"Hong Kong":                    "Asia/Hong_Kong",
	"Urumqi":                       "Asia/Urumqi",
	"Kuala Lumpur":                 "Asia/Kuala_Lumpur",
	"Singapore":                    "Asia/Singapore",
	"Taipei":                       "Asia/Taipei",
	"Perth":                        "Australia/Perth",
	"Irkutsk":                      "Asia/Irkutsk",
	"Ulaanbaatar":                  "Asia/Ulaanbaatar",
	"Seoul":                        "Asia/Seoul",
	"Osaka":                        "Asia/Tokyo",
	"Sapporo":                      "Asia/Tokyo",
	"Tokyo":                        "Asia/Tokyo",
	"Yakutsk":                      "Asia/Yakutsk",
	"Darwin":                       "Australia/Darwin",
	"Adelaide":                     "Australia/Adelaide",
	"Canberra":                     "Australia/Canberra",
	"Melbourne":                    "Australia/Melbourne",
	"Sydney":                       "Australia/Sydney",
	"Brisbane":                     "Australia/Brisbane",
	"Hobart":                       "Australia/Hobart",
	"Vladivostok":                  "Asia/Vladivostok",
	"Guam":                         "Pacific/Guam",
	"Port Moresby":                 "Pacific/Port_Moresby",
	"Magadan":                      "Asia/Magadan",
	"Srednekolymsk":                "Asia/Srednekolymsk",
	"Solomon Is.":                  "Pacific/Guadalcanal",
	"New Caledonia":                "Pacific/Noumea",
	"Fiji":                         "Pacific/Fiji",
	"Kamchatka":                    "Asia/Kamchatka",
	"Marshall Is.":                 "Pacific/Majuro",
	"Auckland":                     "Pacific/Auckland",
	"Wellington":                   "Pacific/Auckland",
	"Nuku'alofa":                   "Pacific/Tongatapu",
	"Tokelau Is.":                  "Pacific/Fakaofo",
	"Chatham Is.":                  "Pacific/Chatham",
	"Samoa":                        "Pacific/Apia",
}

// supersetRead refreshes a list the API reads back with more than was written: a source joins the
// whole connected set of each source it lists, so the read can hold others too. While the read holds
// every prior item, the prior value stays, so neither the extra items nor their order plan a change.
// A prior item missing from the read leaves only the prior items the read still holds, so the plan
// adds the missing one back without listing the extra items as removed, which an update can't do.
// An empty prior value stays only while the read is empty, since an empty list means no items at all.
func supersetRead[T interface {
	attr.Value
	Elements() []attr.Value
}](prior, read T) T {
	if prior.IsNull() || prior.IsUnknown() || len(prior.Elements()) == 0 {
		return read
	}
	missing := false
	for _, want := range prior.Elements() {
		if !holdsValue(read.Elements(), want) {
			missing = true
			break
		}
	}
	if !missing {
		return prior
	}
	if len(read.Elements()) == 0 {
		return read
	}
	kept := []attr.Value{}
	for _, want := range prior.Elements() {
		if holdsValue(read.Elements(), want) {
			kept = append(kept, want)
		}
	}
	var out attr.Value
	switch v := any(read).(type) {
	case types.List:
		out = types.ListValueMust(v.ElementType(context.Background()), kept)
	case types.Set:
		out = types.SetValueMust(v.ElementType(context.Background()), kept)
	default:
		return read
	}
	return out.(T)
}

func holdsValue(values []attr.Value, want attr.Value) bool {
	for _, got := range values {
		if want.Equal(got) {
			return true
		}
	}
	return false
}

// configuredBefore reports whether the last apply configured name (rememberConfigured's record). A
// list the API reads back with more than was written (supersetRead) compares the read with the
// configured list only then: unset, state is the API's whole list, so a refresh reads it all.
func configuredBefore(ctx context.Context, private privateGetter, name string) bool {
	if private == nil {
		return false
	}
	raw, d := private.GetKey(ctx, configuredKey)
	if d.HasError() || len(raw) == 0 {
		return false
	}
	var applied []string
	if json.Unmarshal(raw, &applied) != nil {
		return false
	}
	for _, n := range applied {
		if n == name {
			return true
		}
	}
	return false
}

// warnSupersetRemoval warns, for each of names whose configured list drops an item state holds, that
// Better Stack keeps it: an update only adds to the list (a source keeps every connection it has), so
// the plan would show the item gone while it stays. An empty list removes them all.
func warnSupersetRemoval(ctx context.Context, resourceType string, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, names []string) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	for _, name := range names {
		var config, state attr.Value
		if d := req.Config.GetAttribute(ctx, path.Root(name), &config); d.HasError() || config == nil || config.IsNull() || config.IsUnknown() {
			continue
		}
		if d := req.State.GetAttribute(ctx, path.Root(name), &state); d.HasError() || state == nil || state.IsNull() || state.IsUnknown() {
			continue
		}
		configured, ok1 := config.(interface{ Elements() []attr.Value })
		stored, ok2 := state.(interface{ Elements() []attr.Value })
		if !ok1 || !ok2 || len(configured.Elements()) == 0 {
			continue
		}
		dropped := false
		for _, item := range stored.Elements() {
			if !holdsValue(configured.Elements(), item) {
				dropped = true
				break
			}
		}
		if !dropped {
			continue
		}
		p := path.Root(name)
		if pulumiNames != nil {
			resp.Diagnostics.AddWarning("", "Taking items out of "+pulumiNames(resourceType, p)+" doesn't remove them in Better Stack: an update only adds to it. Set it to an empty list to remove them all, then list the ones to keep.")
			continue
		}
		resp.Diagnostics.AddAttributeWarning(p, "Taking items out of "+name+" doesn't remove them",
			"An update only adds to "+name+", so Better Stack keeps the items this configuration no longer lists. Set it to [] and apply to remove them all, then list the ones to keep.")
	}
}

// vrlBody is a VRL program without surrounding whitespace and its closing `.` line.
func vrlBody(program string) string {
	body := strings.TrimSpace(program)
	if body == "." {
		return ""
	}
	return strings.TrimSpace(strings.TrimSuffix(body, "\n."))
}

// addAPIError shows every sentence of an API failure. One that names a top-level attribute of
// schema (the object type of the plan, state or config at hand; wire keys are attribute names)
// lands on that attribute, so Terraform points at the offending line. The rest, such as a record-
// level message or a nested key, stay together in one resource-level error. A failure with no
// per-attribute sentence stays a single resource-level error carrying the request and message.
func addAPIError(diags *diag.Diagnostics, summary string, err error, schema tftypes.Type) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		diags.AddError(summary, err.Error())
		return
	}
	entries := apiErr.Entries()
	if len(entries) == 0 {
		diags.AddError(summary, err.Error())
		return
	}
	attrs := map[string]tftypes.Type{}
	if obj, ok := schema.(tftypes.Object); ok {
		attrs = obj.AttributeTypes
	}
	var rest []string
	for _, entry := range entries {
		if _, ok := attrs[entry.Attr]; ok {
			diags.AddAttributeError(path.Root(entry.Attr), summary, entry.Text)
			continue
		}
		rest = append(rest, entry.Text)
	}
	if len(rest) > 0 {
		diags.AddError(summary, strings.Join(rest, "\n"))
	}
}

// --- replacement warnings ---------------------------------------------------------------------

// protectedLabel reads deletion_protection from prior state: ok is true only while it is on, and
// label names the object (`the source "prod-logs"`, or `"prod-logs"` without a noun) by its name
// where it has one.
// sameAliasedValue reports whether two spellings read back as the same value, by the schema's
// aliases table (accepted spelling -> value it reads back as).
func sameAliasedValue(aliases map[string]string, a, b string) bool {
	canonical := func(v string) string {
		if to, ok := aliases[v]; ok {
			return to
		}
		return v
	}
	return canonical(a) == canonical(b)
}

// aliasedValue refreshes a string the API reads back in its own spelling (`GET` as `get`): when the
// API's value and the prior value are one value by the aliases table, the prior value stays, so a
// refresh never plans a change the user didn't make.
func aliasedValue(aliases map[string]string, prior types.String, v string) types.String {
	if !prior.IsNull() && !prior.IsUnknown() && sameAliasedValue(aliases, prior.ValueString(), v) {
		return prior
	}
	return types.StringValue(v)
}

func protectedLabel(ctx context.Context, state tfsdk.State, noun string) (label string, ok bool) {
	if state.Schema == nil || state.Raw.IsNull() {
		return "", false
	}
	var protected types.Bool
	if diags := state.GetAttribute(ctx, path.Root("deletion_protection"), &protected); diags.HasError() || !protected.ValueBool() {
		return "", false
	}
	// The object is named as its UI names it: its name, or the attribute that stands for one (a
	// monitor's pronounceable_name, a status page's company_name). An object with none is named by its
	// id when that is a key people know it by (an e-mail).
	var name types.String
	for _, attr := range []string{"name", "pronounceable_name", "company_name"} {
		var v types.String
		if diags := state.GetAttribute(ctx, path.Root(attr), &v); !diags.HasError() && v.ValueString() != "" {
			name = v
			break
		}
	}
	if name.ValueString() == "" {
		var id types.String
		if diags := state.GetAttribute(ctx, path.Root("id"), &id); !diags.HasError() && strings.Contains(id.ValueString(), "@") {
			name = id
		}
	}
	if name.ValueString() != "" {
		if noun == "" {
			return fmt.Sprintf("%q", name.ValueString()), true
		}
		return fmt.Sprintf("%s %q", noun, name.ValueString()), true
	}
	if noun == "" {
		return "this object", true
	}
	return noun, true
}

// protectedMessage is the refusal people see when a plan would destroy a protected object, by a
// destroy, a removed resource or a replacement: one sentence, in the words of each tool's users.
func protectedMessage(label string) string {
	if pulumiNames != nil {
		return "This update destroys " + label + " which is an irreversible action. Set deletionProtection: false and run pulumi up to allow destroying it."
	}
	return "This plan destroys " + label + " which is an irreversible action. Set deletion_protection = false and apply to allow destroying it."
}

// pulumiNames is set when the Pulumi bridge builds this provider (shim.NewForPulumi): it spells a
// resource type's attribute path as a Pulumi program writes it (ipAllowlists, not ip_allowlist).
// One plugin process serves one tool, so the messages below speak that tool's terms: what its
// replacement does first, and the fixes its users have.
var pulumiNames func(resourceType string, p path.Path) string

// UsePulumiMessages switches the plan's replacement and deletion-protection messages to Pulumi's.
func UsePulumiMessages(names func(resourceType string, p path.Path) string) { pulumiNames = names }

// pulumiSchemaDocs rewrites a schema's attribute descriptions for Pulumi, where the bridge reads
// them (QuoteAPIValues says why). Every attribute type has a Description field and nested ones an
// Attributes map, directly or under NestedObject, so one reflective walk covers resources and data
// sources alike.
// pulumiDescriptionEdit, set by the Pulumi provider, rewrites the Terraform-only sentences of a
// description before its values are quoted.
var pulumiDescriptionEdit func(string) string

// pulumiDeletesFirst names the resource types Pulumi replaces by deleting first (DeleteBeforeReplace),
// so the replacement warning says the order Pulumi really uses.
var pulumiDeletesFirst = map[string]bool{}

// SetPulumiDeletesFirst installs pulumiDeletesFirst.
func SetPulumiDeletesFirst(types []string) {
	for _, t := range types {
		pulumiDeletesFirst[t] = true
	}
}

// SetPulumiDescriptionEdit installs pulumiDescriptionEdit.
func SetPulumiDescriptionEdit(edit func(string) string) { pulumiDescriptionEdit = edit }

func pulumiSchemaDocs[A any](attrs map[string]A) {
	if pulumiNames == nil {
		return
	}
	names := map[string]bool{}
	collectAttributeNames(reflect.ValueOf(attrs), names)
	rewriteDescriptions(reflect.ValueOf(attrs), names)
}

func nestedAttributes(v reflect.Value) (reflect.Value, bool) {
	if f := v.FieldByName("Attributes"); f.IsValid() && f.Kind() == reflect.Map {
		return f, true
	}
	if n := v.FieldByName("NestedObject"); n.IsValid() {
		if f := n.FieldByName("Attributes"); f.IsValid() && f.Kind() == reflect.Map {
			return f, true
		}
	}
	return reflect.Value{}, false
}

func collectAttributeNames(attrs reflect.Value, names map[string]bool) {
	for _, key := range attrs.MapKeys() {
		names[key.String()] = true
		if nested, ok := nestedAttributes(attrs.MapIndex(key).Elem()); ok {
			collectAttributeNames(nested, names)
		}
	}
}

func rewriteDescriptions(attrs reflect.Value, names map[string]bool) {
	for _, key := range attrs.MapKeys() {
		v := attrs.MapIndex(key).Elem()
		attr := reflect.New(v.Type()).Elem()
		attr.Set(v)
		// The bridge reads MarkdownDescription where there is one, so both get the same rewrite.
		for _, field := range []string{"Description", "MarkdownDescription"} {
			if d := attr.FieldByName(field); d.IsValid() && d.Kind() == reflect.String && d.String() != "" {
				text := d.String()
				if pulumiDescriptionEdit != nil {
					text = pulumiDescriptionEdit(text)
				}
				d.SetString(QuoteDescriptionValues(text, names))
			}
		}
		if nested, ok := nestedAttributes(attr); ok {
			rewriteDescriptions(nested, names)
		}
		attrs.SetMapIndex(key, attr)
	}
}

// The bridge renames a lowercase word in backticks, and a snake_case word in double quotes, as if
// it were a property: `react_errors` reads reactErrors in TypeScript and Go, `europe` reads Europe
// in C#. Right for an attribute, wrong for a value the API takes, which Terraform's text writes in
// backticks too.
var (
	backtickedWord    = regexp.MustCompile("`([0-9a-z]+(?:_[0-9a-z]+)*)`")
	doubleQuotedSnake = regexp.MustCompile(`"([0-9a-z]+(?:_[0-9a-z]+)+)"`)
	codeFence         = regexp.MustCompile("(?ms)^```.*?^```[^\n]*$")
)

// QuoteAPIValues writes Markdown for the Pulumi bridge: a word that isn't one of names goes in
// single quotes, and true, false and null go bare. Code blocks are left alone. On an attribute line
// of a docs page it escapes a quoted value's underscores: the bridge's Markdown pass otherwise
// pairs them across a list of values ('react_errors', 'vue_errors' became 'react*errors', ...).
func QuoteAPIValues(md string, names map[string]bool) string { return quoteAPIValues(md, names, true) }

// QuoteDescriptionValues quotes a schema description's values without escaping them: no Markdown
// pass reads a description, so an escape would reach every SDK's docs as a backslash.
func QuoteDescriptionValues(md string, names map[string]bool) string {
	return quoteAPIValues(md, names, false)
}

func quoteAPIValues(md string, names map[string]bool, escape bool) string {
	var b strings.Builder
	last := 0
	for _, loc := range codeFence.FindAllStringIndex(md, -1) {
		b.WriteString(quoteProse(md[last:loc[0]], names, escape))
		b.WriteString(md[loc[0]:loc[1]])
		last = loc[1]
	}
	b.WriteString(quoteProse(md[last:], names, escape))
	return b.String()
}

// valueList is a run of text that lists only values, whatever their spelling: a monitor's
// "Valid values: `status`, `keyword`" names its monitor_type values, not its status attribute.
var valueList = regexp.MustCompile("Valid values: [^\n]*?\\.(?:\\s|$)|Defaults to `[^`]*`")

func quoteProse(text string, names map[string]bool, escape bool) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = valueList.ReplaceAllStringFunc(line, func(list string) string { return quoteLine(list, nil, escape && strings.HasPrefix(lines[i], "- `")) })
		lines[i] = quoteLine(line, names, escape && strings.HasPrefix(line, "- `"))
	}
	return strings.Join(lines, "\n")
}

func quoteLine(line string, names map[string]bool, attrLine bool) string {
	for _, re := range []*regexp.Regexp{backtickedWord, doubleQuotedSnake} {
		var b strings.Builder
		last := 0
		for _, loc := range re.FindAllStringSubmatchIndex(line, -1) {
			b.WriteString(line[last:loc[0]])
			m, word := line[loc[0]:loc[1]], line[loc[2]:loc[3]]
			// Markdown keeps a code span's text as written, so an escape inside one would show.
			inCode := strings.Count(line[:loc[0]], "`")%2 == 1
			switch {
			case names[word]:
				b.WriteString(m)
			case word == "true" || word == "false" || word == "null":
				b.WriteString(word)
			case attrLine && !inCode:
				b.WriteString("'" + strings.ReplaceAll(word, "_", `\_`) + "'")
			default:
				b.WriteString("'" + word + "'")
			}
			last = loc[1]
		}
		b.WriteString(line[last:])
		line = b.String()
	}
	return line
}

// refuseProtectedDestroy fails a plan that destroys an object whose state has
// deletion_protection = true. The API refuses that delete as well, but only once the apply is
// under way, after the plan's other changes have been made. Protection is turned off by its own
// apply first, as with Google's and AWS's databases. A replacement is refused per attribute, in
// refuseProtectedReplace, because a resource's ModifyPlan isn't told which attributes force one.
func refuseProtectedDestroy(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, noun string) {
	if !req.Plan.Raw.IsNull() {
		return
	}
	if label, ok := protectedLabel(ctx, req.State, noun); ok {
		addProtectedError(&resp.Diagnostics, "Deletion protection is on for "+label, label)
	}
}

// refuseUnnamedTeam fails the plan of an object that is in another team than the provider's while
// its configuration names no team_id. Unset, the plan keeps the team the object was read in, but a
// replacement or a re-creation plans with no prior state, so it would create the object in the
// provider's team and delete it from its own. Naming the team keeps both in place.
func refuseUnnamedTeam(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, client *Client, noun string) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() || client == nil {
		return
	}
	provider := client.tenants["team_id"]
	var configured, current types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("team_id"), &configured)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("team_id"), &current)...)
	if provider == "" || !configured.IsNull() || current.IsNull() || current.IsUnknown() || current.ValueString() == "" || current.ValueString() == provider {
		return
	}
	resp.Diagnostics.AddAttributeError(path.Root("team_id"), "The "+noun+" is in another team than the provider's",
		fmt.Sprintf("It is in team %s, and the provider's team_id is %s. Set team_id = %q on it, so that a replacement or a re-creation stays in team %s instead of moving to %s.", current.ValueString(), provider, current.ValueString(), current.ValueString(), provider))
}

// refuseProtectedDelete stops a delete of an object whose state still has deletion_protection = true.
// The plan checks above never see a pulumi destroy, a terraform apply -replace or a tainted object,
// so the delete itself checks too, before it sends force=true.
func refuseProtectedDelete(ctx context.Context, state tfsdk.State, noun string, diags *diag.Diagnostics) bool {
	label, ok := protectedLabel(ctx, state, noun)
	if ok {
		addProtectedError(diags, "Deletion protection is on for "+label+", so it was not deleted", label)
	}
	return ok
}

// addProtectedError adds the deletion protection refusal. The Pulumi bridge both logs an error as
// "detail: summary" and fails the operation with "summary: detail", so people read it twice in
// opposite orders. Under Pulumi the whole message is both summary and detail: the operation fails
// with it once, and the provider's log sink drops the logged copy (pulumi/provider says how).
func addProtectedError(diags *diag.Diagnostics, summary, label string) {
	if pulumiNames != nil {
		msg := summary + ". " + protectedMessage(label)
		diags.AddError(msg, msg)
		return
	}
	diags.AddError(summary, protectedMessage(label))
}

// refuseProtectedReplace fails the plan when this attribute's change forces replacing an object
// whose state has deletion_protection = true: a replacement deletes it.
func refuseProtectedReplace(ctx context.Context, resourceType string, state tfsdk.State, p path.Path, before, after bool, diags *diag.Diagnostics) {
	if !after || before {
		return
	}
	// A resource type is the provider's name, an underscore, then the resource's own name.
	noun := strings.ReplaceAll(resourceType[strings.Index(resourceType, "_")+1:], "_", " ")
	label, ok := protectedLabel(ctx, state, noun)
	switch {
	case !ok:
	case pulumiNames != nil:
		addProtectedError(diags, "Deletion protection is on for "+label, label)
	default:
		diags.AddAttributeError(p, "Deletion protection is on for "+label, protectedMessage(label))
	}
}

// warnReplace adds a plan warning on the attribute whose change made this modifier force
// replacement. Terraform marks the line "# forces replacement", but in a long plan that is easy to
// miss while the object, its id and everything it holds are about to be deleted. The bridge prints
// a warning's detail before its summary, so under Pulumi the whole sentence is the detail. Under
// Pulumi no diagnostic carries the attribute: the bridge would append its Terraform path.
func warnReplace(resourceType string, p path.Path, before, after bool, diags *diag.Diagnostics) {
	if !after || before {
		return
	}
	if pulumiNames != nil {
		order := "Pulumi creates the new one first, then deletes this one."
		if pulumiDeletesFirst[resourceType] {
			order = "Pulumi deletes this one first, then creates the new one."
		}
		diags.AddWarning("", "Changing "+pulumiNames(resourceType, p)+" replaces this resource: "+order)
		return
	}
	diags.AddAttributeWarning(p, "Changing "+p.String()+" replaces the resource",
		"Applying this plan deletes the existing object and creates a new one in its place.")
}

// syncIdentity copies a resource's import parts from its state into its identity, after every
// create, read, update and import, so `terraform query` and identity imports find the object.
func syncIdentity(ctx context.Context, state tfsdk.State, identity *tfsdk.ResourceIdentity, names []string, diags *diag.Diagnostics) {
	if diags.HasError() || identity == nil || state.Raw.IsNull() {
		return
	}
	for _, name := range names {
		var v types.String
		if d := state.GetAttribute(ctx, path.Root(name), &v); d.HasError() {
			diags.Append(d...)
			return
		}
		diags.Append(identity.SetAttribute(ctx, path.Root(name), v)...)
	}
}

// stateFromIdentity imports from an `import { identity = {...} }` block: its parts are the
// attributes the import id would carry.
func stateFromIdentity(ctx context.Context, identity *tfsdk.ResourceIdentity, state *tfsdk.State, names []string, diags *diag.Diagnostics) {
	for _, name := range names {
		var v types.String
		diags.Append(identity.GetAttribute(ctx, path.Root(name), &v)...)
		if diags.HasError() {
			return
		}
		diags.Append(state.SetAttribute(ctx, path.Root(name), v)...)
	}
}

// configuredKey is the private-state key holding which kept-on-removal attributes the last apply
// had configured. Only an attribute in it can have been removed: one never configured holds the
// server's value from the start, and a plan that warned about it on every run would be noise.
const configuredKey = "configured"

type privateGetter interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

type privateSetter interface {
	SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics
}

// importedKey marks an object terraform import brought in. The API never returns a create-only
// value, so an imported state holds none, and a configuration that still sets one (the template it was
// created from) plans as an update, which sends nothing, instead of replacing the object just imported.
// The first update clears the mark: from then on state holds the configured value, and a change replaces.
const importedKey = "imported"

func markImported(ctx context.Context, private privateSetter, diags *diag.Diagnostics) {
	diags.Append(private.SetKey(ctx, importedKey, []byte("true"))...)
}

func clearImported(ctx context.Context, private privateSetter, diags *diag.Diagnostics) {
	if diags.HasError() || private == nil {
		return
	}
	diags.Append(private.SetKey(ctx, importedKey, nil)...)
}

// keepImported reports whether a replacement a create-only attribute forced is only the gap an
// import leaves: no value in state, on an object still marked imported.
func keepImported(ctx context.Context, private privateGetter, stateNull bool) bool {
	if !stateNull {
		return false
	}
	raw, d := private.GetKey(ctx, importedKey)
	return !d.HasError() && string(raw) == "true"
}

// rememberConfigured records which of names the applied configuration set. It runs after a create
// or update that succeeded; a refresh carries the record over unchanged.
func rememberConfigured(ctx context.Context, config tfsdk.Config, private privateSetter, names []string, diags *diag.Diagnostics) {
	if diags.HasError() || private == nil {
		return
	}
	set := []string{}
	for _, name := range names {
		var v attr.Value
		if d := config.GetAttribute(ctx, path.Root(name), &v); d.HasError() || v == nil || v.IsNull() {
			continue
		}
		set = append(set, name)
	}
	raw, err := json.Marshal(set)
	if err != nil {
		return
	}
	diags.Append(private.SetKey(ctx, configuredKey, raw)...)
}

// warnKeptOnRemoval warns, for each of names the last apply configured and this configuration no
// longer sets, that the server keeps its value: the attribute has no default to reset to and no
// null to send, so the plan shows no change for the removed line. Setting it again ends the warning.
func warnKeptOnRemoval(ctx context.Context, resourceType string, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, names []string) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || req.Private == nil {
		return
	}
	raw, d := privateGetter(req.Private).GetKey(ctx, configuredKey)
	if d.HasError() || len(raw) == 0 {
		return
	}
	var applied []string
	if json.Unmarshal(raw, &applied) != nil {
		return
	}
	was := map[string]bool{}
	for _, name := range applied {
		was[name] = true
	}
	for _, name := range names {
		if !was[name] {
			continue
		}
		var config, state attr.Value
		if d := req.Config.GetAttribute(ctx, path.Root(name), &config); d.HasError() || config == nil || !config.IsNull() {
			continue
		}
		if d := req.State.GetAttribute(ctx, path.Root(name), &state); d.HasError() || state == nil || state.IsNull() {
			continue
		}
		p := path.Root(name)
		if pulumiNames != nil {
			resp.Diagnostics.AddWarning("", "Removing "+pulumiNames(resourceType, p)+" keeps its current value in Better Stack: it has no default to reset to. Set it to the value you want instead.")
			continue
		}
		resp.Diagnostics.AddAttributeWarning(p, "Removing "+name+" keeps its current value",
			name+" has no default to reset to, so Better Stack keeps the value it has now and the plan shows no change. Set it to the value you want instead.")
	}
}

// resource is the Terraform type, for the Pulumi property name.
// keyReplaceString forces replacement when a key changes, except in letter case only: the API matches
// the key without case, so the object stays the same one.
type keyReplaceString struct{}

func (keyReplaceString) Description(context.Context) string {
	return "Changing it, other than its letter case, replaces the resource."
}

func (m keyReplaceString) MarkdownDescription(ctx context.Context) string { return m.Description(ctx) }

func (keyReplaceString) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if req.PlanValue.IsUnknown() || !strings.EqualFold(req.StateValue.ValueString(), req.PlanValue.ValueString()) {
		resp.RequiresReplace = true
	}
}

type replaceWarningString struct {
	planmodifier.String
	resource string
}

func (m replaceWarningString) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	before := resp.RequiresReplace
	m.String.PlanModifyString(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningInt64 struct {
	planmodifier.Int64
	resource string
}

func (m replaceWarningInt64) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	before := resp.RequiresReplace
	m.Int64.PlanModifyInt64(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningFloat64 struct {
	planmodifier.Float64
	resource string
}

func (m replaceWarningFloat64) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	before := resp.RequiresReplace
	m.Float64.PlanModifyFloat64(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningBool struct {
	planmodifier.Bool
	resource string
}

func (m replaceWarningBool) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	before := resp.RequiresReplace
	m.Bool.PlanModifyBool(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningList struct {
	planmodifier.List
	resource string
}

func (m replaceWarningList) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	before := resp.RequiresReplace
	m.List.PlanModifyList(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningSet struct {
	planmodifier.Set
	resource string
}

func (m replaceWarningSet) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	before := resp.RequiresReplace
	m.Set.PlanModifySet(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningMap struct {
	planmodifier.Map
	resource string
}

func (m replaceWarningMap) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	before := resp.RequiresReplace
	m.Map.PlanModifyMap(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningObject struct {
	planmodifier.Object
	resource string
}

func (m replaceWarningObject) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	before := resp.RequiresReplace
	m.Object.PlanModifyObject(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningDynamic struct {
	planmodifier.Dynamic
	resource string
}

func (m replaceWarningDynamic) PlanModifyDynamic(ctx context.Context, req planmodifier.DynamicRequest, resp *planmodifier.DynamicResponse) {
	before := resp.RequiresReplace
	m.Dynamic.PlanModifyDynamic(ctx, req, resp)
	if !before && resp.RequiresReplace && keepImported(ctx, req.Private, req.StateValue.IsNull()) {
		resp.RequiresReplace = false
	}
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// patternWhenString refuses a value whose form its sibling rules out: a field that holds ids of
// several kinds (a catalog item_id) takes a team only as t123, so a bare team number fails at plan
// rather than reading back as t123 and never planning clean.
type patternWhenString struct {
	Sibling  string
	Patterns map[string]string
}

func (v patternWhenString) Description(context.Context) string {
	parts := make([]string, 0, len(v.Patterns))
	for value, pattern := range v.Patterns {
		parts = append(parts, "matches "+pattern+" when "+v.Sibling+" is "+value)
	}
	sort.Strings(parts)
	return "The value " + strings.Join(parts, ", ") + "."
}
func (v patternWhenString) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }
func (v patternWhenString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	var sibling types.String
	if diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName(v.Sibling), &sibling); diags.HasError() || sibling.IsNull() || sibling.IsUnknown() {
		return
	}
	pattern, ok := v.Patterns[sibling.ValueString()]
	if !ok {
		return
	}
	if re, err := regexp.Compile(pattern); err == nil && !re.MatchString(req.ConfigValue.ValueString()) {
		want := idFormMessage(pattern)
		// No closing period: Pulumi's bridge appends its own sentence after the message.
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid value",
			fmt.Sprintf("%s is %q, but when %s is %s it %s", req.Path, req.ConfigValue.ValueString(), v.Sibling, sibling.ValueString(), want))
	}
}

// Every plan-time refusal of a malformed team or user ID says how to get one, in these words and
// nowhere else. No closing period: Pulumi's bridge appends its own sentence after the message.
const (
	teamIDPattern     = `^t[1-9][0-9]*$`
	userIDPattern     = `^(u[1-9][0-9]*|[^@/\s]+@[^@/\s]+)$`
	userIDOnlyPattern = `^u[1-9][0-9]*$`
	teamIDMessage     = "must be a team ID such as t123 (get it via List teams)"
	userIDMessage     = "must be a user ID such as u123, or an e-mail (get the user ID via List team members)"
	userIDOnlyMessage = "must be a user ID such as u123 (get it via List team members)"
)

// idFormMessage is what a value refused by pattern must look like instead.
func idFormMessage(pattern string) string {
	switch pattern {
	case teamIDPattern:
		return teamIDMessage
	case userIDPattern:
		return userIDMessage
	case userIDOnlyPattern:
		return userIDOnlyMessage
	}
	return "must match " + pattern
}

// --- attributes that apply only for some sibling values (when:) -------------------------------

// siblingCond is one sibling a when: attribute reads: a top-level attribute, or a field inside an
// object one ("escalation_target.policy_id"), and the values it accepts there. Absent accepts a null
// or missing field.
type siblingCond struct {
	Path   string
	Values []string
	Absent bool
}

// siblingMatches reads the siblings from a plan or config and says whether they are known and whether
// every one holds a value it accepts. A known mismatch decides at once, whatever is still unknown.
func siblingMatches(raw tftypes.Value, conds []siblingCond) (known bool, match bool) {
	return siblingMatchesPlan(raw, tftypes.Value{}, tftypes.Value{}, conds)
}

// siblingMatchesPlan reads the siblings from the plan, except a field inside an object sibling:
//   - set in the configuration, the configured object is what the plan settles on (its reads are
//     exact, so a field it leaves unset is absent), while the plan an attribute's modifier sees may
//     still hold the object's previous fields (the policy of a target switched back to a team);
//   - left out of the configuration on create, the object holds no fields: the API can't fill in a
//     policy nobody named (an alert without a target routes to the current team);
//   - left out of the configuration on update, the object keeps its stored value, so state decides.
//     The plan can't: with any other change the object is still unknown when this attribute plans.
func siblingMatchesPlan(raw, config, state tftypes.Value, conds []siblingCond) (known bool, match bool) {
	create := config.Type() != nil && (state.Type() == nil || state.IsNull())
	known, match = true, true
	for _, c := range conds {
		var k, m bool
		switch {
		case !strings.Contains(c.Path, ".") || config.Type() == nil:
			k, m = condMatches(raw, c)
		case objectConfigured(config, c):
			k, m = condMatches(config, c)
		case create:
			k, m = true, c.Absent
		default:
			k, m = condMatches(state, c)
		}
		if k && !m {
			return true, false
		}
		if !k {
			known = false
		}
	}
	if !known {
		return false, false
	}
	return true, match
}

// objectConfigured says whether the configuration sets the object a dotted condition reads into.
func objectConfigured(config tftypes.Value, c siblingCond) bool {
	top := strings.SplitN(c.Path, ".", 2)[0]
	got, _, err := tftypes.WalkAttributePath(config, tftypes.NewAttributePath().WithAttributeName(top))
	if err != nil {
		return false
	}
	v, ok := got.(tftypes.Value)
	return ok && v.IsKnown() && !v.IsNull()
}

// keptOut says whether an existing resource's when: attribute is null in state while the configuration
// leaves out every object its conditions read into.
func keptOut(state tftypes.Value, stateValue nullable, config tftypes.Value, conds []siblingCond) bool {
	return state.Type() != nil && !state.IsNull() && stateValue.IsNull() && objectsUnconfigured(config, conds)
}

// objectsUnconfigured says whether the configuration leaves out every object the conditions read into.
func objectsUnconfigured(config tftypes.Value, conds []siblingCond) bool {
	any := false
	for _, c := range conds {
		if !strings.Contains(c.Path, ".") {
			return false
		}
		if objectConfigured(config, c) {
			return false
		}
		any = true
	}
	return any
}

func condMatches(raw tftypes.Value, c siblingCond) (known bool, match bool) {
	cur := raw
	for _, step := range strings.Split(c.Path, ".") {
		got, _, err := tftypes.WalkAttributePath(cur, tftypes.NewAttributePath().WithAttributeName(step))
		if err != nil {
			return true, c.Absent
		}
		v, ok := got.(tftypes.Value)
		if !ok || !v.IsKnown() {
			return false, false
		}
		if v.IsNull() {
			return true, c.Absent
		}
		cur = v
	}
	var s string
	switch {
	case cur.Type().Is(tftypes.String):
		_ = cur.As(&s)
	case cur.Type().Is(tftypes.Bool):
		var b bool
		_ = cur.As(&b)
		s = strconv.FormatBool(b)
	case cur.Type().Is(tftypes.Number):
		n := new(big.Float)
		_ = cur.As(&n)
		s = n.Text('f', -1)
	default:
		return true, false
	}
	for _, want := range c.Values {
		if s == want {
			return true, true
		}
	}
	return true, false
}

// condsText says when the attribute applies, as the validators' description and error read it.
func condsText(conds []siblingCond) string {
	parts := make([]string, 0, len(conds))
	for _, c := range conds {
		values := append([]string{}, c.Values...)
		if c.Absent {
			values = append(values, "null")
		}
		text := values[len(values)-1]
		if len(values) > 1 {
			text = strings.Join(values[:len(values)-1], ", ") + " or " + text
		}
		parts = append(parts, c.Path+" is "+text)
	}
	return strings.Join(parts, " and ")
}

func appliesWhenText(p path.Path, conds []siblingCond) string {
	if len(conds) == 1 && !strings.Contains(conds[0].Path, ".") && !conds[0].Absent {
		c := conds[0]
		return fmt.Sprintf("%s applies only when %s is %s; remove it for this %s.", p, c.Path, strings.Join(c.Values, " or "), c.Path)
	}
	return fmt.Sprintf("%s applies only when %s; remove it here.", p, condsText(conds))
}

type nullable interface {
	IsNull() bool
	IsUnknown() bool
}

func appliesWhenRefuse(config nullable, raw tftypes.Value, p path.Path, conds []siblingCond, diags *diag.Diagnostics) {
	if config.IsNull() || config.IsUnknown() {
		return
	}
	if known, match := siblingMatchesConfig(raw, conds); known && !match {
		diags.AddAttributeError(p, "Attribute doesn't apply", appliesWhenText(p, conds))
	}
}

// siblingMatchesConfig reads the siblings from the configuration alone. A field inside an object
// sibling the configuration doesn't set is unknown there: the plan takes that object from state, and
// the plan modifier refuses the attribute once it knows.
func siblingMatchesConfig(config tftypes.Value, conds []siblingCond) (known bool, match bool) {
	known, match = true, true
	for _, c := range conds {
		k, m := condMatches(config, c)
		if strings.Contains(c.Path, ".") && !objectConfigured(config, c) {
			k, m = false, false
		}
		if k && !m {
			return true, false
		}
		if !k {
			known = false
		}
	}
	if !known {
		return false, false
	}
	return true, match
}

// appliesWhenPlanBool plans a when: attribute from its siblings: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanBool struct {
	When    []siblingCond
	Default types.Bool
}

func (m appliesWhenPlanBool) Description(context.Context) string {
	return "Plans this attribute from its siblings: its default where it applies (" + condsText(m.When) + "), null elsewhere."
}
func (m appliesWhenPlanBool) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanBool) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	known, match := siblingMatchesPlan(req.Plan.Raw, req.Config.Raw, req.State.Raw, m.When)
	switch {
	case !known:
	case !match:
		if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
			resp.Diagnostics.AddAttributeError(req.Path, "Attribute doesn't apply", appliesWhenText(req.Path, m.When))
		}
		resp.PlanValue = types.BoolNull()
	case req.ConfigValue.IsNull() && !m.Default.IsNull() && keptOut(req.State.Raw, req.StateValue, req.Config.Raw, m.When):
		// The API left it out of the last read although its stored siblings say it applies (an
		// alert whose policy was deleted reads its target as unknown), so sending the default
		// wouldn't bring it back: it stays out until the configuration names the siblings.
		resp.PlanValue = req.StateValue
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigBool struct {
	When []siblingCond
}

func (v appliesWhenConfigBool) Description(context.Context) string {
	return "Applies only when " + condsText(v.When) + "."
}
func (v appliesWhenConfigBool) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigBool) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.When, &resp.Diagnostics)
}

// appliesWhenPlanString plans a when: attribute from its siblings: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanString struct {
	When    []siblingCond
	Default types.String
}

func (m appliesWhenPlanString) Description(context.Context) string {
	return "Plans this attribute from its siblings: its default where it applies (" + condsText(m.When) + "), null elsewhere."
}
func (m appliesWhenPlanString) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanString) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	known, match := siblingMatchesPlan(req.Plan.Raw, req.Config.Raw, req.State.Raw, m.When)
	switch {
	case !known:
	case !match:
		if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
			resp.Diagnostics.AddAttributeError(req.Path, "Attribute doesn't apply", appliesWhenText(req.Path, m.When))
		}
		resp.PlanValue = types.StringNull()
	case req.ConfigValue.IsNull() && !m.Default.IsNull() && keptOut(req.State.Raw, req.StateValue, req.Config.Raw, m.When):
		// The API left it out of the last read although its stored siblings say it applies (an
		// alert whose policy was deleted reads its target as unknown), so sending the default
		// wouldn't bring it back: it stays out until the configuration names the siblings.
		resp.PlanValue = req.StateValue
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigString struct {
	When []siblingCond
}

func (v appliesWhenConfigString) Description(context.Context) string {
	return "Applies only when " + condsText(v.When) + "."
}
func (v appliesWhenConfigString) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.When, &resp.Diagnostics)
}

// appliesWhenPlanInt64 plans a when: attribute from its siblings: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanInt64 struct {
	When    []siblingCond
	Default types.Int64
}

func (m appliesWhenPlanInt64) Description(context.Context) string {
	return "Plans this attribute from its siblings: its default where it applies (" + condsText(m.When) + "), null elsewhere."
}
func (m appliesWhenPlanInt64) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanInt64) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	known, match := siblingMatchesPlan(req.Plan.Raw, req.Config.Raw, req.State.Raw, m.When)
	switch {
	case !known:
	case !match:
		if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
			resp.Diagnostics.AddAttributeError(req.Path, "Attribute doesn't apply", appliesWhenText(req.Path, m.When))
		}
		resp.PlanValue = types.Int64Null()
	case req.ConfigValue.IsNull() && !m.Default.IsNull() && keptOut(req.State.Raw, req.StateValue, req.Config.Raw, m.When):
		// The API left it out of the last read although its stored siblings say it applies (an
		// alert whose policy was deleted reads its target as unknown), so sending the default
		// wouldn't bring it back: it stays out until the configuration names the siblings.
		resp.PlanValue = req.StateValue
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigInt64 struct {
	When []siblingCond
}

func (v appliesWhenConfigInt64) Description(context.Context) string {
	return "Applies only when " + condsText(v.When) + "."
}
func (v appliesWhenConfigInt64) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigInt64) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.When, &resp.Diagnostics)
}

// appliesWhenPlanFloat64 plans a when: attribute from its siblings: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanFloat64 struct {
	When    []siblingCond
	Default types.Float64
}

func (m appliesWhenPlanFloat64) Description(context.Context) string {
	return "Plans this attribute from its siblings: its default where it applies (" + condsText(m.When) + "), null elsewhere."
}
func (m appliesWhenPlanFloat64) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanFloat64) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	known, match := siblingMatchesPlan(req.Plan.Raw, req.Config.Raw, req.State.Raw, m.When)
	switch {
	case !known:
	case !match:
		if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
			resp.Diagnostics.AddAttributeError(req.Path, "Attribute doesn't apply", appliesWhenText(req.Path, m.When))
		}
		resp.PlanValue = types.Float64Null()
	case req.ConfigValue.IsNull() && !m.Default.IsNull() && keptOut(req.State.Raw, req.StateValue, req.Config.Raw, m.When):
		// The API left it out of the last read although its stored siblings say it applies (an
		// alert whose policy was deleted reads its target as unknown), so sending the default
		// wouldn't bring it back: it stays out until the configuration names the siblings.
		resp.PlanValue = req.StateValue
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigFloat64 struct {
	When []siblingCond
}

func (v appliesWhenConfigFloat64) Description(context.Context) string {
	return "Applies only when " + condsText(v.When) + "."
}
func (v appliesWhenConfigFloat64) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigFloat64) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.When, &resp.Diagnostics)
}

// itemMatch names the ways to identify an item of a configured list whose id is not configured: each
// key is a set of fields, tried in order until one has every field known. The nested lists are matched
// the same way, across every parent, so an item that moves to another parent keeps its id.
type itemMatch struct {
	Keys     [][]string
	Children map[string]itemMatch
}

// itemPool is the prior state's items of one list path, each taken at most once. An item whose id a
// planned item names explicitly is reserved for it, so no other item takes it by key first.
type itemPool struct {
	items    []types.Object
	used     []bool
	reserved []bool
}

func newItemPool(items []types.Object, planned []types.Object) *itemPool {
	pool := &itemPool{items: items, used: make([]bool, len(items)), reserved: make([]bool, len(items))}
	for _, item := range planned {
		id := item.Attributes()["id"]
		if !knownValue(id) {
			continue
		}
		for index, prior := range items {
			if value, ok := prior.Attributes()["id"]; ok && value.Equal(id) {
				pool.reserved[index] = true
			}
		}
	}
	return pool
}

func knownValue(v attr.Value) bool { return v != nil && !v.IsNull() && !v.IsUnknown() }

func listObjects(v attr.Value) []types.Object {
	list, ok := v.(types.List)
	if !ok || list.IsNull() || list.IsUnknown() {
		return nil
	}
	out := make([]types.Object, 0, len(list.Elements()))
	for _, element := range list.Elements() {
		if object, ok := element.(types.Object); ok && !object.IsNull() && !object.IsUnknown() {
			out = append(out, object)
		}
	}
	return out
}

// take finds the prior item a planned one stands for: the one with its id when the id is known,
// otherwise the first free one that equals the planned item on the first key whose fields are all known.
func (p *itemPool) take(attrs map[string]attr.Value, keys [][]string) (types.Object, bool) {
	if id := attrs["id"]; knownValue(id) {
		return p.first([]string{"id"}, attrs, true)
	}
	for _, key := range keys {
		known := len(key) > 0
		for _, field := range key {
			known = known && knownValue(attrs[field])
		}
		if known {
			return p.first(key, attrs, false)
		}
	}
	return types.Object{}, false
}

func (p *itemPool) first(key []string, attrs map[string]attr.Value, explicit bool) (types.Object, bool) {
	for index, item := range p.items {
		if p.used[index] || (p.reserved[index] && !explicit) {
			continue
		}
		prior := item.Attributes()
		matches := true
		for _, field := range key {
			if value, ok := prior[field]; !ok || !value.Equal(attrs[field]) {
				matches = false
				break
			}
		}
		if matches {
			p.used[index] = true
			return item, true
		}
	}
	return types.Object{}, false
}

// matchPlannedList gives each planned item that matches a prior one the prior values of the
// attributes the plan leaves unknown (its id, and the optional values it does not configure, which
// the API keeps when a write omits them), so a reorder or a move updates items in place instead of
// replacing them.
//
// A top-level item that matches no prior one by id or key, typically a renamed one, takes the prior
// item at its own position when nothing else took that one: the prior item would otherwise be
// removed and recreated, losing what hangs off it (a catalog column's values).
func matchPlannedList(ctx context.Context, planned, prior attr.Value, match itemMatch, diags *diag.Diagnostics) attr.Value {
	return matchItemsAt(ctx, planned, newItemPool(listObjects(prior), listObjects(planned)), match, true, diags)
}

// A value it can't build fails the plan: planned without ids, the write would recreate every item.
func matchItems(ctx context.Context, planned attr.Value, pool *itemPool, match itemMatch, diags *diag.Diagnostics) attr.Value {
	return matchItemsAt(ctx, planned, pool, match, false, diags)
}

func matchItemsAt(ctx context.Context, planned attr.Value, pool *itemPool, match itemMatch, byPosition bool, diags *diag.Diagnostics) attr.Value {
	list, ok := planned.(types.List)
	if !ok || list.IsNull() || list.IsUnknown() {
		return planned
	}
	children := map[string]*itemPool{}
	for name := range match.Children {
		var all, named []types.Object
		for _, item := range pool.items {
			all = append(all, listObjects(item.Attributes()[name])...)
		}
		for _, item := range listObjects(list) {
			named = append(named, listObjects(item.Attributes()[name])...)
		}
		children[name] = newItemPool(all, named)
	}
	fill := func(attrs map[string]attr.Value, prior types.Object) {
		for name, value := range attrs {
			if priorValue, exists := prior.Attributes()[name]; exists && value.IsUnknown() {
				attrs[name] = priorValue
			}
		}
	}
	all := make([]map[string]attr.Value, len(list.Elements()))
	unmatched := make([]bool, len(list.Elements()))
	for index, element := range list.Elements() {
		object, ok := element.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			continue
		}
		attrs := map[string]attr.Value{}
		for name, value := range object.Attributes() {
			attrs[name] = value
		}
		if prior, ok := pool.take(attrs, match.Keys); ok {
			fill(attrs, prior)
		} else {
			unmatched[index] = !knownValue(attrs["id"])
		}
		all[index] = attrs
	}
	if byPosition {
		for index, attrs := range all {
			if attrs == nil || !unmatched[index] || index >= len(pool.items) || pool.used[index] || pool.reserved[index] {
				continue
			}
			pool.used[index] = true
			fill(attrs, pool.items[index])
		}
	}
	elements := make([]attr.Value, 0, len(list.Elements()))
	for index, element := range list.Elements() {
		object, ok := element.(types.Object)
		attrs := all[index]
		if !ok || attrs == nil {
			elements = append(elements, element)
			continue
		}
		for name, childMatch := range match.Children {
			attrs[name] = matchItems(ctx, attrs[name], children[name], childMatch, diags)
		}
		matched, objectDiags := types.ObjectValue(object.AttributeTypes(ctx), attrs)
		diags.Append(objectDiags...)
		if objectDiags.HasError() {
			return planned
		}
		elements = append(elements, matched)
	}
	out, listDiags := types.ListValue(list.ElementType(ctx), elements)
	diags.Append(listDiags...)
	if listDiags.HasError() {
		return planned
	}
	return out
}

// withPlannedIDs gives each item of a configured list the id ModifyPlan matched it to, and only the
// id: the request still leaves out every value the configuration leaves out, which the API keeps.
// An item it can't pair with its planned one fails the write rather than sending it without its id.
func withPlannedIDs(ctx context.Context, request, planned types.List, match itemMatch, diags *diag.Diagnostics) types.List {
	if request.IsNull() || request.IsUnknown() {
		return request
	}
	requested, plannedItems := listObjects(request), listObjects(planned)
	if len(requested) != len(request.Elements()) || len(requested) != len(plannedItems) {
		diags.AddError("matching list items", fmt.Sprintf("the request has %d items and the plan %d, so their ids can't be paired. This is a provider bug.", len(request.Elements()), len(plannedItems)))
		return request
	}
	elements := make([]attr.Value, 0, len(requested))
	for index, object := range requested {
		attrs := map[string]attr.Value{}
		for name, value := range object.Attributes() {
			attrs[name] = value
		}
		plannedAttrs := plannedItems[index].Attributes()
		if id, ok := attrs["id"]; ok && id.IsNull() && knownValue(plannedAttrs["id"]) {
			attrs["id"] = plannedAttrs["id"]
		}
		for name, childMatch := range match.Children {
			child, childOK := attrs[name].(types.List)
			plannedChild, plannedOK := plannedAttrs[name].(types.List)
			if childOK && plannedOK {
				attrs[name] = withPlannedIDs(ctx, child, plannedChild, childMatch, diags)
			}
		}
		matched, objectDiags := types.ObjectValue(object.AttributeTypes(ctx), attrs)
		diags.Append(objectDiags...)
		if objectDiags.HasError() {
			return request
		}
		elements = append(elements, matched)
	}
	out, listDiags := types.ListValue(request.ElementType(ctx), elements)
	diags.Append(listDiags...)
	if listDiags.HasError() {
		return request
	}
	return out
}

// keepListItemIDs plans the stored list, ids included, when the planned items equal the stored ones
// apart from their read-only "id": for a list the API recreates whole when it's sent, the write then
// leaves it out, so the items keep their ids and a plan made for another change doesn't name it.
func keepListItemIDs(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse, p path.Path) {
	if resp.Diagnostics.HasError() || req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	var planned, stored types.List
	if resp.Plan.GetAttribute(ctx, p, &planned).HasError() || req.State.GetAttribute(ctx, p, &stored).HasError() {
		return
	}
	if planned.IsNull() || planned.IsUnknown() || stored.IsNull() || stored.IsUnknown() {
		return
	}
	plannedItems, storedItems := planned.Elements(), stored.Elements()
	if len(plannedItems) != len(storedItems) {
		return
	}
	for i := range plannedItems {
		plannedItem, ok := plannedItems[i].(types.Object)
		storedItem, ok2 := storedItems[i].(types.Object)
		if !ok || !ok2 || plannedItem.IsNull() || plannedItem.IsUnknown() || storedItem.IsNull() {
			return
		}
		storedAttrs := storedItem.Attributes()
		for name, value := range plannedItem.Attributes() {
			if name == "id" && (value.IsNull() || value.IsUnknown()) {
				continue
			}
			if !value.Equal(storedAttrs[name]) {
				return
			}
		}
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, p, stored)...)
}
