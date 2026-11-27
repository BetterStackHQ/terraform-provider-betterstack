package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// A write the API accepts at once and finishes later: the provider polls the object's read until
// the declared success holds, stops at a declared failure, and a delete ends when the object is gone.
// These tests use a made-up API, since the waiting knows no product.

type settlingAPI struct {
	mu       sync.Mutex
	statuses []string // what each GET answers in turn; the last one repeats
	gets     int
	gone     bool
}

func (a *settlingAPI) handler(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case r.Method == "POST":
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "w1", "status": "starting"})
	case r.Method == "DELETE":
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "w1", "status": "deleting"})
	case a.gone && a.gets >= 1:
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"not_found","message":"Resource not found."}}`))
	default:
		status := a.statuses[min(a.gets, len(a.statuses)-1)]
		a.gets++
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "w1", "status": status, "status_message": "the host is full"})
	}
}

func settlingBinding(url string, waits map[string]string) *apiBinding {
	b := testBinding(url, bareObject)
	b.Create = &endpoint{Method: "POST", Path: "/v1/widgets"}
	b.Read = &endpoint{Method: "GET", Path: "/v1/widgets/{id}"}
	b.Delete = &endpoint{Method: "DELETE", Path: "/v1/widgets/{id}"}
	b.Waits = waits
	return b
}

const runningWait = `{"bindings":{"id":{"source":"result","path":["id"]}},"success":{"path":["status"],"expected":{"source":"literal","value":"running"}},"failure":[{"path":["status"],"in":["failed"]}],"messages":["status_message"]}`

func quickWaits(t *testing.T) {
	delays, limit := waitDelays, waitLimit
	waitDelays, waitLimit = []time.Duration{time.Millisecond}, time.Minute
	t.Cleanup(func() { waitDelays, waitLimit = delays, limit })
}

func TestCreateWaitsUntilTheDeclaredSuccess(t *testing.T) {
	quickWaits(t)
	api := &settlingAPI{statuses: []string{"starting", "starting", "running"}}
	srv := httptest.NewServer(http.HandlerFunc(api.handler))
	defer srv.Close()
	c := NewClient("tok", "", nil, "test")
	out, err := c.Create(context.Background(), settlingBinding(srv.URL, map[string]string{"create": runningWait}), nil, map[string]interface{}{"name": "w"})
	if err != nil {
		t.Fatal(err)
	}
	if out["status"] != "running" || api.gets != 3 {
		t.Fatalf("want the settled object after 3 reads, got %v after %d", out, api.gets)
	}
}

func TestCreateStopsAtADeclaredFailureWithTheObject(t *testing.T) {
	quickWaits(t)
	api := &settlingAPI{statuses: []string{"starting", "failed"}}
	srv := httptest.NewServer(http.HandlerFunc(api.handler))
	defer srv.Close()
	c := NewClient("tok", "", nil, "test")
	_, err := c.Create(context.Background(), settlingBinding(srv.URL, map[string]string{"create": runningWait}), nil, map[string]interface{}{})
	unsettled, ok := err.(*waitError)
	if !ok {
		t.Fatalf("want a waitError, got %v", err)
	}
	if unsettled.Object["id"] != "w1" || !strings.Contains(err.Error(), "failed") || !strings.Contains(err.Error(), "the host is full") {
		t.Fatalf("the error must carry the object, its state and its message: %v %v", unsettled.Object, err)
	}
}

func TestAWriteWithoutADeclaredWaitReturnsAtOnce(t *testing.T) {
	api := &settlingAPI{statuses: []string{"starting"}}
	srv := httptest.NewServer(http.HandlerFunc(api.handler))
	defer srv.Close()
	c := NewClient("tok", "", nil, "test")
	out, err := c.Create(context.Background(), settlingBinding(srv.URL, nil), nil, map[string]interface{}{})
	if err != nil || out["status"] != "starting" || api.gets != 0 {
		t.Fatalf("no wait declared: want the create's answer and no reads, got %v %v after %d", out, err, api.gets)
	}
}

func TestAWaitWhoseConditionDoesNotApplyIsSkipped(t *testing.T) {
	api := &settlingAPI{statuses: []string{"starting"}}
	srv := httptest.NewServer(http.HandlerFunc(api.handler))
	defer srv.Close()
	when := `{"bindings":{"id":{"source":"result","path":["id"]}},"when":{"path":["server_id"],"in":["s1"]},"success":{"path":["status"],"expected":{"source":"literal","value":"running"}}}`
	c := NewClient("tok", "", nil, "test")
	if _, err := c.Create(context.Background(), settlingBinding(srv.URL, map[string]string{"create": when}), nil, map[string]interface{}{"server_id": "s2"}); err != nil || api.gets != 0 {
		t.Fatalf("a request the wait's when refuses is not watched: %v after %d reads", err, api.gets)
	}
}

func TestASuccessCanCompareWithTheRequest(t *testing.T) {
	scope := waitScope{request: map[string]interface{}{"server_id": "s1"}}
	p := waitPredicate{All: []waitPredicate{
		{Path: []string{"status"}, Expected: &waitBinding{Source: "literal", Value: "attached"}},
		{Path: []string{"server_id"}, Expected: &waitBinding{Source: "request", Path: []string{"server_id"}}},
	}}
	if !p.holds(map[string]interface{}{"status": "attached", "server_id": "s1"}, scope) {
		t.Fatal("the attachment to the requested server must count")
	}
	if p.holds(map[string]interface{}{"status": "attached", "server_id": "s9"}, scope) {
		t.Fatal("an attachment elsewhere must not count")
	}
	if p.holds(map[string]interface{}{"status": "attached"}, scope) {
		t.Fatal("a missing value is not yet")
	}
	n := waitPredicate{Path: []string{"count"}, AtLeast: &waitBinding{Source: "literal", Value: json.Number("2")}}
	if !n.holds(map[string]interface{}{"count": json.Number("3")}, scope) || n.holds(map[string]interface{}{"count": 1.0}, scope) {
		t.Fatal("at_least compares numbers")
	}
}

func TestDeleteWaitsUntilTheObjectIsGone(t *testing.T) {
	quickWaits(t)
	api := &settlingAPI{statuses: []string{"deleting"}, gone: true}
	srv := httptest.NewServer(http.HandlerFunc(api.handler))
	defer srv.Close()
	wait := `{"bindings":{"id":{"source":"result","path":["id"]}},"success":{"path":["status"],"expected":{"source":"literal","value":"destroyed"}},"disappearance":true}`
	c := NewClient("tok", "", nil, "test")
	if err := c.Delete(context.Background(), settlingBinding(srv.URL, map[string]string{"delete": wait}), map[string]string{"id": "w1"}); err != nil {
		t.Fatal(err)
	}
	if api.gets != 1 {
		t.Fatalf("want one read that found the object gone, got %d", api.gets)
	}
}

func TestATemplatedBaseURLTakesItsProviderSetting(t *testing.T) {
	b := testBinding("https://{shard}.example.com", bareObject)
	c := NewClient("tok", "", nil, "test")
	err := c.checkBaseURL(b)
	if err == nil || !strings.Contains(err.Error(), "set shard in the provider block or the BETTERSTACK_SHARD environment variable") {
		t.Fatalf("a missing setting must be refused before any request, naming it: %v", err)
	}
	c.baseURLVars = map[string]string{"shard": "us-east-1a"}
	if err := c.checkBaseURL(b); err != nil {
		t.Fatal(err)
	}
	if got := c.baseURL(b); got != "https://us-east-1a.example.com" {
		t.Fatalf("want the zone's host, got %s", got)
	}
	c.baseURLVars = map[string]string{"shard": "evil.com/x?"}
	if err := c.checkBaseURL(b); err == nil {
		t.Fatal("a value that isn't one host label must be refused")
	}
	if _, _, err := c.Get(context.Background(), b, map[string]string{"id": "1"}); err == nil {
		t.Fatal("the request must not go out")
	}
	c.endpoint = "http://127.0.0.1:1"
	if err := c.checkBaseURL(b); err != nil {
		t.Fatalf("an endpoint override replaces the whole base URL: %v", err)
	}
}
