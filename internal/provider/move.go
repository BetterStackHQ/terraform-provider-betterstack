package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type migrationTransform struct {
	Op         string      `json:"op"`
	Path       string      `json:"path"`
	To         string      `json:"to"`
	Key        string      `json:"key"`
	OnConflict string      `json:"on_conflict"`
	Value      interface{} `json:"value"`
	Factor     json.Number `json:"factor"`
	When       *struct {
		Path string        `json:"path"`
		In   []interface{} `json:"in"`
	} `json:"when"`
}

type migrationSourceAttribute struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	Optional  bool   `json:"optional"`
	Computed  bool   `json:"computed"`
	Sensitive bool   `json:"sensitive"`
	MinItems  int    `json:"min_items"`
	MaxItems  int    `json:"max_items"`
}

type migrationResponseProjection struct {
	Source         string   `json:"source"`
	Target         string   `json:"target"`
	ReadonlyFields []string `json:"readonly_fields"`
}

type legacyMigration struct {
	SourceProvider      string                        `json:"source_provider"`
	SourceType          string                        `json:"source_type"`
	Kind                string                        `json:"kind"`
	Handling            string                        `json:"handling"`
	Identity            string                        `json:"identity"`
	SchemaVersions      []int64                       `json:"schema_versions"`
	Renames             map[string]string             `json:"renames"`
	Constants           map[string]interface{}        `json:"constants"`
	Defaults            map[string]interface{}        `json:"defaults"`
	Transforms          []migrationTransform          `json:"transforms"`
	IgnoredFields       []string                      `json:"ignored_fields"`
	RefreshedFields     []string                      `json:"refreshed_fields"`
	SourceAttributes    []migrationSourceAttribute    `json:"source_attributes"`
	ResponseProjections []migrationResponseProjection `json:"response_projections"`
}

