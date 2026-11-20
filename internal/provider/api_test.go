package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// These exercise the part of the provider the API description drives: the envelope walk, the real
// path, the pagination style and the error shape. Getting any of them wrong is a silent production
// bug, so each is pinned here rather than only in the acceptance layer.

var flatWrites = envelope{
	ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
	IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{},
	RelationshipsPath: []string{"relationships"}, IncludedPath: []string{"included"},
}

var wrappedWrites = envelope{
	ItemPath: []string{"data"}, ListPath: []string{"data"}, AttributesPath: []string{"attributes"},
	IDPath: []string{"id"}, TypePath: []string{"type"}, RequestPath: []string{"data", "attributes"},
}

var bareObject = envelope{
	ItemPath: []string{}, ListPath: nil, AttributesPath: []string{},
	IDPath: []string{"id"}, TypePath: nil, RequestPath: []string{},
}

func testBinding(url string, e envelope) *apiBinding {
	return &apiBinding{
		Resource: "widget", TypeName: "widget", IDKey: "id", BaseURL: url,
		Auth:       authScheme{Scheme: "bearer", Location: "header", Name: "Authorization"},
		Envelope:   e,
		Pagination: pagination{Via: "url", Path: []string{"pagination", "next"}},
		Errors: errorShape{MessagePath: []string{"errors"},
			DetailPaths: map[string][]string{"invalid_attributes": {"invalid_attributes"}}},
		Create: &endpoint{Method: "POST", Path: "/api/v2/widgets"},
		Read:   &endpoint{Method: "GET", Path: "/api/v2/widgets/{id}"},
		List:   &endpoint{Method: "GET", Path: "/api/v2/widgets"},
		Update: &endpoint{Method: "PATCH", Path: "/api/v2/widgets/{id}"},
		Delete: &endpoint{Method: "DELETE", Path: "/api/v2/widgets/{id}"},
	}
}

func testClient(endpoint string) *Client {
	c := NewClient("token", endpoint, map[string]string{}, "test")
	c.maxBackoff = 0
	return c
}

// The whole point of the port: whether a write is wrapped is DATA, not a constant.
func TestEncodeRequestFollowsTheEnvelope(t *testing.T) {
	attrs := map[string]interface{}{"name": "web"}
	flat, _ := json.Marshal(flatWrites.encodeRequest("widget", attrs))
	if string(flat) != `{"name":"web"}` {
		t.Fatalf("an empty request path means a flat body, got %s", flat)
	}
	wrapped, _ := json.Marshal(wrappedWrites.encodeRequest("widget", attrs))
	if string(wrapped) != `{"data":{"attributes":{"name":"web"},"type":"widget"}}` {
		t.Fatalf("a request path must wrap and stamp the type, got %s", wrapped)
	}
}

func TestDecodeItemAndListWalkTheEnvelope(t *testing.T) {
	var doc interface{}
	_ = decodeNumbers([]byte(`{"data":{"id":"7","type":"widget","attributes":{"name":"web"}}}`), &doc)
	item := flatWrites.decodeItem(doc, "id", nil)
	if item["id"] != "7" || item["name"] != "web" {
		t.Fatalf("item decode: %#v", item)
	}
	var bare interface{}
	_ = decodeNumbers([]byte(`{"id":"9","name":"solo"}`), &bare)
	if got := bareObject.decodeItem(bare, "id", nil); got["id"] != "9" || got["name"] != "solo" {
		t.Fatalf("a bare object envelope reads the body itself: %#v", got)
	}
	var list interface{}
	_ = decodeNumbers([]byte(`{"data":[{"id":"1","attributes":{"name":"a"}}]}`), &list)
	if rows := flatWrites.decodeList(list, "id", nil); len(rows) != 1 || rows[0]["name"] != "a" {
		t.Fatalf("list decode: %#v", rows)
	}
}

