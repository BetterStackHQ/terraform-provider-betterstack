package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The wire contract, expressed as DATA. Every value below is described by the API schema and shipped
// in bindings.go: the endpoint's real path, the key path each piece of a payload lives at, how the
// next page is addressed, where an error message hides, and how the credential is presented.
// Nothing in this file knows a product, a host, or a URL convention, which is what lets one provider
// speak several Better Stack APIs (and a new one) without a code change.

// endpoint is one real HTTP endpoint, transcribed from the API description rather than derived from
// a naming convention. Path is templated: /api/v2/status-pages/{status_page_id}/resources/{id}.
type endpoint struct {
	Method       string
	Path         string
	ListResponse bool
	Envelope     *envelope
	// Query is fixed query-string values every call to the endpoint sends.
	Query map[string]string
	// QueryVars maps a query param to the variable that fills it on each call: a natural key
	// (?email=) travels there, not in the path.
	QueryVars map[string]string
	// KeyField is the attribute a keyed-by endpoint addresses the object by, carried in the
	// variable of the same name; PathFields are path variables that carry attribute values.
	KeyField   string
	PathFields []string
	// NoSelector marks an endpoint that refuses the binding's tenant selector, such as a release
	// list where the same parameter names the team an SDK reports for: the provider's team isn't sent there.
	NoSelector bool
}

// envelope is a set of key paths into a payload. An EMPTY path means "at the root"; a NIL path means
// this envelope has no such place. That distinction is the whole trick: writes to our /api/v2 are
// flat because the description says request_path is empty, while a JSON:API-writing API would carry
// ["data", "attributes"] and get the wrapper, with no branch here.
// RelationshipsPath and IncludedPath describe sideloading: a value that arrives as a reference in
// the record's relationship map and is resolved against a pool of records at the document root. Nil
// on an API that has no such thing, which is how the walk below stays branch-free per app.
type envelope struct {
	ItemPath          []string
	ListPath          []string
	AttributesPath    []string
	IDPath            []string
	TypePath          []string
	RequestPath       []string
	RelationshipsPath []string
	IncludedPath      []string
}

// pagination says how the NEXT page is addressed: by an absolute URL in the body, by a cursor fed
// back as a request param, or not at all.
type pagination struct {
	Via   string
	Path  []string
	Param string
	// Numbered pages ("increment"): Path is where a page says how many records the whole list holds,
	// Size the query parameter for how many to answer, and Max the most it answers.
	Size string
	Max  int
}

// errorShape locates the human message and any attribute-scoped detail in a failure body.
type errorShape struct {
	MessagePath []string
	DetailPaths map[string][]string
	// RecordKeys are keys of a per-attribute message object that name the record itself (Rails'
	// base), so their messages read as plain sentences instead of "base ...".
	RecordKeys []string
}

// authScheme is how the credential is presented (bearer in an Authorization header, an API key in a
// named header, a query param).
type authScheme struct {
	Scheme   string
	Location string
	Name     string
}

// tenantSelector is how a credential that can reach several tenants picks one.
type tenantSelector struct {
	In   string
	Name string
}

// apiBinding is one resource's complete wire description.
type apiBinding struct {
	Resource string
	TypeName string
	IDKey    string
	BaseURL  string
	Auth     authScheme
	Selector *tenantSelector
	Envelope envelope
	// Fields the schema declares as sideloaded; nothing else is lifted out of a relationship map.
	Relationships   []string
	Pagination      pagination
	Errors          errorShape
	Create          *endpoint
	Read            *endpoint
	List            *endpoint
	Update          *endpoint
	Delete          *endpoint
	SingletonFields map[string]string
	ResponseFields  map[string]string
	QueryID         string
	MapList         *mapListProjection
	Reads           []supplementalRead
	DataSourceReads []supplementalRead
	// Waits are the conditions a create, update or delete settles on, as JSON (wait.go): the API
	// accepts the write at once and finishes the work later.
	Waits map[string]string
	// InPlace are the fields the API changes through actions of their own, not the update (in_place.go).
	InPlace []inPlaceChange
	// DeleteRetryCodes are the refusals of a delete that end by themselves, when what held the object
	// is gone (a destroyed server's reserved IP).
	DeleteRetryCodes []string
	// GoneField and GoneValues say which read means the object is gone for good, though the API still
	// answers it for a while (a server whose status is destroyed).
	GoneField  string
	GoneValues []string
	// GoneParent is the resource the object sits under, named by the path variable GoneParentVar, when
	// that resource's reads say it is gone: the object is gone with it.
	GoneParent    string
	GoneParentVar string
	// ReservedField is the attribute whose value a deleted object keeps for a while, so that creating one
	// with the same value is refused with ReservedMessage ({value} stands for it) until the delete is done (reserved.go).
	ReservedField   string
	ReservedMessage string
}

type supplementalRead struct {
	Endpoint *endpoint
	Fields   []string
}

// mapListProjection exposes a live, keyed collection without inventing a remote resource.
type mapListProjection struct {
	Filter string
	Keys   string
	Values string
	Sort   string
}

// apiValueShape retains field policy when Terraform must represent a collection dynamically.
type apiValueShape struct {
	ReadOnly  bool
	WriteOnly bool
	Identity  string
	Fields    map[string]apiValueShape
	Element   *apiValueShape
}