// Only reviewed metadata may change a value. API coercions, shape guesses, network requests and
// best-effort dropping have no place in state conversion. TargetState is committed only on success.
func legacyStateMover(metadata string) resource.StateMover {
	var migration legacyMigration
	decodeErr := decodeNumbers([]byte(metadata), &migration)
	return resource.StateMover{StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
		if decodeErr != nil {
			resp.Diagnostics.AddError("Invalid migration metadata", "Regenerate the provider from reviewed migration metadata.")
			return
		}
		if req.SourceTypeName != migration.SourceType || !legacyProviderMatches(req.SourceProviderAddress, migration.SourceProvider) || migration.Kind != "resource" || migration.Handling != "move" || migration.Identity != "preserve" {
			return
		}
		fail := func(err error) {
			resp.Diagnostics.AddError("Cannot safely migrate "+migration.SourceType, err.Error()+" No state was changed. Keep the legacy configuration and state backup; follow docs/guides/migration before retrying.")
		}
		supported := false
		for _, version := range migration.SchemaVersions {
			if version == req.SourceSchemaVersion {
				supported = true
				break
			}
		}
		if !supported {
			fail(fmt.Errorf("Source schema version %d is not supported by this provider", req.SourceSchemaVersion))
			return
		}
		if req.SourceRawState == nil || len(req.SourceRawState.JSON) == 0 {
			fail(fmt.Errorf("Source JSON state is missing"))
			return
		}
		if len(migration.SourceAttributes) == 0 {
			fail(fmt.Errorf("The provider has no reviewed source schema for this migration"))
			return
		}
		decoded, err := decodeMigrationJSON(req.SourceRawState.JSON)
		if err != nil {
			fail(fmt.Errorf("Source state is not a single valid JSON value with unique object keys"))
			return
		}
		state, ok := decoded.(map[string]interface{})
		if !ok {
			fail(fmt.Errorf("Source state must be an object"))
			return
		}
		if err := validateMigrationSource(state, "object", "", migration.SourceAttributes); err != nil {
			fail(err)
			return
		}
		if id, exists := state["id"]; !exists || id == nil || id == "" {
			fail(fmt.Errorf("Source object identity is missing"))
			return
		}
		for _, path := range migration.IgnoredFields {
			if err := editMigrationPath(state, strings.Split(path, "."), false, func(parent map[string]interface{}, key string) error { delete(parent, key); return nil }); err != nil {
				fail(fmt.Errorf("Cannot remove declared obsolete field %s", path))
				return
			}
		}
		for _, from := range migrationKeys(migration.Renames) {
			if err := renameMigrationPath(state, from, migration.Renames[from], ""); err != nil {
				fail(err)
				return
			}
		}
		for _, transform := range migration.Transforms {
			if err := applyMigrationTransform(state, transform); err != nil {
				fail(err)
				return
			}
		}
		for _, path := range migrationKeys(migration.Constants) {
			value := migration.Constants[path]
			if err := editMigrationPath(state, strings.Split(path, "."), true, func(parent map[string]interface{}, key string) error {
				if prior := parent[key]; prior != nil && !reflect.DeepEqual(prior, value) {
					return fmt.Errorf("Constant conflicts with source state")
				}
				parent[key] = value
				return nil
			}); err != nil {
				fail(fmt.Errorf("Cannot set declared constant %s without losing state", path))
				return
			}
		}
		targetType := resp.TargetState.Schema.Type().TerraformType(ctx)
		targetAttributes := resp.TargetState.Schema.GetAttributes()
		projected := make(map[string]tftypes.Value, len(migration.ResponseProjections))
		for _, projection := range migration.ResponseProjections {
			target, exists := targetAttributes[projection.Target]
			if !exists || !target.IsComputed() || target.IsOptional() || targetAttributes[projection.Source] == nil {
				fail(fmt.Errorf("Response projection does not name a configured input and a computed result"))
				return
			}
			source, exists := state[projection.Source]
			if !exists {
				continue
			}
			if _, exists := state[projection.Target]; exists {
				fail(fmt.Errorf("Response projection conflicts with existing target state"))
				return
			}
			// Snapshot the complete typed result before removing computed keys from the input.
			// Conversion builds independent Terraform values, so no JSON cloning is necessary.
			result, err := strictMigrationValue(targetType.(tftypes.Object).AttributeTypes[projection.Target], source, projection.Target)
			if err != nil {
				fail(err)
				return
			}
			projected[projection.Target] = result
			for _, path := range projection.ReadonlyFields {
				if !strings.HasPrefix(path, projection.Source+".") && !strings.HasPrefix(path, projection.Source+"[].") {
					fail(fmt.Errorf("Readonly projection field is outside its configured input"))
					return
				}
				if err := editMigrationPath(state, strings.Split(path, "."), false, func(parent map[string]interface{}, key string) error {
					delete(parent, key)
					return nil
				}); err != nil {
					fail(fmt.Errorf("Cannot project readonly field %s", path))
					return
				}
			}
		}
		// A defaulted attribute the legacy state doesn't hold starts at its default, as a configuration
		// that omits it plans; the first refresh overwrites what the API returns.
		for _, name := range migrationKeys(migration.Defaults) {
			if prior, exists := state[name]; !exists || prior == nil {
				state[name] = migration.Defaults[name]
			}
		}
		// A reviewed refreshed field stays empty: the legacy state can't hold it, and the first refresh
		// reads it from the API before anything is planned.
		refreshed := map[string]bool{}
		for _, name := range migration.RefreshedFields {
			refreshed[name] = true
		}
		for _, name := range migrationKeys(resp.TargetState.Schema.GetAttributes()) {
			if refreshed[name] && state[name] == nil {
				continue
			}
			if resp.TargetState.Schema.GetAttributes()[name].IsRequired() && (state[name] == nil || state[name] == "") {
				fail(fmt.Errorf("Required target field %s cannot be recovered from source state", name))
				return
			}
		}
		if _, hasID := resp.TargetState.Schema.GetAttributes()["id"]; hasID && (state["id"] == nil || state["id"] == "") {
			fail(fmt.Errorf("Target object identity cannot be recovered from source state"))
			return
		}
		value, err := strictMigrationValue(targetType, state, "state")
		if err != nil {
			fail(err)
			return
		}
		if len(projected) > 0 {
			var attributes map[string]tftypes.Value
			if err := value.As(&attributes); err != nil {
				fail(err)
				return
			}
			for name, result := range projected {
				attributes[name] = result
			}
			value = tftypes.NewValue(targetType, attributes)
		}
		if _, err := resp.TargetState.Schema.Type().ValueFromTerraform(ctx, value); err != nil {
			fail(fmt.Errorf("Converted state cannot be represented by the target schema (check numeric bounds and nested types)"))
			return
		}
		resp.TargetState.Raw = value
	}}
}

