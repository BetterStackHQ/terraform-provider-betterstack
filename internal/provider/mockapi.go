package provider

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// MockAPI is an in-memory server built FROM THE BINDINGS: it routes on the described endpoints and
// speaks each API's declared envelope, so a provider that wrapped a write the API takes flat (or
// guessed a path) fails here rather than in production. Deterministic and offline. It is not a
// test file so the bridges built on this provider (see pkg/apitest) can run against it too.
type MockAPI struct {
	mu         sync.Mutex
	nextID     int
	store      map[string]map[string]map[string]interface{}
	order      map[string][]string
	inject     map[string]map[string]interface{}
	itemInject map[string]map[string]map[string]interface{}
	itemIDs    map[string][]string
	redact     map[string][]string
	fail       map[string]mockFailure
	failPath   map[string]mockFailure
	lastRaw    map[string][]byte
	afterWrite map[string]func(map[string]interface{})
	agents     []string
	requests   []string
	url        string
	server     *http.Server
}

type mockFailure struct {
	status int
	body   string
}

// mockRoute is one described endpoint plus the CRUD meaning the binding gives it. Slot, not path
// shape, decides the semantics, so a singleton whose read and create share a path still works.
type mockRoute struct {
	binding *apiBinding
	slot    string
	ep      *endpoint
	fields  []string
	change  *inPlaceChange
}

func mockRoutes() []mockRoute {
	var routes []mockRoute
	for _, b := range allBindings {
		for _, pair := range []struct {
			slot string
			ep   *endpoint
		}{{"read", b.Read}, {"update", b.Update}, {"delete", b.Delete}, {"create", b.Create}, {"list", b.List}} {
			// A create that replaces the existing object also serves as the update; one route
			// answers it, as a create that replaces whatever the collection held.
			if pair.slot == "update" && replacingCreate(b) {
				continue
			}
			if pair.ep != nil {
				routes = append(routes, mockRoute{binding: b, slot: pair.slot, ep: pair.ep})
			}
		}
		for i := range b.InPlace {
			change := &b.InPlace[i]
			for _, ep := range []*endpoint{change.Set, change.Unset} {
				if ep != nil {
					routes = append(routes, mockRoute{binding: b, slot: "in_place", ep: ep, change: change})
				}
			}
		}
		for _, reads := range [][]supplementalRead{b.Reads, b.DataSourceReads} {
			for _, read := range reads {
				routes = append(routes, mockRoute{binding: b, slot: "supplemental", ep: read.Endpoint, fields: read.Fields})
			}
		}
	}
	return routes
}

// replacingCreate is true when the binding's update is its create (the schema's replaces_existing).
func replacingCreate(b *apiBinding) bool {
	return b.Create != nil && b.Update != nil && b.Create.Method == b.Update.Method && b.Create.Path == b.Update.Path
}

// storeName is the collection a binding's records live in: its own, or, for a resource without a
// create that reads another resource's endpoint (a source's AWS account reads the source), that
// resource's, so both see the same records.
func storeName(b *apiBinding) string {
	if b.Create != nil || b.Read == nil {
		return b.Resource
	}
	for _, other := range allBindings {
		if other != b && other.Create != nil && other.Read != nil && other.BaseURL == b.BaseURL && other.Read.Path == b.Read.Path {
			return other.Resource
		}
	}
	return b.Resource
}

// bindingFor finds the binding by the resource name the bindings spell (the Terraform type without
// its provider prefix), so a caller outside the package never needs a binding pointer.
func bindingFor(resource string) *apiBinding {
	for _, b := range allBindings {
		if b.Resource == resource {
			return b
		}
	}
	return nil
}

// matchPath compares a templated path with a concrete one, returning the bound variables.
func matchPath(template, actual string) (map[string]string, bool) {
	want := strings.Split(strings.Trim(template, "/"), "/")
	got := strings.Split(strings.Trim(actual, "/"), "/")
	if len(want) != len(got) {
		return nil, false
	}
	vars := map[string]string{}
	for i, seg := range want {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			vars[strings.Trim(seg, "{}")] = got[i]
			continue
		}
		if seg != got[i] {
			return nil, false
		}
	}
	return vars, true
}