func projectAPIWrite(value interface{}, shape apiValueShape) interface{} {
	switch current := value.(type) {
	case map[string]interface{}:
		for key, child := range shape.Fields {
			if child.ReadOnly {
				delete(current, key)
				continue
			}
			if nested, ok := current[key]; ok {
				current[key] = projectAPIWrite(nested, child)
			}
		}
	case []interface{}:
		if shape.Element != nil {
			for index, item := range current {
				current[index] = projectAPIWrite(item, *shape.Element)
			}
		}
	}
	return value
}

// Preserve write-only values by declared child identity, never by list position.
func mergeAPIRead(value, prior interface{}, shape apiValueShape) interface{} {
	if shape.WriteOnly && prior != nil {
		return prior
	}
	switch current := value.(type) {
	case map[string]interface{}:
		if len(shape.Fields) == 0 {
			return value
		}
		previous, _ := prior.(map[string]interface{})
		out := make(map[string]interface{}, len(current))
		for key, item := range current {
			out[key] = item
		}
		for key, child := range shape.Fields {
			item, exists := current[key]
			if child.WriteOnly && previous[key] != nil {
				out[key] = previous[key]
			} else if exists {
				out[key] = mergeAPIRead(item, previous[key], child)
			}
		}
		return out
	case []interface{}:
		if shape.Element == nil {
			return value
		}
		byID := map[string]interface{}{}
		if shape.Identity != "" {
			if previous, ok := prior.([]interface{}); ok {
				for _, item := range previous {
					if object, ok := item.(map[string]interface{}); ok {
						if id := stringifyScalar(object[shape.Identity]); id != "" {
							byID[id] = item
						}
					}
				}
			}
		}
		out := make([]interface{}, len(current))
		for index, item := range current {
			var before interface{}
			if object, ok := item.(map[string]interface{}); ok && shape.Identity != "" {
				before = byID[stringifyScalar(object[shape.Identity])]
			}
			out[index] = mergeAPIRead(item, before, *shape.Element)
		}
		return out
	}
	return value
}

// Writable dynamic collections cannot contain computed-only keys in their configured shape.
// Their separate computed result attribute retains the complete server representation. Without a
// prior value (an import) there is no configured shape to follow, so keep what configuration can
// state: no computed-only keys and no null members, as the legacy state mover also projects.
func configuredAPIRead(value, prior interface{}, shape apiValueShape) interface{} {
	switch current := value.(type) {
	case map[string]interface{}:
		if prior == nil && shape.Fields != nil {
			out := make(map[string]interface{}, len(current))
			for key, item := range current {
				if child := shape.Fields[key]; item != nil && !child.ReadOnly {
					out[key] = configuredAPIRead(item, nil, child)
				}
			}
			return out
		}
		previous, known := prior.(map[string]interface{})
		if !known || shape.Fields == nil {
			return value
		}
		out := make(map[string]interface{}, len(previous))
		for key, before := range previous {
			if item, exists := current[key]; exists {
				out[key] = configuredAPIRead(item, before, shape.Fields[key])
			} else {
				out[key] = before
			}
		}
		return out
	case []interface{}:
		previous, known := prior.([]interface{})
		if !known && prior != nil || shape.Element == nil {
			return value
		}
		out := make([]interface{}, len(current))
		for index, item := range current {
			var before interface{}
			if index < len(previous) {
				before = previous[index]
			}
			out[index] = configuredAPIRead(item, before, *shape.Element)
		}
		return out
	}
	return value
}

// --- payload plumbing -------------------------------------------------------------------------

// digPath walks a decoded JSON value down a key path. A numeric segment indexes an array, so an
// error shape whose message lives at ["errors", "0"] resolves without a special case.
func digPath(v interface{}, path []string) interface{} {
	for _, key := range path {
		if v == nil {
			return nil
		}
		switch node := v.(type) {
		case map[string]interface{}:
			v = node[key]
		case []interface{}:
			i, err := strconv.Atoi(key)
			if err != nil || i < 0 || i >= len(node) {
				return nil
			}
			v = node[i]
		default:
			return nil
		}
	}
	return v
}

// wrapPath nests a value under a path of keys; an empty path returns it untouched, which is exactly
// how a flat request body is spelled.
func wrapPath(path []string, leaf interface{}) interface{} {
	out := leaf
	for i := len(path) - 1; i >= 0; i-- {
		out = map[string]interface{}{path[i]: out}
	}
	return out
}

// setPath writes a value into an already-built nested map, creating objects along the way.
func setPath(root interface{}, path []string, value interface{}) {
	if len(path) == 0 {
		return
	}
	node, ok := root.(map[string]interface{})
	if !ok {
		return
	}
	for _, key := range path[:len(path)-1] {
		next, ok := node[key].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			node[key] = next
		}
		node = next
	}
	node[path[len(path)-1]] = value
}

// encodeRequest renders a flat attribute map as the body this API accepts. When the envelope also
// names a type key, it is stamped as a sibling of the attributes rather than assumed.
func (e envelope) encodeRequest(typeName string, attrs map[string]interface{}) interface{} {
	if len(e.RequestPath) == 0 {
		return attrs
	}
	body := wrapPath(e.RequestPath, attrs)
	if len(e.TypePath) > 0 {
		parent := append([]string{}, e.RequestPath[:len(e.RequestPath)-1]...)
		setPath(body, append(parent, e.TypePath...), typeName)
	}
	return body
}

