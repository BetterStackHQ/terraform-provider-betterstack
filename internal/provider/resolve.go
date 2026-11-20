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
				sub, present = apiMap[name]
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
			out[name] = valueFromAPI(at, apiMap[name])
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

func protectedLabel(ctx context.Context, state tfsdk.State, noun string) (label string, ok bool) {
	if state.Schema == nil || state.Raw.IsNull() {
		return "", false
	}
	var protected types.Bool
	if diags := state.GetAttribute(ctx, path.Root("deletion_protection"), &protected); diags.HasError() || !protected.ValueBool() {
		return "", false
	}
	var name types.String
	if diags := state.GetAttribute(ctx, path.Root("name"), &name); !diags.HasError() && name.ValueString() != "" {
		if noun == "" {
			return fmt.Sprintf("%q", name.ValueString()), true
		}
		return fmt.Sprintf("the %s %q", noun, name.ValueString()), true
	}
	if noun == "" {
		return "this object", true
	}
	return "the " + noun, true
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
		if d := attr.FieldByName("Description"); d.IsValid() && d.Kind() == reflect.String {
			text := d.String()
			if pulumiDescriptionEdit != nil {
				text = pulumiDescriptionEdit(text)
			}
			d.SetString(QuoteAPIValues(text, names))
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
func QuoteAPIValues(md string, names map[string]bool) string {
	var b strings.Builder
	last := 0
	for _, loc := range codeFence.FindAllStringIndex(md, -1) {
		b.WriteString(quoteProse(md[last:loc[0]], names))
		b.WriteString(md[loc[0]:loc[1]])
		last = loc[1]
	}
	b.WriteString(quoteProse(md[last:], names))
	return b.String()
}

// valueList is a run of text that lists only values, whatever their spelling: a monitor's
// "Valid values: `status`, `keyword`" names its monitor_type values, not its status attribute.
var valueList = regexp.MustCompile("Valid values: [^\n]*?\\.(?:\\s|$)|Defaults to `[^`]*`")

func quoteProse(text string, names map[string]bool) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = valueList.ReplaceAllStringFunc(line, func(list string) string { return quoteLine(list, nil, strings.HasPrefix(lines[i], "- `")) })
		lines[i] = quoteLine(line, names, strings.HasPrefix(line, "- `"))
	}
	return strings.Join(lines, "\n")
}

func quoteLine(line string, names map[string]bool, attrLine bool) string {
	for _, re := range []*regexp.Regexp{backtickedWord, doubleQuotedSnake} {
		line = re.ReplaceAllStringFunc(line, func(m string) string {
			word := re.FindStringSubmatch(m)[1]
			switch {
			case names[word]:
				return m
			case word == "true" || word == "false" || word == "null":
				return word
			case attrLine:
				return "'" + strings.ReplaceAll(word, "_", `\_`) + "'"
			default:
				return "'" + word + "'"
			}
		})
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
		if pulumiNames != nil {
			resp.Diagnostics.AddError("Deletion protection is on for "+label, "this update deletes it and everything stored in it. "+pulumiProtectionFix)
			return
		}
		resp.Diagnostics.AddError("Deletion protection is on for "+label,
			"This plan destroys "+label+", which permanently deletes it and everything stored in it. "+
				"Set deletion_protection = false and apply that first. To stop managing it without deleting it, "+
				"use a removed block with lifecycle { destroy = false }.")
	}
}

// pulumiProtectionFix is what a Pulumi user does instead: the same turn-off-first, or Pulumi's own
// way to let go of the object while keeping it.
const pulumiProtectionFix = "Set deletionProtection: false and run pulumi up first. To stop managing it without deleting it, " +
	"set the retainOnDelete resource option before removing it from the program, or run pulumi state delete."

// refuseProtectedDelete stops a delete of an object whose state still has deletion_protection = true.
// The plan checks above never see a pulumi destroy, a terraform apply -replace or a tainted object,
// so the delete itself checks too, before it sends force=true.
func refuseProtectedDelete(ctx context.Context, state tfsdk.State, noun string, diags *diag.Diagnostics) bool {
	label, ok := protectedLabel(ctx, state, noun)
	if ok && pulumiNames != nil {
		diags.AddError("Deletion protection is on for "+label+", so it was not deleted", "deleting it would delete everything stored in it. "+pulumiProtectionFix)
	} else if ok {
		diags.AddError("Deletion protection is on for "+label,
			"Deleting "+label+" permanently deletes it and everything stored in it, so it was not deleted. "+
				"Set deletion_protection = false and apply that first. To stop managing it without deleting it, "+
				"use a removed block with lifecycle { destroy = false }, or pulumi state delete.")
	}
	return ok
}