// A declared sideloaded field arrives as a relationship reference and is resolved against the
// document's included records, meta keys included. Undeclared relationships stay out.
func TestDeclaredRelationshipsResolveThroughIncluded(t *testing.T) {
	body := `{"data":{"id":"5","type":"on_call_calendar","attributes":{"name":"Primary"},` +
		`"relationships":{"on_call_users":{"data":[{"id":"77","type":"user","meta":{"email":"sre@example.com"}}]},` +
		`"team":{"data":{"id":"3","type":"team"}}}},` +
		`"included":[{"id":"77","type":"user","attributes":{"first_name":"Sam","phone_numbers":["+1555"]}},` +
		`{"id":"3","type":"team","attributes":{"name":"SRE"}}]}`
	var doc interface{}
	_ = decodeNumbers([]byte(body), &doc)
	item := flatWrites.decodeItem(doc, "id", []string{"on_call_users"})
	if item["name"] != "Primary" || item["id"] != "5" {
		t.Fatalf("attributes and id must still land: %#v", item)
	}
	if _, undeclared := item["team"]; undeclared {
		t.Fatalf("only declared relationships are merged: %#v", item)
	}
	users, ok := item["on_call_users"].([]interface{})
	if !ok || len(users) != 1 {
		t.Fatalf("on_call_users must resolve to a list: %#v", item["on_call_users"])
	}
	user := users[0].(map[string]interface{})
	if user["email"] != "sre@example.com" || user["first_name"] != "Sam" || user["id"] != "77" {
		t.Fatalf("ref meta plus the included record's attributes: %#v", user)
	}
	if phones, ok := user["phone_numbers"].([]interface{}); !ok || len(phones) != 1 || phones[0] != "+1555" {
		t.Fatalf("nested values survive the merge: %#v", user["phone_numbers"])
	}
	// An envelope with no relationships_path has no sideloading, whatever the body carries.
	plain := envelope{ItemPath: []string{"data"}, AttributesPath: []string{"attributes"}, IDPath: []string{"id"}}
	if got := plain.decodeItem(doc, "id", []string{"on_call_users"}); got["on_call_users"] != nil {
		t.Fatalf("an envelope that declares no relationships_path must not sideload: %#v", got)
	}
}

func TestDigPathIndexesArrays(t *testing.T) {
	var doc interface{}
	_ = decodeNumbers([]byte(`{"errors":["first","second"]}`), &doc)
	if got := digPath(doc, []string{"errors", "0"}); got != "first" {
		t.Fatalf("a numeric segment must index the array, got %#v", got)
	}
	if got := digPath(doc, []string{"errors", "9"}); got != nil {
		t.Fatalf("out of range must be nil, got %#v", got)
	}
}

// The path is transcribed from the API, including its parent variables.
func TestRequestURLExpandsTheDescribedPath(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
		w.Write([]byte(`{"data":{"id":"1","attributes":{}}}`))
	}))
	defer srv.Close()
	b := testBinding(srv.URL, flatWrites)
	b.Read = &endpoint{Method: "GET", Path: "/api/v2/status-pages/{status_page_id}/resources/{id}"}
	_, _, err := testClient(srv.URL).Get(context.Background(), b, map[string]string{"status_page_id": "5", "id": "8"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/api/v2/status-pages/5/resources/8" {
		t.Fatalf("nested path not expanded: %s", got)
	}
}

func TestCreateSendsTheEnvelopeTheAPIDeclares(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		body = string(buf)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"1","attributes":{"name":"web"}}}`))
	}))
	defer srv.Close()
	out, err := testClient(srv.URL).Create(context.Background(), testBinding(srv.URL, flatWrites),
		map[string]string{}, map[string]interface{}{"name": "web"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, `"data"`) {
		t.Fatalf("this API takes flat writes; the body must not be wrapped: %s", body)
	}
	if out["id"] != "1" {
		t.Fatalf("create response: %#v", out)
	}
}

func TestSingletonProjectionPreservesCanonicalIdentityAndRejectsFanout(t *testing.T) {
	response := `{"data":[{"id":"r7","attributes":{"application_id":"a3","environments":["prod","old"]}}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		apps, ok := body["applications"].([]interface{})
		if !ok || len(apps) != 1 || apps[0] != "a3" || body["application_id"] != nil {
			t.Errorf("expected one application on wire: %#v", body)
		}
		w.Write([]byte(response))
	}))
	defer srv.Close()
	b := testBinding(srv.URL, flatWrites)
	b.SingletonFields = map[string]string{"applications": "application_id"}
	b.Create.ListResponse = true
	out, err := testClient(srv.URL).Create(context.Background(), b, nil, map[string]interface{}{"application_id": "a3"})
	if err != nil || out["id"] != "r7" || out["application_id"] != "a3" {
		t.Fatalf("canonical response: %#v %v", out, err)
	}
	response = `{"data":[{"id":"one"},{"id":"two"}]}`
	if _, err := testClient(srv.URL).Create(context.Background(), b, nil, map[string]interface{}{"application_id": "a3"}); err == nil {
		t.Fatal("ambiguous multi-object create response accepted")
	}
}