// flattenRecord turns one record into a flat attribute map keyed the way the schema names fields,
// lifting the record's id in under idKey and resolving the declared sideloaded fields against the
// document's included records. An attribute of the same name always wins: it is the direct value.
func (e envelope) flattenRecord(record interface{}, idKey string, rels []string, included []interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	if attrs, ok := digPath(record, e.AttributesPath).(map[string]interface{}); ok {
		for k, v := range attrs {
			out[k] = v
		}
	}
	if len(rels) > 0 && e.RelationshipsPath != nil {
		links, _ := digPath(record, e.RelationshipsPath).(map[string]interface{})
		for _, name := range rels {
			if _, taken := out[name]; taken {
				continue
			}
			if v, ok := e.resolveRelationship(links[name], included); ok {
				out[name] = v
			}
		}
	}
	// A list-only record has no path to its id: one of its attributes is the id.
	if id := digPath(record, e.IDPath); e.IDPath != nil && id != nil {
		out[idKey] = id
	}
	return out
}

// resolveRelationship turns {"data": ref} / {"data": [ref, ...]} into attribute maps. A ref is the
// JSON:API {id, type, meta} triple; its meta rides along and the included record's attributes win.
func (e envelope) resolveRelationship(link interface{}, included []interface{}) (interface{}, bool) {
	holder, ok := link.(map[string]interface{})
	if !ok {
		return nil, false
	}
	switch data := holder["data"].(type) {
	case map[string]interface{}:
		return e.resolveRef(data, included)
	case []interface{}:
		out := make([]interface{}, 0, len(data))
		for _, ref := range data {
			m, ok := ref.(map[string]interface{})
			if !ok {
				continue
			}
			if resolved, ok := e.resolveRef(m, included); ok {
				out = append(out, resolved)
			}
		}
		if len(out) == 0 {
			return nil, false
		}
		return out, true
	}
	return nil, false
}

func (e envelope) resolveRef(ref map[string]interface{}, included []interface{}) (map[string]interface{}, bool) {
	out := map[string]interface{}{}
	if meta, ok := ref["meta"].(map[string]interface{}); ok {
		for k, v := range meta {
			out[k] = v
		}
	}
	id, kind := stringifyScalar(digPath(ref, e.IDPath)), stringifyScalar(digPath(ref, e.TypePath))
	for _, rec := range included {
		if stringifyScalar(digPath(rec, e.IDPath)) != id {
			continue
		}
		if kind != "" && !strings.EqualFold(stringifyScalar(digPath(rec, e.TypePath)), kind) {
			continue
		}
		if attrs, ok := digPath(rec, e.AttributesPath).(map[string]interface{}); ok {
			for k, v := range attrs {
				out[k] = v
			}
		}
		break
	}
	if len(out) == 0 {
		return nil, false
	}
	if id != "" {
		out["id"] = id
	}
	return out, true
}

// includedPool is the document-level record pool sideloaded relationships resolve against.
func (e envelope) includedPool(doc interface{}) []interface{} {
	if e.IncludedPath == nil {
		return nil
	}
	pool, _ := digPath(doc, e.IncludedPath).([]interface{})
	return pool
}

func (e envelope) decodeItem(doc interface{}, idKey string, rels []string) map[string]interface{} {
	return e.flattenRecord(digPath(doc, e.ItemPath), idKey, rels, e.includedPool(doc))
}

func (e envelope) decodeList(doc interface{}, idKey string, rels []string) []map[string]interface{} {
	raw, ok := digPath(doc, e.ListPath).([]interface{})
	if !ok {
		return nil
	}
	included := e.includedPool(doc)
	out := make([]map[string]interface{}, 0, len(raw))
	for _, rec := range raw {
		out = append(out, e.flattenRecord(rec, idKey, rels, included))
	}
	return out
}

// --- errors ---------------------------------------------------------------------------------

// APIError is a failure decoded through the API's declared error shape: the message, the
// per-attribute messages when the message slot holds an object ({attr: [msg, ...]}, a validation
// failure), and any attribute-scoped detail the shape documents.
type APIError struct {
	Status  int
	Method  string
	URL     string
	Message string
	Fields  map[string][]string
	Details map[string]map[string]string
	Body    string
	// Record holds the messages about the record as a whole, in the API's order.
	Record []string
}

// apiErrorEntry is one sentence the API wrote, with the request attribute it names ("" when it
// names none).
type apiErrorEntry struct {
	Attr string
	Text string
}

// Entries lists every sentence in the failure, in a stable order: each message of each attribute,
// then each documented detail. A detail in list form ([attr, ...]) carries no sentence of its own,
// so it reads as the top-level message naming the attribute.
func (e *APIError) Entries() []apiErrorEntry {
	var out []apiErrorEntry
	for _, msg := range e.Record {
		out = append(out, apiErrorEntry{Text: msg})
	}
	for _, attr := range sortedKeys(e.Fields) {
		for _, msg := range e.Fields[attr] {
			out = append(out, apiErrorEntry{Attr: attr, Text: attr + " " + msg})
		}
	}
	for _, kind := range sortedKeys(e.Details) {
		for _, attr := range sortedKeys(e.Details[kind]) {
			out = append(out, apiErrorEntry{Attr: attr, Text: e.detailText(kind, attr, e.Details[kind][attr])})
		}
	}
	return out
}

func (e *APIError) detailText(kind, attr, detail string) string {
	if detail != "" {
		return attr + " " + detail
	}
	lead := e.Message
	if lead == "" {
		lead = strings.ReplaceAll(kind, "_", " ")
	}
	return strings.TrimRight(lead, ".:") + ": " + attr
}