// recordVar is the trailing path variable of a member endpoint, i.e. the one carrying the object's
// own id. A singleton (addressed entirely by its parent) has none.
func recordVar(ep *endpoint) string {
	segs := strings.Split(strings.Trim(ep.Path, "/"), "/")
	last := segs[len(segs)-1]
	if strings.HasPrefix(last, "{") && strings.HasSuffix(last, "}") {
		return strings.Trim(last, "{}")
	}
	return ""
}

// NewMockAPI starts the server; the caller owns it and must Close it. It listens through net and
// http.Server rather than httptest, because this file ships in the provider binary (pkg/apitest
// re-exports it) and httptest would register its flag and init there.
func NewMockAPI() *MockAPI {
	m := &MockAPI{
		store:      map[string]map[string]map[string]interface{}{},
		order:      map[string][]string{},
		inject:     map[string]map[string]interface{}{},
		itemInject: map[string]map[string]map[string]interface{}{},
		itemIDs:    map[string][]string{},
		redact:     map[string][]string{},
		fail:       map[string]mockFailure{},
		failPath:   map[string]mockFailure{},
		lastRaw:    map[string][]byte{},
		afterWrite: map[string]func(map[string]interface{}){},
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("mock API: listen: " + err.Error())
	}
	m.url = "http://" + listener.Addr().String()
	m.server = &http.Server{Handler: http.HandlerFunc(m.handle)}
	go func() { _ = m.server.Serve(listener) }()
	return m
}

// Close stops the listener and every open connection.
func (m *MockAPI) Close() { _ = m.server.Close() }

// URL is the base URL to point the provider's endpoint override at.
func (m *MockAPI) URL() string { return m.url }

// InjectOnCreate adds server-assigned attributes to every record created for a resource.
func (m *MockAPI) InjectOnCreate(resource string, attrs map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inject[resource] = attrs
}

// InjectItemFields adds server-assigned fields to each item of the named collection attributes
// on every write, the way the real API fills in ids on nested objects.
func (m *MockAPI) InjectItemFields(resource string, fields map[string]map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.itemInject[resource] = fields
}

// AssignItemIDs gives every item of the named nested collections that arrives without an id a
// fresh one on each write, the way the real API numbers new nested records while keeping the ids
// it is sent. A path names a collection attribute, dotted for one inside another's items
// ("sections.resources").
func (m *MockAPI) AssignItemIDs(resource string, paths ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.itemIDs[resource] = paths
}

func (m *MockAPI) applyItemInjections(b *apiBinding, attrs map[string]interface{}) {
	for field, additions := range m.itemInject[b.Resource] {
		items, _ := attrs[field].([]interface{})
		for _, item := range items {
			if object, ok := item.(map[string]interface{}); ok {
				for key, value := range additions {
					object[key] = value
				}
			}
		}
	}
	for _, path := range m.itemIDs[b.Resource] {
		m.assignItemIDs(attrs, strings.Split(path, "."))
	}
}

func (m *MockAPI) assignItemIDs(object map[string]interface{}, path []string) {
	items, _ := object[path[0]].([]interface{})
	for _, item := range items {
		child, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if len(path) > 1 {
			m.assignItemIDs(child, path[1:])
			continue
		}
		if id, _ := child["id"].(string); id == "" {
			m.nextID++
			child["id"] = strconv.Itoa(m.nextID)
		}
	}
}

// Seed places an empty record under an existing id, for resources the provider adopts rather
// than creates. vars are the parent path variables scoping the collection.
func (m *MockAPI) Seed(resource string, vars map[string]string, id string) error {
	return m.SeedRecord(resource, vars, id, map[string]interface{}{})
}