// refuseProtectedReplace fails the plan when this attribute's change forces replacing an object
// whose state has deletion_protection = true: a replacement deletes it.
func refuseProtectedReplace(ctx context.Context, resourceType string, state tfsdk.State, p path.Path, before, after bool, diags *diag.Diagnostics) {
	if !after || before {
		return
	}
	label, ok := protectedLabel(ctx, state, "")
	switch {
	case !ok:
	case pulumiNames != nil:
		// Pulumi creates the replacement first, so this object is deleted after the new one exists.
		diags.AddError("Deletion protection is on for "+label, "changing "+pulumiNames(resourceType, p)+" replaces it, "+
			"and the replacement deletes it and everything stored in it once the new one is created. "+
			"Undo the change, or set deletionProtection: false and run pulumi up first.")
	default:
		diags.AddAttributeError(p, "Deletion protection is on for "+label,
			"Changing "+p.String()+" replaces it, and a replacement deletes it and everything stored in it before creating a new one. "+
				"Undo the change, or set deletion_protection = false and apply that first.")
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
		diags.AddWarning("", "Changing "+pulumiNames(resourceType, p)+" replaces this resource: Pulumi creates the new one first, then deletes this one.")
		return
	}
	diags.AddAttributeWarning(p, "Changing "+p.String()+" replaces the resource",
		"Applying this plan deletes the existing object and creates a new one in its place.")
}

// resource is the Terraform type, for the Pulumi property name.
type replaceWarningString struct {
	planmodifier.String
	resource string
}

func (m replaceWarningString) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	before := resp.RequiresReplace
	m.String.PlanModifyString(ctx, req, resp)
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
	warnReplace(m.resource, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
	refuseProtectedReplace(ctx, m.resource, req.State, req.Path, before, resp.RequiresReplace, &resp.Diagnostics)
}

// --- attributes that apply only for some sibling values (when:) -------------------------------

// siblingMatches reads a top-level sibling attribute from a plan or config and says whether its
// value is known and whether it is one of values.
func siblingMatches(raw tftypes.Value, sibling string, values []string) (known bool, match bool) {
	got, _, err := tftypes.WalkAttributePath(raw, tftypes.NewAttributePath().WithAttributeName(sibling))
	if err != nil {
		return true, false
	}
	v, ok := got.(tftypes.Value)
	if !ok || !v.IsKnown() {
		return false, false
	}
	if v.IsNull() {
		return true, false
	}
	var s string
	switch {
	case v.Type().Is(tftypes.String):
		_ = v.As(&s)
	case v.Type().Is(tftypes.Bool):
		var b bool
		_ = v.As(&b)
		s = strconv.FormatBool(b)
	case v.Type().Is(tftypes.Number):
		n := new(big.Float)
		_ = v.As(&n)
		s = n.Text('f', -1)
	default:
		return true, false
	}
	for _, want := range values {
		if s == want {
			return true, true
		}
	}
	return true, false
}

func appliesWhenText(p path.Path, sibling string, values []string) string {
	return fmt.Sprintf("%s applies only when %s is %s; remove it for this %s.", p, sibling, strings.Join(values, " or "), sibling)
}

type nullable interface {
	IsNull() bool
	IsUnknown() bool
}

func appliesWhenRefuse(config nullable, raw tftypes.Value, p path.Path, sibling string, values []string, diags *diag.Diagnostics) {
	if config.IsNull() || config.IsUnknown() {
		return
	}
	if known, match := siblingMatches(raw, sibling, values); known && !match {
		diags.AddAttributeError(p, "Attribute doesn't apply", appliesWhenText(p, sibling, values))
	}
}

// appliesWhenPlanBool plans a when: attribute from its sibling: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanBool struct {
	Sibling string
	Values  []string
	Default types.Bool
}

func (m appliesWhenPlanBool) Description(context.Context) string {
	return "Plans this attribute from " + m.Sibling + ": its default where it applies, null elsewhere."
}
func (m appliesWhenPlanBool) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanBool) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	known, match := siblingMatches(req.Plan.Raw, m.Sibling, m.Values)
	switch {
	case !known:
	case !match:
		resp.PlanValue = types.BoolNull()
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigBool struct {
	Sibling string
	Values  []string
}

func (v appliesWhenConfigBool) Description(context.Context) string {
	return "Applies only when " + v.Sibling + " is " + strings.Join(v.Values, " or ") + "."
}
func (v appliesWhenConfigBool) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigBool) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.Sibling, v.Values, &resp.Diagnostics)
}