func (e *APIError) Error() string {
	msg := e.Message
	entries := e.Entries()
	if msg == "" && len(entries) == 0 {
		msg = strings.TrimSpace(e.Body)
	}
	if msg == "" && len(entries) == 0 {
		msg = http.StatusText(e.Status)
	}
	parts := []string{strings.TrimSpace(fmt.Sprintf("%s %s: %d %s", e.Method, e.URL, e.Status, msg))}
	for _, entry := range entries {
		parts = append(parts, entry.Text)
	}
	return strings.Join(parts, "; ")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// parseAPIError decodes a failure body through the shape the API declares: the message comes from
// MessagePath, and each documented detail key yields attribute-scoped entries. Both the map form
// ({attr: msg}) and the list form ([attr, ...]) are tolerated, because real APIs use both.
func parseAPIError(shape errorShape, method, reqURL string, status int, body []byte) *APIError {
	apiErr := &APIError{Status: status, Method: method, URL: reqURL, Body: string(body), Details: map[string]map[string]string{}}
	var decoded interface{}
	if err := decodeNumbers(body, &decoded); err != nil {
		return apiErr
	}
	message := digPath(decoded, shape.MessagePath)
	if fields, ok := message.(map[string]interface{}); ok {
		apiErr.Fields = fieldMessages(fields)
		for _, key := range shape.RecordKeys {
			apiErr.Record = append(apiErr.Record, apiErr.Fields[key]...)
			delete(apiErr.Fields, key)
		}
	} else {
		apiErr.Message = joinMessages(message)
	}
	for _, kind := range sortedKeys(shape.DetailPaths) {
		entries := attributeMessages(digPath(decoded, shape.DetailPaths[kind]))
		if len(entries) > 0 {
			apiErr.Details[kind] = entries
		}
	}
	return apiErr
}

func joinMessages(v interface{}) string {
	switch msg := v.(type) {
	case string:
		return msg
	case []interface{}:
		parts := make([]string, 0, len(msg))
		for _, m := range msg {
			if s, ok := m.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "; ")
	case nil:
		return ""
	default:
		return stringifyScalar(v)
	}
}

// fieldMessages reads a validation failure keyed by attribute: each value is a message or a list of
// them, as Rails renders a record's errors.
func fieldMessages(fields map[string]interface{}) map[string][]string {
	out := map[string][]string{}
	for attr, v := range fields {
		switch msgs := v.(type) {
		case []interface{}:
			for _, m := range msgs {
				out[attr] = append(out[attr], stringifyScalar(m))
			}
		default:
			out[attr] = append(out[attr], stringifyScalar(msgs))
		}
	}
	return out
}

func attributeMessages(v interface{}) map[string]string {
	out := map[string]string{}
	switch detail := v.(type) {
	case map[string]interface{}:
		for attr, msg := range detail {
			out[attr] = joinMessages(msg)
		}
	case []interface{}:
		for _, attr := range detail {
			if s, ok := attr.(string); ok {
				out[s] = ""
			}
		}
	}
	return out
}

// --- the client -------------------------------------------------------------------------------

// Client speaks every described API: the binding carries the host, the credential presentation, the
// envelope and the pagination style, so one HTTP layer serves all of them. An endpoint override
// points everything at a single base URL (the acceptance tests' mock server). Network failures and
// 5xx responses are retried only for idempotent methods. A 429 is safe to replay for every method
// because the server did not admit the request.
type Client struct {
	reserved reservedValues
	token    string
	endpoint string
	tenants  map[string]string
	// baseURLVars fill a templated base URL's {variables}, from the provider settings.
	baseURLVars map[string]string
	http        *http.Client
	userAgent   string
	maxRetries  int
	maxBackoff  time.Duration
	minBackoff  time.Duration
	rateMu      sync.Mutex
	rate        float64
	burst       float64
	tokens      float64
	refilled    time.Time
}

// NewClient takes the complete User-Agent so a bridge built on this provider can name itself
// instead of posing as the Terraform provider; the provider passes its own product/version.
func NewClient(token, endpoint string, tenants map[string]string, userAgent string) *Client {
	c := &Client{
		token: token, endpoint: endpoint, tenants: tenants,
		http: &http.Client{}, userAgent: userAgent,
	}
	_ = c.configureControls(4, 10, 300, 60, 8, 0)
	return c
}

func (c *Client) configureControls(retries, waitMin, waitMax, timeout, rate, burst int64) error {
	for _, value := range []int64{retries, waitMin, waitMax, timeout, rate, burst} {
		if value < 0 || value > (1<<63-1)/int64(time.Second) {
			return errors.New("API controls must be nonnegative integers within duration limits")
		}
	}
	if waitMin > waitMax {
		return errors.New("api_retry_wait_min must not exceed api_retry_wait_max")
	}
	c.maxRetries = int(retries)
	c.minBackoff, c.maxBackoff = time.Duration(waitMin)*time.Second, time.Duration(waitMax)*time.Second
	c.http.Timeout = time.Duration(timeout) * time.Second
	if burst == 0 {
		burst = max(10, 2*rate)
	}
	c.rate, c.burst, c.tokens = float64(rate), float64(burst), float64(burst)
	c.refilled = time.Now()
	return nil
}

func (c *Client) waitRate(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		c.rateMu.Lock()
		if c.rate == 0 {
			c.rateMu.Unlock()
			return nil
		}
		now := time.Now()
		c.tokens = min(c.burst, c.tokens+now.Sub(c.refilled).Seconds()*c.rate)
		c.refilled = now
		if c.tokens >= 1 {
			c.tokens--
			c.rateMu.Unlock()
			return nil
		}
		delay := time.Duration((1 - c.tokens) / c.rate * float64(time.Second))
		c.rateMu.Unlock()
		if err := c.wait(ctx, delay); err != nil {
			return err
		}
	}
}