// SeedRecord places a record with these attributes under an existing id, for data sources that
// read what no resource creates.
func (m *MockAPI) SeedRecord(resource string, vars map[string]string, id string, attrs map[string]interface{}) error {
	b := bindingFor(resource)
	if b == nil {
		return fmt.Errorf("mock API: no binding describes %q", resource)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.storeKey(mockRoute{binding: b, slot: "read", ep: b.Read}, vars)
	if m.store[key] == nil {
		m.store[key] = map[string]map[string]interface{}{}
	}
	m.store[key][id] = attrs
	m.order[key] = append(m.order[key], id)
	return nil
}

// keepParent stores the parent of a created record, as it must exist, unless it is stored already.
func (m *MockAPI) keepParent(b *apiBinding, vars map[string]string) {
	parent, id := bindingFor(b.GoneParent), vars[b.GoneParentVar]
	if b.GoneParent == "" || parent == nil || id == "" {
		return
	}
	key := storeName(parent)
	if _, ok := m.store[key][id]; ok {
		return
	}
	if m.store[key] == nil {
		m.store[key] = map[string]map[string]interface{}{}
	}
	m.store[key][id] = map[string]interface{}{parent.GoneField: "running"}
	m.order[key] = append(m.order[key], id)
}

// AfterWrite runs fn on a resource's record after each create and update, for what the real API
// derives from a write (a role name from the role id a role change sets).
func (m *MockAPI) AfterWrite(resource string, fn func(map[string]interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.afterWrite[resource] = fn
}

// RedactInResponses drops attributes from every response for a resource, the way the API never
// echoes a secret back.
func (m *MockAPI) RedactInResponses(resource string, keys ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.redact[resource] = keys
}

// FailWrites answers every non-GET request for a resource with the given status and body.
func (m *MockAPI) FailWrites(resource string, status int, body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fail[resource] = mockFailure{status: status, body: body}
}

// FailRequest answers every request with this method and path with the given status and body, and
// stops doing so when status is 0.
func (m *MockAPI) FailRequest(method, path string, status int, body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if status == 0 {
		delete(m.failPath, method+" "+path)
		return
	}
	m.failPath[method+" "+path] = mockFailure{status: status, body: body}
}

// changeInPlace makes a described in-place change on the stored record, and settles it at once on the
// literals its wait expects. It is false when no stored record is meant.
func (m *MockAPI) changeInPlace(w http.ResponseWriter, route mockRoute, vars map[string]string, body map[string]interface{}, answered bool) bool {
	b, change := route.binding, route.change
	set := route.ep == change.Set
	id, item := vars[recordVar(b.Read)], vars[change.ItemVar]
	if change.Items && change.SelfParam == "" && set {
		item = stringifyScalar(body[change.Param])
	}
	if change.ItemList {
		if items := itemStrings(body[change.Param]); len(items) == 1 {
			item = items[0]
		}
	}
	if change.SelfParam != "" {
		id = stringifyScalar(body[change.SelfParam])
		if change.SelfList {
			if ids := itemStrings(body[change.SelfParam]); len(ids) == 1 {
				id = ids[0]
			}
		}
		if change.SelfVar != "" && vars[change.SelfVar] != "" {
			id = vars[change.SelfVar]
		}
	}
	// An item's own removal can leave this object unnamed (a reserved IP's unassign): whichever holds it.
	if change.Items && change.SelfParam != "" && id == "" {
		for recID, held := range m.store[storeName(b)] {
			if containsString(itemStrings(held[change.Field]), item) {
				id = recID
			}
		}
	}
	attrs, recID, ok := m.lookup(storeName(b), id)
	if !ok || id == "" {
		return false
	}
	wait := change.Wait
	switch {
	case change.Items:
		items := withoutString(itemStrings(attrs[change.Field]), item)
		if set {
			items = append(items, item)
		}
		list := []interface{}{}
		for _, one := range items {
			list = append(list, one)
		}
		attrs[change.Field] = list
	case set:
		attrs[change.Field] = body[change.Param]
	default:
		attrs[change.Field], wait = nil, change.UnsetWait
	}
	var policy waitPolicy
	if wait != "" && json.Unmarshal([]byte(wait), &policy) == nil && policy.Success != nil {
		for _, leaf := range append([]waitPredicate{*policy.Success}, policy.Success.All...) {
			if len(leaf.Path) == 1 && leaf.Expected != nil && leaf.Expected.Source == "literal" {
				attrs[leaf.Path[0]] = leaf.Expected.Value
			}
		}
	}
	if !answered {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(wrapPath(b.Envelope.ItemPath, m.record(b, recID, m.redacted(b, attrs))))
	}
	return true
}

// LastRequestBody returns the raw bytes of the last write to a resource: the envelope assertions
// read it, since "did the provider send a flat body" is exactly what the schema now decides.
func (m *MockAPI) LastRequestBody(resource string) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastRaw[resource]
}

// LastUserAgent returns the User-Agent of the most recent request, whatever it was routed to.
func (m *MockAPI) LastUserAgent() string {
	agents := m.UserAgents()
	if len(agents) == 0 {
		return ""
	}
	return agents[len(agents)-1]
}

// Requests returns every request so far as "METHOD /path?query", in arrival order.
func (m *MockAPI) Requests() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.requests...)
}