// appliesWhenPlanString plans a when: attribute from its sibling: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanString struct {
	Sibling string
	Values  []string
	Default types.String
}

func (m appliesWhenPlanString) Description(context.Context) string {
	return "Plans this attribute from " + m.Sibling + ": its default where it applies, null elsewhere."
}
func (m appliesWhenPlanString) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanString) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	known, match := siblingMatches(req.Plan.Raw, m.Sibling, m.Values)
	switch {
	case !known:
	case !match:
		resp.PlanValue = types.StringNull()
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigString struct {
	Sibling string
	Values  []string
}

func (v appliesWhenConfigString) Description(context.Context) string {
	return "Applies only when " + v.Sibling + " is " + strings.Join(v.Values, " or ") + "."
}
func (v appliesWhenConfigString) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigString) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.Sibling, v.Values, &resp.Diagnostics)
}

// appliesWhenPlanInt64 plans a when: attribute from its sibling: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanInt64 struct {
	Sibling string
	Values  []string
	Default types.Int64
}

func (m appliesWhenPlanInt64) Description(context.Context) string {
	return "Plans this attribute from " + m.Sibling + ": its default where it applies, null elsewhere."
}
func (m appliesWhenPlanInt64) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanInt64) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	known, match := siblingMatches(req.Plan.Raw, m.Sibling, m.Values)
	switch {
	case !known:
	case !match:
		resp.PlanValue = types.Int64Null()
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigInt64 struct {
	Sibling string
	Values  []string
}

func (v appliesWhenConfigInt64) Description(context.Context) string {
	return "Applies only when " + v.Sibling + " is " + strings.Join(v.Values, " or ") + "."
}
func (v appliesWhenConfigInt64) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigInt64) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.Sibling, v.Values, &resp.Diagnostics)
}

// appliesWhenPlanFloat64 plans a when: attribute from its sibling: null where it doesn't apply, so no
// request carries it there, and its default where it does and the line is unset, so removing the
// line resets it.
type appliesWhenPlanFloat64 struct {
	Sibling string
	Values  []string
	Default types.Float64
}

func (m appliesWhenPlanFloat64) Description(context.Context) string {
	return "Plans this attribute from " + m.Sibling + ": its default where it applies, null elsewhere."
}
func (m appliesWhenPlanFloat64) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (m appliesWhenPlanFloat64) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	known, match := siblingMatches(req.Plan.Raw, m.Sibling, m.Values)
	switch {
	case !known:
	case !match:
		resp.PlanValue = types.Float64Null()
	case req.ConfigValue.IsNull() && !m.Default.IsNull():
		resp.PlanValue = m.Default
	}
}

type appliesWhenConfigFloat64 struct {
	Sibling string
	Values  []string
}

func (v appliesWhenConfigFloat64) Description(context.Context) string {
	return "Applies only when " + v.Sibling + " is " + strings.Join(v.Values, " or ") + "."
}
func (v appliesWhenConfigFloat64) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v appliesWhenConfigFloat64) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	appliesWhenRefuse(req.ConfigValue, req.Config.Raw, req.Path, v.Sibling, v.Values, &resp.Diagnostics)
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
func matchPlannedList(ctx context.Context, planned, prior attr.Value, match itemMatch, diags *diag.Diagnostics) attr.Value {
	return matchItems(ctx, planned, newItemPool(listObjects(prior), listObjects(planned)), match, diags)
}

// A value it can't build fails the plan: planned without ids, the write would recreate every item.
func matchItems(ctx context.Context, planned attr.Value, pool *itemPool, match itemMatch, diags *diag.Diagnostics) attr.Value {
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
	elements := make([]attr.Value, 0, len(list.Elements()))
	for _, element := range list.Elements() {
		object, ok := element.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			elements = append(elements, element)
			continue
		}
		attrs := map[string]attr.Value{}
		for name, value := range object.Attributes() {
			attrs[name] = value
		}
		if prior, ok := pool.take(attrs, match.Keys); ok {
			for name, value := range attrs {
				if priorValue, exists := prior.Attributes()[name]; exists && value.IsUnknown() {
					attrs[name] = priorValue
				}
			}
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