func (c *Client) baseURL(b *apiBinding) string {
	if c.endpoint != "" {
		return c.endpoint
	}
	out := b.BaseURL
	for name, value := range c.baseURLVars {
		out = strings.ReplaceAll(out, "{"+name+"}", value)
	}
	return out
}

var baseURLVariable = regexp.MustCompile(`\{([a-z0-9_]+)\}`)

// baseURLVariableValue is one host label: nothing that could point a request, and its token, elsewhere.
var baseURLVariableValue = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// checkBaseURL refuses, before any request, a base URL whose variable has no valid value.
func (c *Client) checkBaseURL(b *apiBinding) error {
	if c.endpoint != "" {
		return nil
	}
	for _, match := range baseURLVariable.FindAllStringSubmatch(b.BaseURL, -1) {
		name := match[1]
		value, ok := c.baseURLVars[name]
		env := "BETTERSTACK_" + strings.ToUpper(name)
		if !ok || value == "" {
			return fmt.Errorf("%s: set %s in the provider block or the %s environment variable", b.Resource, name, env)
		}
		if !baseURLVariableValue.MatchString(value) {
			return fmt.Errorf("%s: %s %q must be one host label, such as the value the API's docs show", b.Resource, name, value)
		}
	}
	return nil
}

// expandPath substitutes the endpoint's named path variables. A variable with no value is left in
// place so the resulting 404 names the missing piece instead of silently hitting a collection.
func expandPath(path string, vars map[string]string) string {
	out := path
	for name, value := range vars {
		out = strings.ReplaceAll(out, "{"+name+"}", url.PathEscape(value))
	}
	return out
}

// teamOverride is a resource's own team, which its requests select instead of the provider's.
// fromState marks a team read back from the API rather than configured: it only steers requests
// away from a provider team that points elsewhere, so with no provider team it selects nothing.
type teamOverride struct {
	team      string
	fromState bool
}

type teamKey struct{}

// inTeam makes the requests under ctx select team instead of the provider's. An empty team keeps
// the provider's.
func inTeam(ctx context.Context, team string) context.Context {
	if team == "" {
		return ctx
	}
	return context.WithValue(ctx, teamKey{}, teamOverride{team: team})
}

// inStateTeam makes the requests under ctx, which reach an existing object by its ID, select the team
// the object was read in, but only instead of the provider's team. Without one they select no team: the
// ID already fixes the team, and a team token refuses any team selector.
func inStateTeam(ctx context.Context, team string) context.Context {
	if team == "" {
		return ctx
	}
	return context.WithValue(ctx, teamKey{}, teamOverride{team: team, fromState: true})
}

// refusedWithoutTeam says the API refused a request for selecting no team (a token that reaches
// several teams must name one): a client error short of not found. Every other failure is reported.
func refusedWithoutTeam(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && (apiErr.Status == http.StatusBadRequest || apiErr.Status == http.StatusForbidden || apiErr.Status == http.StatusUnprocessableEntity)
}

// inNoTeam makes the requests under ctx select no team, so a read by ID finds the object in
// whichever team the token reaches.
func inNoTeam(ctx context.Context) context.Context {
	return context.WithValue(ctx, teamKey{}, teamOverride{})
}

func (c *Client) requestURL(ctx context.Context, b *apiBinding, ep *endpoint, vars map[string]string, query url.Values) string {
	q := url.Values{}
	for k, v := range query {
		q[k] = v
	}
	if b.Selector != nil && b.Selector.In == "query" && !ep.NoSelector {
		v := c.tenants[b.Selector.Name]
		if o, ok := ctx.Value(teamKey{}).(teamOverride); ok && !(o.fromState && v == "") {
			v = o.team
		}
		if v != "" {
			q.Set(b.Selector.Name, v)
		}
	}
	reqURL := c.baseURL(b) + expandPath(ep.Path, vars)
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}
	return reqURL
}

func (c *Client) applyAuth(b *apiBinding, req *http.Request) {
	if b.Auth.Scheme == "" || b.Auth.Scheme == "none" {
		return
	}
	value := c.token
	if b.Auth.Scheme == "bearer" {
		value = "Bearer " + c.token
	}
	switch b.Auth.Location {
	case "query":
		q := req.URL.Query()
		q.Set(b.Auth.Name, c.token)
		req.URL.RawQuery = q.Encode()
	default:
		name := b.Auth.Name
		if name == "" {
			name = "Authorization"
		}
		req.Header.Set(name, value)
	}
}

func decodeNumbers(raw []byte, into interface{}) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	return dec.Decode(into)
}

func retryableMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func retryableResponse(method string, status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	return retryableMethod(method) && status >= 500
}

func (c *Client) do(ctx context.Context, b *apiBinding, method, reqURL string, payload []byte) ([]byte, int, error) {
	if b.Auth.Scheme != "" && b.Auth.Scheme != "none" && c.token == "" {
		return nil, 0, errors.New("set api_token or BETTERSTACK_API_TOKEN before calling authenticated endpoints")
	}
	if err := c.checkBaseURL(b); err != nil {
		return nil, 0, err
	}
	var lastErr error
	retryable := retryableMethod(method)
	for attempt := 0; ; attempt++ {
		if err := c.waitRate(ctx); err != nil {
			return nil, 0, err
		}
		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
		if err != nil {
			return nil, 0, err
		}
		c.applyAuth(b, req)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			if !retryable {
				return nil, 0, fmt.Errorf("%s %s: %w", method, reqURL, lastErr)
			}
			if attempt >= c.maxRetries {
				return nil, 0, fmt.Errorf("%s %s: giving up after %d attempts: %w", method, reqURL, attempt+1, lastErr)
			}
			if waitErr := c.wait(ctx, c.backoff(attempt, "")); waitErr != nil {
				return nil, 0, waitErr
			}
			continue
		}
		raw, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, resp.StatusCode, readErr
		}
		if retryableResponse(method, resp.StatusCode) && attempt < c.maxRetries {
			if waitErr := c.wait(ctx, c.backoff(attempt, resp.Header.Get("Retry-After"))); waitErr != nil {
				return nil, 0, waitErr
			}
			continue
		}
		if resp.StatusCode >= 400 {
			return raw, resp.StatusCode, parseAPIError(b.Errors, method, reqURL, resp.StatusCode, raw)
		}
		return raw, resp.StatusCode, nil
	}
}