func TestPublicMapListLookupIsLiveAndDoesNotSendCredentials(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "" {
			t.Fatal("public lookup leaked bearer token")
		}
		w.Write([]byte(`{"z":["2001:db8::1","10.0.0.10","10.0.0.2"],"a":["192.0.2.1"]}`))
	}))
	defer srv.Close()
	b := testBinding(srv.URL, bareObject)
	b.Auth = authScheme{}
	b.QueryID = "public-list"
	b.MapList = &mapListProjection{Filter: "filter", Keys: "clusters", Values: "ips", Sort: "ip"}
	c := testClient(srv.URL)
	out, err := c.Query(context.Background(), b, []string{"z"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out["ips"], []interface{}{"10.0.0.2", "10.0.0.10", "2001:db8::1"}) ||
		!reflect.DeepEqual(out["clusters"], []interface{}{"a", "z"}) || out["id"] != "public-list" {
		t.Fatalf("lookup projection lost filtering/order: %#v", out)
	}
	if _, err := c.Query(context.Background(), b, []string{"missing"}); err == nil {
		t.Fatal("unknown cluster accepted")
	}
	if calls != 2 {
		t.Fatalf("expected live requests, got %d", calls)
	}
}

func TestConfiguredClientControlsAndRateWaitHonorCancellation(t *testing.T) {
	c := testClient("")
	if err := c.configureControls(2, 3, 9, 12, 1, 1); err != nil {
		t.Fatal(err)
	}
	if c.backoff(0, "") != 3*time.Second || c.backoff(2, "") != 9*time.Second || c.http.Timeout != 12*time.Second {
		t.Fatal("configured retry/timeout controls ignored")
	}
	if err := c.waitRate(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.waitRate(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("rate wait ignored cancellation: %v", err)
	}
	if err := c.configureControls(0, 5, 4, 1, 0, 0); err == nil {
		t.Fatal("inverted retry bounds accepted")
	}
}

func TestDataSourceReadsFetchActualSupplementWithoutChangingManagedReads(t *testing.T) {
	exports := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatal("read projection attempted a write")
		}
		if strings.HasSuffix(r.URL.Path, "/export") {
			exports++
			w.Write([]byte(`{"data":{"attributes":{"data":{"name":"saved","charts":[{"id":"c1"}]}}}}`))
			return
		}
		w.Write([]byte(`{"data":{"id":"7","attributes":{"name":"saved"}}}`))
	}))
	defer srv.Close()
	b := testBinding(srv.URL, flatWrites)
	b.DataSourceReads = []supplementalRead{{Endpoint: &endpoint{Method: "GET", Path: "/objects/{id}/export"}, Fields: []string{"data"}}}
	c := testClient(srv.URL)
	if _, found, err := c.Get(context.Background(), b, map[string]string{"id": "7"}); err != nil || !found {
		t.Fatal(err)
	}
	if exports != 0 {
		t.Fatal("managed read performed datasource-only export")
	}
	out, found, err := c.GetDataSource(context.Background(), b, map[string]string{"id": "7"})
	if err != nil || !found || exports != 1 {
		t.Fatalf("supplement failed: %#v %v", out, err)
	}
	document, ok := out["data"].(map[string]interface{})
	if !ok || document["name"] != "saved" || len(document["charts"].([]interface{})) != 1 {
		t.Fatalf("export lost children: %#v", out)
	}
}