// UserAgents returns the User-Agent of every request so far, in arrival order, so a caller can
// assert each request it caused rather than only the last one.
func (m *MockAPI) UserAgents() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.agents...)
}

func (m *MockAPI) handle(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents = append(m.agents, r.UserAgent())
	m.requests = append(m.requests, r.Method+" "+r.URL.RequestURI())
	route, vars, ok := m.route(r)
	if !ok {
		http.Error(w, `{"errors":"no described endpoint matches `+r.Method+` `+r.URL.Path+`"}`, http.StatusNotFound)
		return
	}
	b := route.binding
	if route.ep.Envelope != nil {
		scoped := *b
		scoped.Envelope = *route.ep.Envelope
		b = &scoped
	}
	if f, ok := m.failPath[r.Method+" "+r.URL.Path]; ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		fmt.Fprint(w, f.body)
		return
	}
	if f, ok := m.fail[b.Resource]; ok && r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.status)
		fmt.Fprint(w, f.body)
		return
	}
	if route.slot == "in_place" {
		// One path can change a field of two resources (attaching an SSH key to a server).
		var body map[string]interface{}
		dec := json.NewDecoder(r.Body)
		dec.UseNumber()
		_ = dec.Decode(&body)
		answered := false
		for _, candidate := range mockRoutes() {
			if candidate.slot != "in_place" || candidate.ep.Method != r.Method {
				continue
			}
			if vars, ok := matchPath(candidate.ep.Path, r.URL.Path); ok {
				answered = m.changeInPlace(w, candidate, vars, body, answered) || answered
			}
		}
		if !answered {
			http.Error(w, `{"error":{"code":"not_found","message":"Resource with the provided ID was not found."}}`, http.StatusNotFound)
		}
		return
	}
	key := m.storeKey(route, vars)
	id := vars[route.recordVar()]
	if field := route.ep.KeyField; field != "" && route.slot != "create" && route.slot != "list" {
		// A natural key addresses the record by one of its attributes, in the path or a query param.
		value := vars[field]
		for param, name := range route.ep.QueryVars {
			if name == field {
				value = r.URL.Query().Get(param)
			}
		}
		id = m.findByAttribute(key, field, value)
		if id == "" {
			http.Error(w, `{"errors":"Resource with the provided ID was not found."}`, http.StatusNotFound)
			return
		}
		if route.slot == "update" {
			if attrs, ok := m.store[key][id]; ok {
				for _, name := range route.ep.PathFields {
					attrs[name] = vars[name]
				}
			}
		}
	}
	switch route.slot {
	case "create":
		attrs := m.decodeAttributes(b, r)
		for source, target := range b.SingletonFields {
			if values, ok := attrs[source].([]interface{}); ok && len(values) == 1 {
				attrs[target] = values[0]
				delete(attrs, source)
			}
		}
		m.nextID++
		newID := strconv.Itoa(m.nextID)
		// An object keyed by a field the request names (an SSH key's slug) keeps that key as its id.
		if path := b.Envelope.IDPath; len(path) == 1 && path[0] != "id" && len(b.Envelope.AttributesPath) == 0 {
			if chosen, ok := attrs[path[0]].(string); ok && chosen != "" {
				newID = chosen
			}
		}
		for k, v := range m.inject[b.Resource] {
			attrs[k] = v
		}
		m.applyItemInjections(b, attrs)
		if fn := m.afterWrite[b.Resource]; fn != nil {
			fn(attrs)
		}
		if m.store[key] == nil || replacingCreate(b) {
			m.store[key] = map[string]map[string]interface{}{}
			if replacingCreate(b) {
				m.order[key] = nil
			}
		}
		m.store[key][newID] = attrs
		m.order[key] = append(m.order[key], newID)
		m.keepParent(b, vars)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		m.writeItem(w, b, route.ep, newID, attrs)
	case "list":
		var records []interface{}
		for _, recID := range m.order[key] {
			attrs, ok := m.store[key][recID]
			if !ok || !m.matchesQuery(b, attrs, recID, r) {
				continue
			}
			records = append(records, m.record(b, recID, m.redacted(b, attrs)))
		}
		total := len(records)
		if b.Pagination.Via == "increment" {
			records = pageOf(records, r, b.Pagination.Param)
		}
		doc := wrapPath(b.Envelope.ListPath, records)
		if b.Pagination.Via == "increment" {
			setPath(doc, b.Pagination.Path, total)
		}
		if len(b.Pagination.Path) > 0 && b.Pagination.Via != "increment" {
			setPath(doc, b.Pagination.Path, nil)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(doc)
	case "read", "supplemental":
		attrs, recID, ok := m.lookup(key, id)
		if !ok {
			http.Error(w, `{"errors":"Resource with the provided ID was not found."}`, http.StatusNotFound)
			return
		}
		if route.slot == "supplemental" {
			selected := make(map[string]interface{}, len(route.fields))
			for _, field := range route.fields {
				selected[field] = attrs[field]
			}
			attrs = selected
		}
		m.writeItem(w, b, route.ep, recID, attrs)
	case "update":
		attrs, recID, ok := m.lookup(key, id)
		if !ok {
			http.Error(w, `{"errors":"Resource with the provided ID was not found."}`, http.StatusNotFound)
			return
		}
		for k, v := range m.decodeAttributes(b, r) {
			attrs[k] = v
		}
		m.applyItemInjections(b, attrs)
		if fn := m.afterWrite[b.Resource]; fn != nil {
			fn(attrs)
		}
		m.writeItem(w, b, route.ep, recID, attrs)
	case "delete":
		_, recID, ok := m.lookup(key, id)
		if !ok {
			http.Error(w, `{"errors":"Resource with the provided ID was not found."}`, http.StatusNotFound)
			return
		}
		delete(m.store[key], recID)
		w.WriteHeader(http.StatusNoContent)
	}
}