// backoff is exponential (1s, 2s, 4s, ...) but a server-provided Retry-After (seconds) wins; both
// are capped at maxBackoff.
func (c *Client) backoff(attempt int, retryAfter string) time.Duration {
	wait := c.minBackoff
	for i := 0; i < attempt && wait < c.maxBackoff; i++ {
		if wait > c.maxBackoff/2 {
			wait = c.maxBackoff
			break
		}
		wait *= 2
	}
	if retryAfter != "" {
		if secs, err := strconv.ParseInt(strings.TrimSpace(retryAfter), 10, 64); err == nil && secs >= 0 {
			if secs > int64(c.maxBackoff/time.Second) {
				return c.maxBackoff
			}
			wait = time.Duration(secs) * time.Second
		}
	}
	if wait > c.maxBackoff {
		wait = c.maxBackoff
	}
	return wait
}

func (c *Client) wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *Client) call(ctx context.Context, b *apiBinding, ep *endpoint, vars map[string]string, attrs map[string]interface{}) (map[string]interface{}, int, error) {
	if ep == nil {
		return nil, 0, fmt.Errorf("%s: the API describes no such endpoint", b.Resource)
	}
	e := b.Envelope
	if ep.Envelope != nil {
		e = *ep.Envelope
	}
	var payload []byte
	if attrs != nil {
		var err error
		projected := make(map[string]interface{}, len(attrs))
		for key, value := range attrs {
			projected[key] = value
		}
		for array, scalar := range b.SingletonFields {
			if value, ok := projected[scalar]; ok {
				if _, exists := projected[array]; exists {
					return nil, 0, fmt.Errorf("configure %s only, not %s", scalar, array)
				}
				projected[array] = []interface{}{value}
				delete(projected, scalar)
			}
		}
		payload, err = json.Marshal(e.encodeRequest(b.TypeName, projected))
		if err != nil {
			return nil, 0, err
		}
	}
	var query url.Values
	if len(ep.Query) > 0 || len(ep.QueryVars) > 0 {
		query = url.Values{}
		for k, v := range ep.Query {
			query.Set(k, v)
		}
		for k, name := range ep.QueryVars {
			query.Set(k, vars[name])
		}
	}
	reqURL := c.requestURL(ctx, b, ep, vars, query)
	raw, status, err := c.do(ctx, b, ep.Method, reqURL, payload)
	if err != nil {
		return nil, status, err
	}
	if status == http.StatusNoContent || len(raw) == 0 {
		return nil, status, nil
	}
	var doc interface{}
	if err := decodeNumbers(raw, &doc); err != nil {
		return nil, status, fmt.Errorf("%s %s: decoding response: %w", ep.Method, reqURL, err)
	}
	var out map[string]interface{}
	if ep.ListResponse {
		items := e.decodeList(doc, b.IDKey, b.Relationships)
		if len(items) != 1 {
			return nil, status, fmt.Errorf("expected exactly one response object, received %d", len(items))
		}
		out = items[0]
	} else {
		item := digPath(doc, e.ItemPath)
		// A keyed create can answer with no object at all (someone already in the organization joins
		// the team at once, with no invitation to return): the caller reads it by its key.
		if item == nil && ep == b.Create && b.Read != nil && b.Read.KeyField != "" {
			return nil, status, nil
		}
		if _, ok := item.(map[string]interface{}); !ok {
			return nil, status, errors.New("expected an object response")
		}
		out = e.decodeItem(doc, b.IDKey, b.Relationships)
	}
	for array, scalar := range b.SingletonFields {
		if values, exists := out[array]; exists {
			items, ok := values.([]interface{})
			if !ok || len(items) != 1 {
				return nil, status, fmt.Errorf("expected one value for %s", array)
			}
			if existing, ok := out[scalar]; ok && stringifyScalar(existing) != stringifyScalar(items[0]) {
				return nil, status, fmt.Errorf("conflicting singleton response field %s", scalar)
			}
			out[scalar] = items[0]
			delete(out, array)
		}
	}
	for source, target := range b.ResponseFields {
		if value, ok := out[source]; ok {
			out[target] = value
		}
	}
	return out, status, nil
}

func (c *Client) Create(ctx context.Context, b *apiBinding, vars map[string]string, attrs map[string]interface{}) (map[string]interface{}, error) {
	out, _, err := c.callCreate(ctx, b, vars, attrs)
	if err != nil {
		return out, err
	}
	return c.settle(ctx, b, "create", vars, attrs, out)
}