func TestListFollowsTheDeclaredPagination(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			w.Write([]byte(`{"data":[{"id":"2","attributes":{"name":"b"}}],"pagination":{"next":null}}`))
			return
		}
		w.Write([]byte(`{"data":[{"id":"1","attributes":{"name":"a"}}],"pagination":{"next":"` + srv.URL + `/api/v2/widgets?page=2"}}`))
	}))
	defer srv.Close()
	records, err := testClient(srv.URL).List(context.Background(), testBinding(srv.URL, flatWrites), map[string]string{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected both pages, got %d", len(records))
	}
}

func TestPaginationNextURLAcceptsRelativeAndNormalizedSameOriginURLs(t *testing.T) {
	b := testBinding("https://api.example.com", flatWrites)
	for _, tc := range []struct {
		name, token, want string
	}{
		{"relative path", "/api/v2/widgets?page=2", "https://api.example.com/api/v2/widgets?page=2"},
		{"relative query", "?page=2", "https://api.example.com/api/v2/widgets?page=2"},
		{"same normalized origin", "https://API.EXAMPLE.COM:443/api/v2/widgets?page=2", "https://API.EXAMPLE.COM:443/api/v2/widgets?page=2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := map[string]interface{}{"pagination": map[string]interface{}{"next": tc.token}}
			got, ok, err := nextPageURL(b, doc, "https://api.example.com/api/v2/widgets?page=1", "https://api.example.com")
			if err != nil || !ok || got != tc.want {
				t.Fatalf("nextPageURL = %q, %v, %v; want %q, true, nil", got, ok, err, tc.want)
			}
		})
	}
}

func TestListRejectsCrossOriginPaginationBeforeForwardingBearerCredentials(t *testing.T) {
	evilCalls := 0
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		evilCalls++
		if auth := r.Header.Get("Authorization"); auth != "" {
			t.Errorf("bearer credential reached cross-origin server: %q", auth)
		}
	}))
	defer evil.Close()

	base := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[],"pagination":{"next":"` + evil.URL + `/steal"}}`))
	}))
	defer base.Close()

	_, err := testClient(base.URL).List(context.Background(), testBinding(base.URL, flatWrites), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "refusing cross-origin pagination next URL") {
		t.Fatalf("expected clear cross-origin error, got %v", err)
	}
	if evilCalls != 0 {
		t.Fatalf("cross-origin next URL was requested %d times", evilCalls)
	}
}

func TestListFollowsACursorStyle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "abc" {
			w.Write([]byte(`{"data":[{"id":"2","attributes":{}}],"pagination":{"next_cursor":null}}`))
			return
		}
		w.Write([]byte(`{"data":[{"id":"1","attributes":{}}],"pagination":{"next_cursor":"abc"}}`))
	}))
	defer srv.Close()
	b := testBinding(srv.URL, flatWrites)
	b.Pagination = pagination{Via: "param", Path: []string{"pagination", "next_cursor"}, Param: "cursor"}
	records, err := testClient(srv.URL).List(context.Background(), b, map[string]string{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("a cursor style must be walked too, got %d", len(records))
	}
}

func TestErrorsAreParsedThroughTheDeclaredShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"errors":"Validation failed","invalid_attributes":{"url":"is not a URL"}}`))
	}))
	defer srv.Close()
	_, err := testClient(srv.URL).Create(context.Background(), testBinding(srv.URL, flatWrites),
		map[string]string{}, map[string]interface{}{})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected an APIError, got %T", err)
	}
	if apiErr.Message != "Validation failed" {
		t.Fatalf("message must come from message_path: %q", apiErr.Message)
	}
	if got := apiErr.Details["invalid_attributes"]["url"]; got != "is not a URL" {
		t.Fatalf("detail path not honored: %#v", apiErr.Details)
	}
}

// apiFailure answers one request with status and body, as the API did, and returns the error.
func apiFailure(t *testing.T, status int, body string) error {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	b := testBinding(srv.URL, flatWrites)
	b.Errors = errorShape{MessagePath: []string{"errors"}, DetailPaths: map[string][]string{
		"invalid_attributes": {"invalid_attributes"}, "required_attributes": {"required_attributes"}},
		RecordKeys: []string{"base"}}
	_, err := testClient(srv.URL).Create(context.Background(), b, map[string]string{}, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected the failure to surface")
	}
	return err
}

