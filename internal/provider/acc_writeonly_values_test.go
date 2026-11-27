package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// A monitor's environment_variables come back as their names with null values. Terraform keeps the
// configured values, plans nothing while the names match, and shows a name added or removed outside it.
func TestAccWriteonlyValuesKeepConfiguredValuesAndShowNameDrift(t *testing.T) {
	var mu sync.Mutex
	attrs := map[string]interface{}{}
	stored := map[string]interface{}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		write := func() bool {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return false
			}
			for key, value := range body {
				if key != "environment_variables" {
					attrs[key] = value
					continue
				}
				sent, _ := value.(map[string]interface{})
				next := map[string]interface{}{}
				for name, v := range sent {
					if v == nil {
						if before, ok := stored[name]; ok {
							next[name] = before
						}
						continue
					}
					next[name] = v
				}
				stored = next
			}
			return true
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/monitors":
			if !write() {
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v3/monitors/7":
			if !write() {
				return
			}
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/monitors/7":
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v3/monitors/7":
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "unexpected monitor route "+r.Method+" "+r.URL.Path, http.StatusNotFound)
			return
		}
		names := map[string]interface{}{}
		for name := range stored {
			names[name] = nil
		}
		out := map[string]interface{}{"environment_variables": names}
		for key, value := range attrs {
			out[key] = value
		}
		out["environment_variables"] = names
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "7", "type": "monitor", "attributes": out},
		})
	}))
	defer srv.Close()

	addr := "betterstack_monitor.test"
	config := providerConfig(srv.URL) + `resource "betterstack_monitor" "test" {
  url                   = "https://example.com"
  monitor_type          = "playwright"
  environment_variables = { API_KEY = "one", OTHER = "two" }
  deletion_protection   = false
}
`
	storedIs := func(want map[string]interface{}) resource.TestCheckFunc {
		return func(*terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if !reflect.DeepEqual(stored, want) {
				return fmt.Errorf("stored environment variables are %v, want %v", stored, want)
			}
			return nil
		}
	}
	configured := map[string]interface{}{"API_KEY": "one", "OTHER": "two"}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "environment_variables.API_KEY", "one"),
					resource.TestCheckResourceAttr(addr, "environment_variables.OTHER", "two"),
					storedIs(configured),
				),
			},
			// The read returns the names with null values: state keeps the configured ones.
			{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
			{
				PreConfig: func() { mu.Lock(); stored["ADDED"] = "outside"; mu.Unlock() },
				Config:    config, PlanOnly: true, ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() { mu.Lock(); delete(stored, "ADDED"); delete(stored, "OTHER"); mu.Unlock() },
				Config:    config, PlanOnly: true, ExpectNonEmptyPlan: true,
			},
			{Config: config, Check: storedIs(configured)},
			{Config: config, PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}

// A status page's password turns password protection on, unless the same request sends
// password_enabled too. A configuration that sets only the password must leave that to the API,
// on create and on an update that sets or changes the password, and then plan no changes.
func TestAccStatusPagePasswordAloneTurnsProtectionOn(t *testing.T) {
	var mu sync.Mutex
	attrs := map[string]interface{}{}
	var sentEnabled []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		write := func() bool {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return false
			}
			if _, ok := body["password_enabled"]; ok {
				sentEnabled = append(sentEnabled, r.Method)
			}
			// As uptime's API: a password without password_enabled turns protection on, and an
			// empty one turns it off. The password itself is never returned.
			if password, ok := body["password"]; ok {
				if _, sent := body["password_enabled"]; !sent {
					body["password_enabled"] = password != nil && password != ""
				}
				delete(body, "password")
			}
			for key, value := range body {
				attrs[key] = value
			}
			if _, ok := attrs["password_enabled"]; !ok {
				attrs["password_enabled"] = false
			}
			return true
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v5/status-pages":
			if !write() {
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v5/status-pages/9":
			if !write() {
				return
			}
		case r.Method == http.MethodGet && r.URL.Path == "/api/v5/status-pages/9":
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v5/status-pages/9":
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "unexpected status page route "+r.Method+" "+r.URL.Path, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "9", "type": "status_page", "attributes": attrs},
		})
	}))
	defer srv.Close()

	addr := "betterstack_status_page.test"
	config := func(password string) string {
		line := ""
		if password != "" {
			line = fmt.Sprintf("  password            = %q\n", password)
		}
		return providerConfig(srv.URL) + `resource "betterstack_status_page" "test" {
  company_name        = "Example"
  subdomain           = "example-password"
  timezone            = "UTC"
  deletion_protection = false
` + line + "}\n"
	}
	enabledIs := func(want bool) resource.TestCheckFunc {
		return func(*terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			if attrs["password_enabled"] != want {
				return fmt.Errorf("the API stores password_enabled %v, want %v", attrs["password_enabled"], want)
			}
			if len(sentEnabled) > 0 {
				return fmt.Errorf("password_enabled was sent on %v though the configuration leaves it out", sentEnabled)
			}
			return nil
		}
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("first"),
				Check:  resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "password_enabled", "true"), enabledIs(true)),
			},
			{Config: config("first"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{
				Config: config("second"),
				Check:  resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(addr, "password_enabled", "true"), enabledIs(true)),
			},
			{Config: config("second"), PlanOnly: true, ExpectNonEmptyPlan: false},
			{
				ResourceName: addr, ImportState: true, ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"password", "deletion_protection"},
			},
		},
	})
}

// A page created without a password stays unprotected, and setting one later turns protection on.
func TestAccStatusPagePasswordAddedLaterTurnsProtectionOn(t *testing.T) {
	var mu sync.Mutex
	attrs := map[string]interface{}{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			if password, ok := body["password"]; ok {
				if _, sent := body["password_enabled"]; !sent {
					body["password_enabled"] = password != nil && password != ""
				}
				delete(body, "password")
			}
			for key, value := range body {
				attrs[key] = value
			}
			if _, ok := attrs["password_enabled"]; !ok {
				attrs["password_enabled"] = false
			}
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v5/status-pages":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
		case r.URL.Path == "/api/v5/status-pages/9" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
			return
		case r.URL.Path == "/api/v5/status-pages/9":
		default:
			http.Error(w, "unexpected status page route "+r.Method+" "+r.URL.Path, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"id": "9", "type": "status_page", "attributes": attrs},
		})
	}))
	defer srv.Close()

	addr := "betterstack_status_page.test"
	base := providerConfig(srv.URL) + `resource "betterstack_status_page" "test" {
  company_name        = "Example"
  subdomain           = "example-password"
  timezone            = "UTC"
  deletion_protection = false
`
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: base + "}\n", Check: resource.TestCheckResourceAttr(addr, "password_enabled", "false")},
			{Config: base + "}\n", PlanOnly: true, ExpectNonEmptyPlan: false},
			{Config: base + "  password = \"later\"\n}\n", Check: resource.TestCheckResourceAttr(addr, "password_enabled", "true")},
			{Config: base + "  password = \"later\"\n}\n", PlanOnly: true, ExpectNonEmptyPlan: false},
		},
	})
}