// Terraform supplies HOST/NAMESPACE/TYPE. Hosts may differ, but an absent host or a namespace/type
// substring is never a provider identity. Provider names in the registry are case-insensitive.
func legacyProviderMatches(address, expected string) bool {
	for _, char := range address {
		if char > 127 {
			return false
		}
	}
	parts, want := strings.Split(strings.ToLower(address), "/"), strings.Split(strings.ToLower(expected), "/")
	if len(parts) != 3 || len(want) != 3 || parts[0] == "" || want[0] == "" {
		return false
	}
	for _, part := range parts {
		if part == "" || strings.ContainsAny(part, " \t\r\n?#@\\") {
			return false
		}
	}
	return parts[1] == want[1] && parts[2] == want[2]
}

func migrationKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// json.Decoder otherwise accepts duplicate keys (last value wins). Do not let ambiguous state
// silently replace a secret or identity; errors deliberately exclude source tokens and values.
func decodeMigrationJSON(raw []byte) (interface{}, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, err := migrationJSONValue(decoder, 0)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, fmt.Errorf("Trailing JSON data")
	}
	return value, nil
}

func migrationJSONValue(decoder *json.Decoder, depth int) (interface{}, error) {
	if depth > 512 {
		return nil, fmt.Errorf("JSON nesting is too deep")
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("Invalid JSON")
	}
	delim, container := token.(json.Delim)
	if !container {
		return token, nil
	}
	switch delim {
	case '{':
		object := map[string]interface{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, fmt.Errorf("Invalid JSON object")
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, fmt.Errorf("Invalid JSON object key")
			}
			if _, exists := object[key]; exists {
				return nil, fmt.Errorf("Duplicate JSON object key")
			}
			value, err := migrationJSONValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return nil, fmt.Errorf("Invalid JSON object")
		}
		return object, nil
	case '[':
		array := []interface{}{}
		for decoder.More() {
			value, err := migrationJSONValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return nil, fmt.Errorf("Invalid JSON array")
		}
		return array, nil
	default:
		return nil, fmt.Errorf("Unexpected JSON delimiter")
	}
}