var heartbeatSchema = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"name": tftypes.String, "period": tftypes.Number, "grace": tftypes.Number, "server_timezone": tftypes.String,
}}

// diagnosticsOf renders what addAPIError shows: the attribute each diagnostic points at ("" for
// the resource) and its detail.
func diagnosticsOf(err error) map[string]string {
	var diags diag.Diagnostics
	addAPIError(&diags, "creating heartbeat", err, heartbeatSchema)
	out := map[string]string{}
	for _, d := range diags {
		at := ""
		if wp, ok := d.(diag.DiagnosticWithPath); ok {
			at = wp.Path().String()
		}
		if d.Summary() != "creating heartbeat" {
			panic("summary must name the operation: " + d.Summary())
		}
		out[at] = d.Detail()
	}
	return out
}

// Bodies the Incidents API returned on 2026-11-18 to a create on a test team.
func TestValidationMessagesLandOnTheirAttributes(t *testing.T) {
	err := apiFailure(t, 422, `{"errors":{"period":["must be greater than or equal to 30 seconds"],"grace":["is not a number"]}}`)
	got := diagnosticsOf(err)
	want := map[string]string{
		path.Root("period").String(): "period must be greater than or equal to 30 seconds",
		path.Root("grace").String():  "grace is not a number",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("every attribute message must be shown on its attribute:\n got %#v\nwant %#v", got, want)
	}
	if msg := err.Error(); !strings.Contains(msg, "period must be greater than or equal to 30 seconds") || !strings.Contains(msg, "grace is not a number") {
		t.Fatalf("the plain error must carry every sentence: %q", msg)
	}
}

func TestRecordLevelMessagesStayOnTheResource(t *testing.T) {
	got := diagnosticsOf(apiFailure(t, 422, `{"errors":{"base":["Keyword must be present when checking for keyword"],"server_timezone":["requires heartbeat period to be at least 1 hour"]}}`))
	want := map[string]string{
		path.Root("server_timezone").String(): "server_timezone requires heartbeat period to be at least 1 hour",
		"":                                    "Keyword must be present when checking for keyword",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("a key the resource has no attribute for belongs to the resource:\n got %#v\nwant %#v", got, want)
	}
}

func TestListedAttributesCarryTheTopLevelMessage(t *testing.T) {
	got := diagnosticsOf(apiFailure(t, 422, `{"errors":"Sorry, you misspelled some attributes","invalid_attributes":["bogus"]}`))
	want := map[string]string{"": "Sorry, you misspelled some attributes: bogus"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
	got = diagnosticsOf(apiFailure(t, 422, `{"errors":"Sorry, you are missing some required attributes","required_attributes":["name","period"]}`))
	want = map[string]string{
		path.Root("name").String():   "Sorry, you are missing some required attributes: name",
		path.Root("period").String(): "Sorry, you are missing some required attributes: period",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestAPlainMessageIsShownWithTheRequest(t *testing.T) {
	got := diagnosticsOf(apiFailure(t, 403, `{"errors":"This API token is read-only and cannot be used for write operations."}`))
	if len(got) != 1 || !strings.Contains(got[""], "403 This API token is read-only and cannot be used for write operations.") {
		t.Fatalf("a message naming no attribute must stay one resource-level error: %#v", got)
	}
}

func TestErrorMessageCanBeAnArrayElement(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		w.Write([]byte(`{"errors":["first","second"]}`))
	}))
	defer srv.Close()
	b := testBinding(srv.URL, flatWrites)
	b.Errors = errorShape{MessagePath: []string{"errors", "0"}, DetailPaths: map[string][]string{}}
	_, err := testClient(srv.URL).Create(context.Background(), b, map[string]string{}, map[string]interface{}{})
	if apiErr, ok := err.(*APIError); !ok || apiErr.Message != "first" {
		t.Fatalf("expected the first array element, got %v", err)
	}
}

func TestDeleteToleratesGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":"not found"}`))
	}))
	defer srv.Close()
	if err := testClient(srv.URL).Delete(context.Background(), testBinding(srv.URL, flatWrites), map[string]string{"id": "1"}); err != nil {
		t.Fatalf("a 404 on delete is the desired end state: %v", err)
	}
}

