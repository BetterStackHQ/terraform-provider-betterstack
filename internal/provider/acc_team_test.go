package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// One provider manages objects in several teams: a resource without team_id is in the provider's
// team, one with team_id is created and read in that team, and changing it replaces the object.
func TestAccAResourceTeamIDPicksItsTeam(t *testing.T) {
	var mu sync.Mutex
	type group struct{ name, team string }
	groups := map[string]*group{}
	var creates []string
	next := 40
	reply := func(w http.ResponseWriter, id string, g *group) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"id": id, "type": "monitor_group", "attributes": map[string]interface{}{
			"name": g.name, "team_id": g.team, "team_name": "Team " + g.team, "paused": false, "sort_index": nil, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}}})
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		team, selected := r.URL.Query().Get("team_id"), r.URL.Query().Has("team_id")
		id := strings.TrimPrefix(r.URL.Path, "/api/v3/monitor-groups/")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/monitor-groups":
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			next++
			id := fmt.Sprint(next)
			groups[id] = &group{name: fmt.Sprint(body["name"]), team: team}
			creates = append(creates, body["name"].(string)+"@"+team)
			reply(w, id, groups[id])
		case groups[id] == nil || selected && groups[id].team != team:
			// The API finds an object only in its own team, or by ID with no team selected.
			http.Error(w, `{"errors":"not found"}`, http.StatusNotFound)
		case r.Method == http.MethodGet:
			reply(w, id, groups[id])
		case r.Method == http.MethodPatch:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			groups[id].name = fmt.Sprint(body["name"])
			reply(w, id, groups[id])
		case r.Method == http.MethodDelete:
			delete(groups, id)
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()
	config := func(team string) string {
		return fmt.Sprintf("provider \"betterstack\" {\n  api_token = \"test-token\"\n  endpoint  = %q\n  team_id   = \"t123\"\n}\n", srv.URL) +
			"resource \"betterstack_monitor_group\" \"home\" {\n  name = \"home\"\n}\n" +
			fmt.Sprintf("resource \"betterstack_monitor_group\" \"other\" {\n  name    = \"other\"\n  team_id = %q\n}\n", team)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("t456"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betterstack_monitor_group.home", "team_id", "t123"),
					resource.TestCheckResourceAttr("betterstack_monitor_group.other", "team_id", "t456"),
				),
			},
			{Config: config("t456"), PlanOnly: true},
			{
				Config: config("t789"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("betterstack_monitor_group.other", plancheck.ResourceActionDestroyBeforeCreate),
					plancheck.ExpectResourceAction("betterstack_monitor_group.home", plancheck.ResourceActionNoop),
				}},
				Check: resource.TestCheckResourceAttr("betterstack_monitor_group.other", "team_id", "t789"),
			},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	sort.Strings(creates)
	if want := "home@t123 other@t456 other@t789"; strings.Join(creates, " ") != want {
		t.Fatalf("created in: %q, want %q", creates, want)
	}
}

// A moved or imported object's state has no team yet: its first read finds it by ID in another team
// than the provider's, and records that team.
func TestAccAnImportedObjectInAnotherTeamIsFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("team_id") && r.URL.Query().Get("team_id") != "t456" {
			http.Error(w, `{"errors":"not found"}`, http.StatusNotFound)
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":"7","type":"monitor_group","attributes":{"name":"other","team_id":"t456","team_name":"Staging","paused":false,"sort_index":null,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}}`))
	}))
	defer srv.Close()
	config := fmt.Sprintf("provider \"betterstack\" {\n  api_token = \"test-token\"\n  endpoint  = %q\n  team_id   = \"t123\"\n}\n", srv.URL) +
		"import {\n  to = betterstack_monitor_group.other\n  id = \"7\"\n}\n" +
		"resource \"betterstack_monitor_group\" \"other\" {\n  name    = \"other\"\n  team_id = \"t456\"\n}\n"
	// Without team_id, a replacement would create it in the provider's team: the plan says to name it.
	unnamed := strings.Replace(config, "  team_id = \"t456\"\n", "", 1)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: unnamed, ExpectError: regexp.MustCompile(`(?s)It is in team t456, and the provider's team_id is t123.*Set team_id = "t456"`)},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("betterstack_monitor_group.other", plancheck.ResourceActionNoop),
				}},
				Check: resource.TestCheckResourceAttr("betterstack_monitor_group.other", "team_id", "t456"),
			},
		},
	})
}

// A team token already picks its team and the API refuses any team selector from it. With no
// team_id on the provider, the team each read records in state is never sent back: create,
// update and destroy select no team.
func TestAccATeamTokenSendsNoTeamSelector(t *testing.T) {
	var mu sync.Mutex
	name, exists := "", false
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, r.Method)
		if r.URL.Query().Has("team_id") {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"errors":"Sorry, you misspelled some attributes","invalid_attributes":[{"team_id":"This token already works in team Ops. Leave out team_id."}]}`))
			return
		}
		switch r.Method {
		case http.MethodPost, http.MethodPatch:
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			name, exists = fmt.Sprint(body["name"]), true
		case http.MethodDelete:
			exists = false
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if !exists {
			http.Error(w, `{"errors":"not found"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"id": "7", "type": "monitor_group", "attributes": map[string]interface{}{
			"name": name, "team_id": "t4", "team_name": "Ops", "paused": false, "sort_index": nil, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}}})
	}))
	defer srv.Close()
	config := func(name string) string {
		return fmt.Sprintf("provider \"betterstack\" {\n  api_token = \"team-token\"\n  endpoint  = %q\n}\n", srv.URL) +
			fmt.Sprintf("resource \"betterstack_monitor_group\" \"ops\" {\n  name = %q\n}\n", name)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("ops"), Check: resource.TestCheckResourceAttr("betterstack_monitor_group.ops", "team_id", "t4")},
			{Config: config("ops renamed"), Check: resource.TestCheckResourceAttr("betterstack_monitor_group.ops", "name", "ops renamed")},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if got := strings.Join(calls, " "); !strings.Contains(got, "PATCH") || !strings.Contains(got, "DELETE") {
		t.Fatalf("requests: %s, want an update and a delete", got)
	}
}

// A failed read with no team (an outage) is reported, never taken for "not found": the object stays
// in state instead of being planned again.
func TestAccAFailedNoTeamReadIsReported(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("team_id") {
			http.Error(w, `{"errors":"not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"errors":"upstream unavailable"}`, http.StatusInternalServerError)
	}))
	defer srv.Close()
	config := fmt.Sprintf("provider \"betterstack\" {\n  api_token = \"test-token\"\n  endpoint  = %q\n  team_id   = \"t123\"\n  api_retry_max = 0\n}\n", srv.URL) +
		"import {\n  to = betterstack_monitor_group.other\n  id = \"7\"\n}\n" +
		"resource \"betterstack_monitor_group\" \"other\" {\n  name    = \"other\"\n  team_id = \"t456\"\n}\n"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, ExpectError: regexp.MustCompile(`(?s)500.*upstream unavailable`)},
		},
	})
}