// The catalog is extracted from released SDK schemas, including supported historical aliases.
// Validate BEFORE renaming/decoding so newer fields and wrong source types cannot masquerade as
// already-migrated target fields. Optional absent values stay absent; defaults are not invented.
func validateMigrationSource(value interface{}, kind, path string, attributes []migrationSourceAttribute) error {
	if value == nil {
		return nil
	}
	bad := func() error { return fmt.Errorf("Source field %s has an incompatible %s shape", path, kind) }
	switch kind {
	case "string":
		if _, ok := value.(string); !ok {
			return bad()
		}
		return nil
	case "bool":
		if _, ok := value.(bool); !ok {
			return bad()
		}
		return nil
	case "int", "float", "number":
		number, ok := value.(json.Number)
		if !ok {
			return bad()
		}
		rational, ok := new(big.Rat).SetString(number.String())
		if !ok || (kind == "int" && (!rational.IsInt() || !rational.Num().IsInt64())) {
			return bad()
		}
		return nil
	case "object":
		object, ok := value.(map[string]interface{})
		if !ok {
			return bad()
		}
		prefix := path
		if prefix != "" {
			prefix += "."
		}
		fields := map[string]migrationSourceAttribute{}
		for _, attribute := range attributes {
			if strings.HasPrefix(attribute.Name, prefix) {
				name := strings.TrimPrefix(attribute.Name, prefix)
				if !strings.Contains(name, ".") {
					fields[name] = attribute
				}
			}
		}
		for _, key := range migrationKeys(object) {
			if _, exists := fields[key]; !exists {
				return fmt.Errorf("Source object %s contains an unrecognized attribute; refresh with a supported legacy provider or use the reviewed guided recipe", path)
			}
		}
		for _, name := range migrationKeys(fields) {
			field := fields[name]
			child := object[name]
			// Computed-only secrets (collector enrollment and connection passwords) cannot
			// be recovered on read. Optional passwords may legitimately be absent.
			required := field.Required || (field.Computed && field.Sensitive && !field.Optional)
			if required && (child == nil || child == "") {
				return fmt.Errorf("Required source field %s is missing", field.Name)
			}
			if list, ok := child.([]interface{}); ok && ((field.MaxItems > 0 && len(list) > field.MaxItems) || (field.MinItems > 0 && len(list) < field.MinItems)) {
				return fmt.Errorf("Source field %s has unsupported list cardinality", field.Name)
			}
			if err := validateMigrationSource(child, field.Type, field.Name, attributes); err != nil {
				return err
			}
		}
		return nil
	}
	for _, container := range []string{"list", "set", "map"} {
		if !strings.HasPrefix(kind, container+"(") || !strings.HasSuffix(kind, ")") {
			continue
		}
		element := kind[len(container)+1 : len(kind)-1]
		if container == "map" {
			object, ok := value.(map[string]interface{})
			if !ok {
				return bad()
			}
			for _, key := range migrationKeys(object) {
				if err := validateMigrationSource(object[key], element, path, attributes); err != nil {
					return err
				}
			}
		} else {
			array, ok := value.([]interface{})
			if !ok {
				return bad()
			}
			for _, child := range array {
				if err := validateMigrationSource(child, element, path, attributes); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return fmt.Errorf("Source field %s has no reviewed schema type", path)
}

// [] is explicit traversal, never a guessed list/object conversion. Missing optional paths are
// no-ops; a present incompatible container fails. Only constants/rename destinations create maps.
func editMigrationPath(root interface{}, path []string, create bool, edit func(map[string]interface{}, string) error) error {
	if len(path) == 0 || path[0] == "" {
		return fmt.Errorf("Invalid migration path")
	}
	object, ok := root.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Expected object in migration path")
	}
	key := strings.TrimSuffix(path[0], "[]")
	if strings.HasSuffix(path[0], "[]") {
		value, exists := object[key]
		if !exists || value == nil {
			return nil
		}
		array, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("Expected list in migration path")
		}
		for i, child := range array {
			if len(path) == 1 {
				holder := map[string]interface{}{"value": child}
				if err := edit(holder, "value"); err != nil {
					return err
				}
				array[i] = holder["value"]
			} else if err := editMigrationPath(child, path[1:], create, edit); err != nil {
				return err
			}
		}
		return nil
	}
	if len(path) == 1 {
		if _, exists := object[key]; exists || create {
			return edit(object, key)
		}
		return nil
	}
	next := object[key]
	if next == nil {
		if !create {
			return nil
		}
		next = map[string]interface{}{}
		object[key] = next
	}
	return editMigrationPath(next, path[1:], create, edit)
}

func renameMigrationPath(root map[string]interface{}, from, to, onConflict string) error {
	if from == to {
		return nil
	}
	if onConflict != "" && onConflict != "target" {
		return fmt.Errorf("Unsupported rename conflict policy")
	}
	// Array renames operate within the same element; never zip, flatten or reorder collections.
	if index := strings.LastIndex(from, "[]."); index >= 0 {
		prefix := from[:index+3]
		if !strings.HasPrefix(to, prefix) {
			return fmt.Errorf("Rename %s must preserve its array parent", from)
		}
		return editMigrationPath(root, strings.Split(strings.TrimSuffix(prefix, "."), "."), false, func(parent map[string]interface{}, key string) error {
			object, ok := parent[key].(map[string]interface{})
			if !ok {
				return fmt.Errorf("Rename %s requires an object element", from)
			}
			return renameMigrationPath(object, from[len(prefix):], to[len(prefix):], onConflict)
		})
	}
	if strings.Contains(from, "[]") || strings.Contains(to, "[]") {
		return fmt.Errorf("Rename cannot change array cardinality")
	}
	return editMigrationPath(root, strings.Split(from, "."), false, func(parent map[string]interface{}, key string) error {
		value := parent[key]
		if err := editMigrationPath(root, strings.Split(to, "."), true, func(target map[string]interface{}, targetKey string) error {
			prior := target[targetKey]
			if prior != nil && value != nil && !reflect.DeepEqual(prior, value) && onConflict != "target" {
				return fmt.Errorf("Rename %s to %s conflicts with an existing value", from, to)
			}
			if prior == nil {
				target[targetKey] = value
			}
			return nil
		}); err != nil {
			return err
		}
		delete(parent, key)
		return nil
	})
}

func applyMigrationTransform(state map[string]interface{}, transform migrationTransform) error {
	if transform.When != nil {
		matches := false
		err := editMigrationPath(state, strings.Split(transform.When.Path, "."), false, func(parent map[string]interface{}, key string) error {
			for _, candidate := range transform.When.In {
				if reflect.DeepEqual(parent[key], candidate) {
					matches = true
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Cannot evaluate migration condition for %s", transform.Path)
		}
		if !matches {
			return nil
		}
	}
	if transform.Op == "rename" {
		return renameMigrationPath(state, transform.Path, transform.To, transform.OnConflict)
	}
	if transform.Op == "sentinel_omit" && strings.HasSuffix(transform.Path, "[]") {
		return fmt.Errorf("Sentinel omission requires a field path, not an array element")
	}
	switch transform.Op {
	case "json_decode", "singleton", "seconds", "string_id", "sentinel_null", "sentinel_omit", "index_by", "wrap", "composite_id":
	default:
		return fmt.Errorf("Unsupported migration transform %s", transform.Op)
	}
	err := editMigrationPath(state, strings.Split(transform.Path, "."), false, func(parent map[string]interface{}, key string) error {
		value := parent[key]
		if value == nil && transform.Op != "sentinel_omit" {
			return nil
		}
		bad := func() error {
			return fmt.Errorf("Field %s cannot be converted with %s without losing state", transform.Path, transform.Op)
		}
		switch transform.Op {
		case "index_by":
			array, ok := value.([]interface{})
			if !ok || transform.Key == "" {
				return bad()
			}
			index := make(map[string]interface{}, len(array))
			for _, row := range array {
				object, ok := row.(map[string]interface{})
				if !ok {
					return bad()
				}
				name, ok := object[transform.Key].(string)
				if !ok || name == "" {
					return bad()
				}
				if _, duplicate := index[name]; duplicate {
					return fmt.Errorf("Field %s contains duplicate index keys", transform.Path)
				}
				child := make(map[string]interface{}, len(object)-1)
				for field, value := range object {
					if field != transform.Key {
						child[field] = value
					}
				}
				index[name] = child
			}
			parent[key] = index
		case "wrap":
			// A list of plain values becomes a list of objects that hold each value under Key, in order.
			array, ok := value.([]interface{})
			if !ok || transform.Key == "" {
				return bad()
			}
			wrapped := make([]interface{}, len(array))
			for i, item := range array {
				switch item.(type) {
				case nil, map[string]interface{}, []interface{}:
					return bad()
				}
				wrapped[i] = map[string]interface{}{transform.Key: item}
			}
			parent[key] = wrapped
		case "composite_id":
			// A legacy parent/child ID ("dashboard/chart") becomes the child's own last part.
			text, ok := value.(string)
			if !ok {
				return bad()
			}
			parts := strings.Split(text, "/")
			for _, part := range parts {
				if part == "" || strings.TrimSpace(part) != part {
					return bad()
				}
			}
			parent[key] = parts[len(parts)-1]
		case "json_decode":
			encoded, ok := value.(string)
			if !ok {
				return bad()
			}
			decoded, err := decodeMigrationJSON([]byte(encoded))
			if err != nil {
				return bad()
			}
			parent[key] = decoded
		case "singleton":
			array, ok := value.([]interface{})
			if !ok || len(array) > 1 {
				return bad()
			}
			parent[key] = nil
			if len(array) == 1 {
				parent[key] = array[0]
			}
		case "sentinel_null", "sentinel_omit":
			matches := reflect.DeepEqual(value, transform.Value)
			if !matches {
				number, numeric := value.(json.Number)
				sentinel, numericSentinel := transform.Value.(json.Number)
				if numeric && numericSentinel {
					actual, ok := new(big.Rat).SetString(number.String())
					expected, valid := new(big.Rat).SetString(sentinel.String())
					if !ok || !valid {
						return bad()
					}
					matches = actual.Cmp(expected) == 0
				}
			}
			if matches {
				if transform.Op == "sentinel_omit" {
					delete(parent, key)
				} else {
					parent[key] = nil
				}
			}
		case "seconds", "string_id":
			if transform.Op == "string_id" {
				if text, ok := value.(string); ok {
					if text == "" {
						return bad()
					}
					return nil
				}
			}
			number, ok := value.(json.Number)
			if !ok {
				return bad()
			}
			rational, ok := new(big.Rat).SetString(number.String())
			if !ok {
				return bad()
			}
			switch transform.Op {
			case "string_id":
				if !rational.IsInt() || rational.Sign() < 0 {
					return bad()
				}
				parent[key] = rational.Num().String()
			case "seconds":
				factorText := transform.Factor.String()
				if factorText == "" {
					factorText = "60"
				}
				factor, ok := new(big.Rat).SetString(factorText)
				if !ok || factor.Sign() <= 0 {
					return bad()
				}
				product := rational.Mul(rational, factor)
				// JSON decimals have only factors 2 and 5 in the denominator. Use their
				// exact decimal precision rather than a float64 (500 * 0.001 is 0.5).
				denominator := new(big.Int).Set(product.Denom())
				digits := 0
				for _, base := range []int64{2, 5} {
					count := 0
					for new(big.Int).Mod(denominator, big.NewInt(base)).Sign() == 0 {
						denominator.Div(denominator, big.NewInt(base))
						count++
					}
					if count > digits {
						digits = count
					}
				}
				parent[key] = json.Number(product.FloatString(digits))
			}
		default:
			return fmt.Errorf("Unsupported migration transform %s", transform.Op)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if (transform.Op == "index_by" || transform.Op == "wrap") && transform.To != "" && transform.To != transform.Path {
		return renameMigrationPath(state, transform.Path, transform.To, "")
	}
	return nil
}

// Unlike API refresh conversion, this never coerces primitives, drops object members, unwraps
// lists, or turns a conversion failure into null. Dynamic fields preserve their complete structure.
func strictMigrationValue(t tftypes.Type, value interface{}, path string) (tftypes.Value, error) {
	if value == nil {
		return tftypes.NewValue(t, nil), nil
	}
	bad := func() (tftypes.Value, error) {
		return tftypes.Value{}, fmt.Errorf("Field %s is incompatible with its target schema", path)
	}
	if t.Is(tftypes.DynamicPseudoType) {
		switch value.(type) {
		case string:
			t = tftypes.String
		case bool:
			t = tftypes.Bool
		case json.Number:
			t = tftypes.Number
		case map[string]interface{}:
			attributes := map[string]tftypes.Type{}
			for key := range value.(map[string]interface{}) {
				attributes[key] = tftypes.DynamicPseudoType
			}
			t = tftypes.Object{AttributeTypes: attributes}
		case []interface{}:
			elements := make([]tftypes.Type, len(value.([]interface{})))
			for i := range elements {
				elements[i] = tftypes.DynamicPseudoType
			}
			t = tftypes.Tuple{ElementTypes: elements}
		default:
			return bad()
		}
		// The concrete child types of dynamic objects/tuples must be inferred as well.
		return concreteMigrationValue(t, value, path)
	}
	switch {
	case t.Is(tftypes.String):
		if text, ok := value.(string); ok {
			return tftypes.NewValue(t, text), nil
		}
		return bad()
	case t.Is(tftypes.Bool):
		if boolean, ok := value.(bool); ok {
			return tftypes.NewValue(t, boolean), nil
		}
		return bad()
	case t.Is(tftypes.Number):
		if number, ok := value.(json.Number); ok {
			f, _, err := big.ParseFloat(number.String(), 10, 512, big.ToNearestEven)
			if err == nil && !f.IsInf() {
				return tftypes.NewValue(t, f), nil
			}
		}
		return bad()
	}
	switch typ := t.(type) {
	case tftypes.Object:
		object, ok := value.(map[string]interface{})
		if !ok {
			return bad()
		}
		for _, key := range migrationKeys(object) {
			if _, exists := typ.AttributeTypes[key]; !exists {
				return tftypes.Value{}, fmt.Errorf("Object %s contains a field with no declared target; refusing to discard it", path)
			}
		}
		attributes := make(map[string]tftypes.Value, len(typ.AttributeTypes))
		for _, key := range migrationKeys(typ.AttributeTypes) {
			child, err := strictMigrationValue(typ.AttributeTypes[key], object[key], path+"."+key)
			if err != nil {
				return tftypes.Value{}, err
			}
			attributes[key] = child
		}
		return migrationTypedValue(t, attributes, path)
	case tftypes.Map:
		object, ok := value.(map[string]interface{})
		if !ok {
			return bad()
		}
		attributes := make(map[string]tftypes.Value, len(object))
		for _, key := range migrationKeys(object) {
			child, err := strictMigrationValue(typ.ElementType, object[key], path+"[map value]")
			if err != nil {
				return tftypes.Value{}, err
			}
			attributes[key] = child
		}
		return migrationTypedValue(t, attributes, path)
	case tftypes.List, tftypes.Set, tftypes.Tuple:
		array, ok := value.([]interface{})
		if !ok {
			return bad()
		}
		if tuple, ok := t.(tftypes.Tuple); ok && len(tuple.ElementTypes) != len(array) {
			return bad()
		}
		elements := make([]tftypes.Value, len(array))
		for i, element := range array {
			var elementType tftypes.Type
			switch list := t.(type) {
			case tftypes.List:
				elementType = list.ElementType
			case tftypes.Set:
				elementType = list.ElementType
			case tftypes.Tuple:
				elementType = list.ElementTypes[i]
			}
			child, err := strictMigrationValue(elementType, element, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return tftypes.Value{}, err
			}
			elements[i] = child
		}
		return migrationTypedValue(t, elements, path)
	}
	return bad()
}

func concreteMigrationValue(t tftypes.Type, value interface{}, path string) (tftypes.Value, error) {
	switch typ := t.(type) {
	case tftypes.Object:
		object := value.(map[string]interface{})
		attributes := make(map[string]tftypes.Value, len(object))
		for _, key := range migrationKeys(object) {
			child, err := strictMigrationValue(tftypes.DynamicPseudoType, object[key], path+"[dynamic value]")
			if err != nil {
				return tftypes.Value{}, err
			}
			attributes[key], typ.AttributeTypes[key] = child, child.Type()
		}
		return migrationTypedValue(typ, attributes, path)
	case tftypes.Tuple:
		array := value.([]interface{})
		elements := make([]tftypes.Value, len(array))
		for i, element := range array {
			child, err := strictMigrationValue(tftypes.DynamicPseudoType, element, path+"[dynamic value]")
			if err != nil {
				return tftypes.Value{}, err
			}
			elements[i], typ.ElementTypes[i] = child, child.Type()
		}
		return migrationTypedValue(typ, elements, path)
	default:
		return strictMigrationValue(t, value, path)
	}
}

func migrationTypedValue(t tftypes.Type, value interface{}, path string) (tftypes.Value, error) {
	if err := tftypes.ValidateValue(t, value); err != nil {
		return tftypes.Value{}, fmt.Errorf("Field %s cannot be represented by the target collection type", path)
	}
	return tftypes.NewValue(t, value), nil
}