func TestRetriesHonorRetryAfter(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte(`{"data":{"id":"1","attributes":{}}}`))
	}))
	defer srv.Close()
	_, found, err := testClient(srv.URL).Get(context.Background(), testBinding(srv.URL, flatWrites), map[string]string{"id": "1"})
	if err != nil || !found || calls != 2 {
		t.Fatalf("expected one retry: calls=%d found=%v err=%v", calls, found, err)
	}
}

func TestNonIdempotentWritesRetry429ButNotGeneric5xx(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPatch} {
		for _, status := range []int{http.StatusTooManyRequests, http.StatusInternalServerError} {
			t.Run(method+" "+http.StatusText(status), func(t *testing.T) {
				calls := 0
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					if calls == 1 {
						w.Header().Set("Retry-After", "3600")
						w.WriteHeader(status)
						w.Write([]byte(`{"errors":"transient"}`))
						return
					}
					w.Write([]byte(`{}`))
				}))
				defer srv.Close()
				c := testClient(srv.URL)
				_, _, err := c.do(context.Background(), testBinding(srv.URL, flatWrites), method, srv.URL, []byte(`{}`))
				wantCalls := 1
				if status == http.StatusTooManyRequests {
					wantCalls = 2
				}
				if calls != wantCalls || (status == http.StatusTooManyRequests && err != nil) || (status >= 500 && err == nil) {
					t.Fatalf("calls=%d want=%d err=%v", calls, wantCalls, err)
				}
			})
		}
	}
}

func TestNonIdempotentWritesDoNotRetryAmbiguousNetworkFailures(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			calls := 0
			c := testClient("https://example.invalid")
			c.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("connection lost after write")
			})
			_, _, err := c.do(context.Background(), testBinding("https://example.invalid", flatWrites), method,
				"https://example.invalid/api/v2/widgets", []byte(`{}`))
			if err == nil || calls != 1 {
				t.Fatalf("ambiguous %s failure must not be replayed: calls=%d err=%v", method, calls, err)
			}
		})
	}
}

func TestRetryAfterIsBounded(t *testing.T) {
	c := testClient("https://example.invalid")
	c.maxBackoff = 3 * time.Second
	if got := c.backoff(0, "3600"); got != 3*time.Second {
		t.Fatalf("large Retry-After was not capped: %s", got)
	}
	if got := c.backoff(3, "0"); got != 0 {
		t.Fatalf("Retry-After zero was not honored: %s", got)
	}
}

func TestCredentialIsPresentedAsDescribed(t *testing.T) {
	var header, query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("Authorization")
		query = r.URL.Query().Get("team_id")
		w.Write([]byte(`{"data":{"id":"1","attributes":{}}}`))
	}))
	defer srv.Close()
	b := testBinding(srv.URL, flatWrites)
	b.Selector = &tenantSelector{In: "query", Name: "team_id"}
	c := NewClient("tok", srv.URL, map[string]string{"team_id": "t7"}, "test")
	_, _, _ = c.Get(context.Background(), b, map[string]string{"id": "1"})
	if header != "Bearer tok" {
		t.Fatalf("bearer scheme not applied: %q", header)
	}
	if query != "t7" {
		t.Fatalf("the credential's tenant selector must travel: %q", query)
	}
}

// configuredClient runs Configure the way Terraform would, with only the token and the endpoint
// set, and returns the client the provider hands its resources.
func configuredClient(t *testing.T, p provider.Provider, endpoint string) *Client {
	t.Helper()
	ctx := context.Background()
	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	objectType := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	values := map[string]tftypes.Value{}
	for name, attrType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attrType, nil)
	}
	values["api_token"] = tftypes.NewValue(tftypes.String, "token")
	values["endpoint"] = tftypes.NewValue(tftypes.String, endpoint)
	var resp provider.ConfigureResponse
	p.Configure(ctx, provider.ConfigureRequest{Config: tfsdk.Config{Raw: tftypes.NewValue(objectType, values), Schema: schemaResp.Schema}}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Configure: %v", resp.Diagnostics)
	}
	return resp.ResourceData.(*Client)
}