func (c *Client) Get(ctx context.Context, b *apiBinding, vars map[string]string) (map[string]interface{}, bool, error) {
	out, status, err := c.call(ctx, b, b.Read, vars, nil)
	if status == http.StatusNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if err := c.supplement(ctx, b, vars, out, b.Reads); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (c *Client) GetDataSource(ctx context.Context, b *apiBinding, vars map[string]string) (map[string]interface{}, bool, error) {
	out, found, err := c.Get(ctx, b, vars)
	if err != nil || !found {
		return out, found, err
	}
	if err := c.supplement(ctx, b, vars, out, b.DataSourceReads); err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (c *Client) supplement(ctx context.Context, b *apiBinding, vars map[string]string, out map[string]interface{}, reads []supplementalRead) error {
	for _, read := range reads {
		extra, _, err := c.call(ctx, b, read.Endpoint, vars, nil)
		if err != nil {
			return err
		}
		for _, field := range read.Fields {
			value, ok := extra[field]
			if !ok {
				return fmt.Errorf("supplemental read omitted %s", field)
			}
			out[field] = value
		}
	}
	return nil
}

func (c *Client) Update(ctx context.Context, b *apiBinding, vars map[string]string, attrs map[string]interface{}) (map[string]interface{}, error) {
	out, _, err := c.call(ctx, b, b.Update, vars, attrs)
	if err != nil {
		return out, err
	}
	return c.settle(ctx, b, "update", vars, attrs, out)
}

// gone is true when the object answered is one that no longer exists.
func (b *apiBinding) gone(row map[string]interface{}) bool {
	value, ok := row[b.GoneField].(string)
	return b.GoneField != "" && ok && containsString(b.GoneValues, value)
}

func notFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
}

// parentGone is true only on a definite answer: the parent isn't found, or its status says it is gone. A
// parent that can't be read leaves its child as it is.
func (c *Client) parentGone(ctx context.Context, b *apiBinding, vars map[string]string) bool {
	parent := bindingFor(b.GoneParent)
	if b.GoneParent == "" || parent == nil || vars[b.GoneParentVar] == "" {
		return false
	}
	row, found, err := c.Get(ctx, parent, map[string]string{recordVar(parent.Read): vars[b.GoneParentVar]})
	return err == nil && (!found || parent.gone(row))
}

// Delete treats 404 as success: the object is already gone, which is the desired end state.
func (c *Client) Delete(ctx context.Context, b *apiBinding, vars map[string]string) error {
	out, status, err := c.callDelete(ctx, b, vars)
	if status == http.StatusNotFound {
		return nil
	}
	if err != nil {
		// Refused because it is gone already, though it still answers.
		if conflicts(err) && b.GoneField != "" {
			if row, found, readErr := c.Get(ctx, b, vars); readErr == nil && (!found || b.gone(row)) {
				return nil
			}
		}
		if conflicts(err) && c.parentGone(ctx, b, vars) {
			return nil
		}
		return err
	}
	_, err = c.settle(ctx, b, "delete", vars, nil, out)
	return err
}

func (c *Client) Query(ctx context.Context, b *apiBinding, filters []string) (map[string]interface{}, error) {
	if b.Read == nil || b.MapList == nil {
		return nil, errors.New("query projection is not configured")
	}
	raw, _, err := c.do(ctx, b, b.Read.Method, c.requestURL(ctx, b, b.Read, nil, nil), nil)
	if err != nil {
		return nil, err
	}
	var doc interface{}
	if err := decodeNumbers(raw, &doc); err != nil {
		return nil, err
	}
	e := b.Envelope
	if b.Read.Envelope != nil {
		e = *b.Read.Envelope
	}
	groups, ok := digPath(doc, e.ItemPath).(map[string]interface{})
	if !ok {
		return nil, errors.New("expected an object of lists")
	}
	for _, filter := range filters {
		if _, ok := groups[filter]; !ok {
			return nil, fmt.Errorf("unknown collection filter %q", filter)
		}
	}
	keys, values := []interface{}{}, []interface{}{}
	for _, key := range sortedKeys(groups) {
		keys = append(keys, key)
		selected := len(filters) == 0
		for _, filter := range filters {
			if key == filter {
				selected = true
			}
		}
		items, ok := groups[key].([]interface{})
		if !ok {
			return nil, fmt.Errorf("collection %q is not a list", key)
		}
		if selected {
			values = append(values, items...)
		}
	}
	if b.MapList.Sort == "ip" {
		addresses := make([]netip.Addr, len(values))
		for i, value := range values {
			text, ok := value.(string)
			if !ok {
				return nil, errors.New("IP collection contains a non-string value")
			}
			address, err := netip.ParseAddr(text)
			if err != nil {
				return nil, errors.New("IP collection contains an invalid address")
			}
			addresses[i] = address
		}
		sort.SliceStable(addresses, func(i, j int) bool { return addresses[i].Compare(addresses[j]) < 0 })
		for i, address := range addresses {
			values[i] = address.String()
		}
	}
	return map[string]interface{}{b.IDKey: b.QueryID, b.MapList.Keys: keys, b.MapList.Values: values}, nil
}

// List fetches the whole collection, following the pagination style the API declares (an absolute
// next URL, or a cursor echoed back as a request param) and merging the pages.
// degrade: a style this client does not implement yields the first page only, rather than a
// silently truncated result that pretends to be complete.
func (c *Client) List(ctx context.Context, b *apiBinding, vars map[string]string, query url.Values) ([]map[string]interface{}, error) {
	if b.List == nil {
		return nil, fmt.Errorf("%s: the API describes no list endpoint", b.Resource)
	}
	if b.Pagination.Via == "increment" {
		paged := url.Values{}
		for k, v := range query {
			paged[k] = v
		}
		paged.Set(b.Pagination.Param, "1")
		paged.Set(b.Pagination.Size, strconv.Itoa(b.Pagination.Max))
		query = paged
	}
	reqURL := c.requestURL(ctx, b, b.List, vars, query)
	var out []map[string]interface{}
	read := 0
	for page := 0; page < 100; page++ {
		raw, _, err := c.do(ctx, b, b.List.Method, reqURL, nil)
		if err != nil {
			return nil, err
		}
		var doc interface{}
		if err := decodeNumbers(raw, &doc); err != nil {
			return nil, fmt.Errorf("GET %s: decoding response: %w", reqURL, err)
		}
		e := b.Envelope
		if b.List.Envelope != nil {
			e = *b.List.Envelope
		}
		rows, ok := digPath(doc, e.ListPath).([]interface{})
		if !ok {
			return nil, errors.New("expected a list response")
		}
		read += len(rows)
		out = append(out, e.decodeList(doc, b.IDKey, b.Relationships)...)
		if b.Pagination.Via == "increment" {
			total, known := waitNumber(digPath(doc, b.Pagination.Path))
			if !known {
				return nil, fmt.Errorf("GET %s: the response doesn't say how many there are; refusing incomplete results", reqURL)
			}
			if float64(read) >= total {
				return out, nil
			}
			if len(rows) == 0 {
				return nil, fmt.Errorf("GET %s: read %d of %v; refusing incomplete results", reqURL, read, total)
			}
			next, err := url.Parse(reqURL)
			if err != nil {
				return nil, err
			}
			q := next.Query()
			q.Set(b.Pagination.Param, strconv.Itoa(page+2))
			next.RawQuery = q.Encode()
			reqURL = next.String()
			continue
		}
		next, ok, err := nextPageURL(b, doc, reqURL, c.baseURL(b))
		if err != nil {
			return nil, err
		}
		if !ok {
			return out, nil
		}
		reqURL = next
	}
	return nil, errors.New("pagination exceeded 100 pages; refusing incomplete results")
}

func nextPageURL(b *apiBinding, doc interface{}, current, endpointBase string) (string, bool, error) {
	token := stringifyScalar(digPath(doc, b.Pagination.Path))
	via := b.Pagination.Via
	if token == "" && (via == "url" || via == "param") {
		return "", false, nil
	}
	switch via {
	case "url":
		currentURL, err := url.Parse(current)
		if err != nil {
			return "", false, fmt.Errorf("invalid current pagination URL %q: %w", current, err)
		}
		nextURL, err := url.Parse(token)
		if err != nil {
			return "", false, fmt.Errorf("invalid pagination next URL %q: %w", token, err)
		}
		resolved := currentURL.ResolveReference(nextURL)
		baseURL, err := url.Parse(endpointBase)
		if err != nil {
			return "", false, fmt.Errorf("invalid endpoint base URL %q: %w", endpointBase, err)
		}
		nextOrigin, err := normalizedOrigin(resolved)
		if err != nil {
			return "", false, fmt.Errorf("invalid pagination next URL %q: %w", token, err)
		}
		baseOrigin, err := normalizedOrigin(baseURL)
		if err != nil {
			return "", false, fmt.Errorf("invalid endpoint base URL %q: %w", endpointBase, err)
		}
		if nextOrigin != baseOrigin {
			return "", false, fmt.Errorf("refusing cross-origin pagination next URL %q: origin %s differs from endpoint origin %s", token, nextOrigin, baseOrigin)
		}
		if resolved.String() == current {
			return "", false, errors.New("pagination next URL repeated the current page")
		}
		return resolved.String(), true, nil
	case "param":
		parsed, err := url.Parse(current)
		if err != nil {
			return "", false, nil
		}
		q := parsed.Query()
		if q.Get(b.Pagination.Param) == token {
			return "", false, errors.New("pagination cursor repeated the current page")
		}
		q.Set(b.Pagination.Param, token)
		parsed.RawQuery = q.Encode()
		return parsed.String(), true, nil
	case "", "none":
		return "", false, nil
	default:
		return "", false, fmt.Errorf("%s: lists page by %q, which the provider doesn't follow", b.Resource, b.Pagination.Via)
	}
}

func normalizedOrigin(u *url.URL) (string, error) {
	scheme := strings.ToLower(u.Scheme)
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if scheme == "" || host == "" {
		return "", errors.New("URL must include a scheme and host")
	}
	port := u.Port()
	if port == "" {
		switch scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return "", fmt.Errorf("unsupported URL scheme %q", scheme)
		}
	}
	return scheme + "://" + net.JoinHostPort(host, port), nil
}

// FindInList emulates a member GET for a resource the API gives no show endpoint: list (optionally
// narrowed by filters) and match the id client-side.
func (c *Client) FindInList(ctx context.Context, b *apiBinding, vars map[string]string, id string, query url.Values) (map[string]interface{}, bool, error) {
	records, err := c.List(ctx, b, vars, query)
	if err != nil {
		return nil, false, err
	}
	for _, rec := range records {
		if recID, ok := rec[b.IDKey]; ok && stringifyScalar(recID) == id {
			return rec, true, nil
		}
	}
	return nil, false, nil
}

// idFromResponse extracts the created object's id, guarding against responses without one.
func idFromResponse(out map[string]interface{}, idKey string) (string, error) {
	raw, ok := out[idKey]
	if !ok || raw == nil {
		return "", errors.New("API response did not include a resource id")
	}
	id := stringifyScalar(raw)
	if id == "" {
		return "", errors.New("API response did not include a resource id")
	}
	return id, nil
}
