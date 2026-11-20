package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccOmittedComputedCreateOnlyFieldsDoNotReplace(t *testing.T) {
	for _, name := range []string{"source", "collector"} {
		t.Run(name, func(t *testing.T) {
			mock := newMockAPI(t)
			addr := "betterstack_" + name + ".test"
			config := func(label string) string {
				return providerConfig(mock.URL()) + fmt.Sprintf("resource %q \"test\" {\n name = %q\n platform = \"docker\"\n deletion_protection = false\n}\n", "betterstack_"+name, label)
			}
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testProviderFactories,
				Steps: []resource.TestStep{
					{Config: config("before")},
					{
						Config: config("after"),
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate),
						}},
					},
					{Config: config("after"), PlanOnly: true, ExpectNonEmptyPlan: false},
				},
			})
		})
	}
}

func TestAccExplorationGroupCanBeClearedByRemovingTheAttribute(t *testing.T) {
	var mu sync.Mutex
	attrs := map[string]interface{}{}
	sawNull := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Query().Get("better_stack_team_id") != "t123" {
			http.Error(w, "expected literal better_stack_team_id selector", http.StatusBadRequest)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/explorations":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid create body", http.StatusBadRequest)
				return
			}
			if body["data"] != nil || body["exploration_group_id"] == nil {
				http.Error(w, "expected a flat create body with exploration_group_id", http.StatusBadRequest)
				return
			}
			attrs = body
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v2/explorations/41":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid update body", http.StatusBadRequest)
				return
			}
			update := body
			value, present := update["exploration_group_id"]
			if !present || value != nil {
				http.Error(w, "expected explicit JSON null after removing exploration_group_id", http.StatusBadRequest)
				return
			}
			sawNull = true
			for key, value := range update {
				attrs[key] = value
			}
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/explorations/41":
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v2/explorations/41":
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "unexpected literal exploration route", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id": "41", "type": "exploration", "attributes": attrs,
			},
		})
	}))
	defer srv.Close()

	provider := fmt.Sprintf(`provider "betterstack" {
  api_token = "test-token"
  endpoint  = %q
  better_stack_team_id = "t123"
}

`, srv.URL)
	create := provider + `resource "betterstack_exploration" "nullable" {
  name                 = "nullable-contract"
  exploration_group_id = 17
  chart                = { chart_type = "line_chart" }
  queries              = [{ query_type = "sql_expression" }]
}
`
	clear := provider + `resource "betterstack_exploration" "nullable" {
  name    = "nullable-contract-updated"
  chart   = { chart_type = "line_chart" }
  queries = [{ query_type = "sql_expression" }]
}
`
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: create,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("betterstack_exploration.nullable", "id", "41"),
					resource.TestCheckResourceAttr("betterstack_exploration.nullable", "exploration_group_id", "17"),
				),
			},
			{
				Config: clear,
				Check:  resource.TestCheckNoResourceAttr("betterstack_exploration.nullable", "exploration_group_id"),
			},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if !sawNull {
		t.Fatal("removing exploration_group_id did not send JSON null")
	}
}

func TestAccAdoptionAndRegistrationRetainRemoteObjects(t *testing.T) {
	var mu sync.Mutex
	jiraName := "existing"
	patches, registrations := 0, 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v2/jira-integrations/77":
			if r.Method == http.MethodPatch {
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				name, ok := body["name"].(string)
				if !ok {
					t.Error("adoption did not apply configured attributes")
					return
				}
				jiraName, patches = name, patches+1
			} else if r.Method != http.MethodGet {
				t.Errorf("unexpected Jira mutation: %s", r.Method)
				w.WriteHeader(500)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{
				"id": "77", "type": "jira_integration", "attributes": map[string]interface{}{"name": jiraName},
			}})
		case "/api/v1/releases":
			if r.Method != http.MethodPost {
				t.Errorf("unexpected release request: %s", r.Method)
				w.WriteHeader(500)
				return
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			apps, ok := body["applications"].([]interface{})
			if !ok || len(apps) != 1 || fmt.Sprint(apps[0]) != "7" {
				t.Errorf("release registration did not target exactly one configured application: %#v", body)
			}
			registrations++
			// A bare object: one entry per registered release, with only its id and application_id.
			w.Write([]byte(`{"version":"v1","applications":[7],"environment":["production"],"releases":[{"id":5,"application_id":7}]}`))
		case "/api/v1/releases/5":
			if r.Method != http.MethodGet {
				t.Errorf("release history mutated by %s", r.Method)
				w.WriteHeader(500)
				return
			}
			w.Write([]byte(`{"data":{"id":"5","type":"release","attributes":{"version":"v1","application_id":7,"environments":["production","historic"]}}}`))
		default:
			t.Errorf("unexpected lifecycle route %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	config := fmt.Sprintf(`provider "betterstack" {
  api_token = "test-token"
  endpoint = %q
  api_rate_limit = 0
}
resource "betterstack_jira_integration" "adopted" {
  id = "77"
  name = "managed"
}
resource "betterstack_release" "registered" {
  application_id = 7
  version = "v1"
  environment = ["production"]
}
`, srv.URL)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("betterstack_jira_integration.adopted", "id", "77"),
				resource.TestCheckResourceAttr("betterstack_jira_integration.adopted", "name", "managed"),
				resource.TestCheckResourceAttr("betterstack_release.registered", "id", "5"),
				resource.TestCheckResourceAttr("betterstack_release.registered", "application_id", "7"),
				resource.TestCheckResourceAttr("betterstack_release.registered", "environment.#", "1"),
			)},
			{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if patches != 1 || registrations != 1 {
		t.Fatalf("unexpected lifecycle mutations: PATCH=%d registration=%d", patches, registrations)
	}
}