// The User-Agent names the caller: the provider by its own version, a bridge by its own product,
// so API-side traffic can tell them apart.
func TestUserAgentNamesTheProviderOrItsBridge(t *testing.T) {
	mock := newMockAPI(t)
	for _, tc := range []struct {
		name    string
		factory func() provider.Provider
		want    string
	}{
		{"terraform", New("1.2.3"), "terraform-provider-betterstack/1.2.3"},
		{"bridge", NewWithUserAgent("1.2.3", "pulumi-example/4.5.6"), "pulumi-example/4.5.6"},
	} {
		c := configuredClient(t, tc.factory(), mock.URL())
		_, _, _ = c.Get(context.Background(), allBindings[0], map[string]string{})
		if got := mock.LastUserAgent(); got != tc.want {
			t.Fatalf("%s: User-Agent %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestEveryBindingIsUsable(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range allBindings {
		if seen[b.Resource] {
			t.Fatalf("duplicate binding for %s", b.Resource)
		}
		seen[b.Resource] = true
		if b.BaseURL == "" || b.IDKey == "" {
			t.Fatalf("%s: incomplete binding", b.Resource)
		}
		if b.Read == nil && b.List == nil {
			t.Fatalf("%s: no way to read", b.Resource)
		}
		for _, ep := range []*endpoint{b.Create, b.Read, b.List, b.Update, b.Delete} {
			if ep == nil {
				continue
			}
			if !strings.HasPrefix(ep.Path, "/") {
				t.Fatalf("%s: %s is not a real path", b.Resource, ep.Path)
			}
		}
	}
}

func TestWrapAndSetPathRoundTrip(t *testing.T) {
	doc := wrapPath([]string{"data", "attributes"}, map[string]interface{}{"a": 1})
	setPath(doc, []string{"data", "type"}, "widget")
	want := map[string]interface{}{"data": map[string]interface{}{
		"attributes": map[string]interface{}{"a": 1}, "type": "widget"}}
	if !reflect.DeepEqual(doc, want) {
		t.Fatalf("got %#v", doc)
	}
	if got := wrapPath([]string{}, "leaf"); got != "leaf" {
		t.Fatalf("an empty path must not wrap: %#v", got)
	}
}

func TestAForcedReplacementWarnsOnTheAttributeThatForcedIt(t *testing.T) {
	objType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"data_region": tftypes.String}}
	raw := func(v string) tftypes.Value {
		return tftypes.NewValue(objType, map[string]tftypes.Value{"data_region": tftypes.NewValue(tftypes.String, v)})
	}
	modify := func(prior, planned string) planmodifier.StringResponse {
		req := planmodifier.StringRequest{
			Path:        path.Root("data_region"),
			State:       tfsdk.State{Raw: raw(prior)},
			Plan:        tfsdk.Plan{Raw: raw(planned)},
			Config:      tfsdk.Config{Raw: raw(planned)},
			StateValue:  types.StringValue(prior),
			PlanValue:   types.StringValue(planned),
			ConfigValue: types.StringValue(planned),
		}
		resp := planmodifier.StringResponse{PlanValue: req.PlanValue}
		replaceWarningString{stringplanmodifier.RequiresReplace(), "betterstack_test"}.PlanModifyString(context.Background(), req, &resp)
		return resp
	}
	changed := modify("eu-nbg-2", "us-east-9")
	if !changed.RequiresReplace || changed.Diagnostics.WarningsCount() != 1 {
		t.Fatalf("a change that forces replacement must warn once: %v", changed.Diagnostics)
	}
	w := changed.Diagnostics.Warnings()[0]
	if w.Summary() != "Changing data_region replaces the resource" {
		t.Fatalf("the warning must name the attribute: %q", w.Summary())
	}
	if wp, ok := w.(diag.DiagnosticWithPath); !ok || !wp.Path().Equal(path.Root("data_region")) {
		t.Fatalf("the warning must point at the attribute: %#v", w)
	}
	if same := modify("eu-nbg-2", "eu-nbg-2"); same.RequiresReplace || same.Diagnostics.WarningsCount() != 0 {
		t.Fatalf("an unchanged value must not warn: %v", same.Diagnostics)
	}
}