// pageOf is the page a list request asks for, 25 records a page unless it says how many, up to 1000.
func pageOf(records []interface{}, r *http.Request, param string) []interface{} {
	page, _ := strconv.Atoi(r.URL.Query().Get(param))
	size, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 25
	}
	if size > 1000 {
		size = 1000
	}
	from, to := (page-1)*size, page*size
	if from > len(records) {
		from = len(records)
	}
	if to > len(records) {
		to = len(records)
	}
	return records[from:to]
}

// route picks the endpoint a request is for. When several templates match, the one with the fewest
// variables wins (the first of equals): GET /api/v3/dashboards/templates lists templates, not the
// dashboard "templates", as in the apps' routers.
func (m *MockAPI) route(r *http.Request) (mockRoute, map[string]string, bool) {
	var best mockRoute
	var bestVars map[string]string
	found := false
	for _, route := range mockRoutes() {
		if route.ep.Method != r.Method {
			continue
		}
		vars, ok := matchPath(route.ep.Path, r.URL.Path)
		if !ok {
			continue
		}
		// An endpoint that takes its key as a query param answers only requests that send it: the
		// same path without it is the list.
		missing := false
		for param := range route.ep.QueryVars {
			if r.URL.Query().Get(param) == "" {
				missing = true
			}
		}
		if missing {
			continue
		}
		if !found || len(vars) < len(bestVars) {
			best, bestVars, found = route, vars, true
		}
	}
	return best, bestVars, found
}

func (route mockRoute) recordVar() string {
	ep := route.ep
	if route.slot == "supplemental" {
		ep = route.binding.Read
	}
	if ep == nil {
		return ""
	}
	return recordVar(ep)
}

// storeKey scopes records to their parent collection, so a chart's alerts and another chart's
// alerts are different collections exactly as the API says.
func (m *MockAPI) storeKey(route mockRoute, vars map[string]string) string {
	record := route.recordVar()
	var keyField string
	var pathFields []string
	if route.ep != nil {
		keyField, pathFields = route.ep.KeyField, route.ep.PathFields
	}
	parts := []string{storeName(route.binding)}
	for _, name := range sortedKeys(vars) {
		if name == record || (keyField != "" && name == keyField) || containsString(pathFields, name) {
			continue
		}
		parts = append(parts, name+"="+vars[name])
	}
	return strings.Join(parts, "|")
}

// findByAttribute is the id of the collection's record whose attribute has this value, letter case
// aside (the e-mail a team member is keyed by is stored in lowercase), or "".
func (m *MockAPI) findByAttribute(key, field, value string) string {
	for _, recID := range m.order[key] {
		if attrs, ok := m.store[key][recID]; ok && strings.EqualFold(stringifyScalar(attrs[field]), value) {
			return recID
		}
	}
	return ""
}

// SetFields changes attributes of every record of a resource whose field has this value, the way
// something outside Terraform changes it (an invitation accepted in the web app).
func (m *MockAPI) SetFields(resource, field, value string, attrs map[string]interface{}) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	changed := 0
	for key, records := range m.store {
		if key != resource && !strings.HasPrefix(key, resource+"|") {
			continue
		}
		for _, rec := range records {
			if strings.EqualFold(stringifyScalar(rec[field]), value) {
				for k, v := range attrs {
					if v == nil {
						delete(rec, k)
					} else {
						rec[k] = v
					}
				}
				changed++
			}
		}
	}
	return changed
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// lookup finds a record by id, or the collection's single record when the endpoint is a singleton
// (no id in the path at all).
func (m *MockAPI) lookup(key, id string) (map[string]interface{}, string, bool) {
	if id != "" {
		attrs, ok := m.store[key][id]
		return attrs, id, ok
	}
	for _, recID := range m.order[key] {
		if attrs, ok := m.store[key][recID]; ok {
			return attrs, recID, true
		}
	}
	return nil, "", false
}

// decodeAttributes reads the request body through the binding's declared envelope. A provider that
// wrapped a flat-write API in {"data":{"attributes":…}} yields an empty attribute set here.
func (m *MockAPI) decodeAttributes(b *apiBinding, r *http.Request) map[string]interface{} {
	var body interface{}
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		return map[string]interface{}{}
	}
	if raw, err := json.Marshal(body); err == nil {
		m.lastRaw[b.Resource] = raw
	}
	attrs, ok := digPath(body, b.Envelope.RequestPath).(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	return out
}

func (m *MockAPI) matchesQuery(b *apiBinding, attrs map[string]interface{}, id string, r *http.Request) bool {
	for k, vs := range r.URL.Query() {
		if k == "page" || k == "per_page" || len(vs) == 0 {
			continue
		}
		if b.Selector != nil && k == b.Selector.Name {
			continue
		}
		if k == b.IDKey {
			if id != vs[0] {
				return false
			}
			continue
		}
		if stringifyScalar(attrs[k]) != vs[0] {
			return false
		}
	}
	return true
}

func (m *MockAPI) redacted(b *apiBinding, attrs map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	for _, k := range m.redact[b.Resource] {
		delete(out, k)
	}
	return out
}

// record renders one object in the binding's read envelope: attributes where the envelope puts
// them, with the id (and type, when the envelope names one) alongside.
func (m *MockAPI) record(b *apiBinding, id string, attrs map[string]interface{}) interface{} {
	rec := wrapPath(b.Envelope.AttributesPath, attrs)
	if len(b.Envelope.IDPath) > 0 {
		setPath(rec, b.Envelope.IDPath, id)
	}
	if len(b.Envelope.TypePath) > 0 {
		setPath(rec, b.Envelope.TypePath, b.TypeName)
	}
	return rec
}

func (m *MockAPI) writeItem(w http.ResponseWriter, b *apiBinding, ep *endpoint, id string, attrs map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	// A keyed collection (the IP list) is the bare map the API returns: no id, no envelope.
	if b.MapList != nil {
		_ = json.NewEncoder(w).Encode(attrs)
		return
	}
	record := m.record(b, id, m.redacted(b, attrs))
	if ep.ListResponse {
		_ = json.NewEncoder(w).Encode(wrapPath(b.Envelope.ListPath, []interface{}{record}))
	} else {
		_ = json.NewEncoder(w).Encode(wrapPath(b.Envelope.ItemPath, record))
	}
}
