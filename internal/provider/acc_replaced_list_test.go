package provider

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// A monitor's request_headers is a list the API replaces whole when it's sent, numbering every
// header anew. An edit made outside Terraform to another attribute (uptime copies one monitor's
// domain_expiration to every monitor on the same domain) must plan only that attribute: the headers
// keep their ids, and the update leaves them out. A header change is still sent.

const replacedListAddr = "betterstack_monitor.test"

func replacedListConfig(url, value string) string {
	return providerConfig(url) + fmt.Sprintf("resource \"betterstack_monitor\" \"test\" {\n url = \"https://example.com/\"\n monitor_type = \"status\"\n deletion_protection = false\n request_headers = [{ name = \"X-E2E\", value = %q }]\n}\n", value)
}

// listUnchanged fails a plan that changes or recomputes the attribute, the item ids included.
type listUnchanged struct{ attribute string }

func (c listUnchanged) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	for _, change := range req.Plan.ResourceChanges {
		if change.Address != replacedListAddr {
			continue
		}
		before, _ := change.Change.Before.(map[string]interface{})
		after, _ := change.Change.After.(map[string]interface{})
		unknown, _ := change.Change.AfterUnknown.(map[string]interface{})
		if !reflect.DeepEqual(before[c.attribute], after[c.attribute]) {
			resp.Error = fmt.Errorf("%s planned to change: %v -> %v", c.attribute, before[c.attribute], after[c.attribute])
		} else if marks, ok := unknown[c.attribute]; ok && strings.Contains(fmt.Sprint(marks), "true") {
			resp.Error = fmt.Errorf("%s planned as unknown: %v", c.attribute, marks)
		}
		return
	}
	resp.Error = fmt.Errorf("no planned change for %s", replacedListAddr)
}

func TestAccMonitorRequestHeadersKeepTheirIDsWhenAnotherAttributeChanges(t *testing.T) {
	mock := newMockAPI(t)
	mock.AssignItemIDs("monitor", "request_headers")
	headerID := ""
	readID := func(s *terraform.State) string {
		return s.RootModule().Resources[replacedListAddr].Primary.Attributes["request_headers.0.id"]
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: replacedListConfig(mock.URL(), "1"),
				Check: func(s *terraform.State) error {
					if headerID = readID(s); headerID == "" {
						return fmt.Errorf("the header has no id")
					}
					return nil
				},
			},
			{
				PreConfig: func() {
					if mock.SetFields("monitor", "url", "https://example.com/", map[string]interface{}{"domain_expiration": 30}) != 1 {
						t.Fatal("no monitor to change")
					}
				},
				Config: replacedListConfig(mock.URL(), "1"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(replacedListAddr, plancheck.ResourceActionUpdate),
						listUnchanged{"request_headers"},
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(replacedListAddr, "domain_expiration"),
					func(s *terraform.State) error {
						body := string(mock.LastRequestBody("monitor"))
						if strings.Contains(body, "request_headers") {
							return fmt.Errorf("the update sent the unchanged headers: %s", body)
						}
						if !strings.Contains(body, `"domain_expiration":null`) {
							return fmt.Errorf("the update didn't clear domain_expiration: %s", body)
						}
						if got := readID(s); got != headerID {
							return fmt.Errorf("the header's id changed from %s to %s", headerID, got)
						}
						return nil
					},
				),
			},
			{
				Config: replacedListConfig(mock.URL(), "2"),
				Check: func(s *terraform.State) error {
					if body := string(mock.LastRequestBody("monitor")); !strings.Contains(body, "request_headers") {
						return fmt.Errorf("a header change wasn't sent: %s", body)
					}
					if got := readID(s); got == headerID || got == "" {
						return fmt.Errorf("the replaced header kept id %q", got)
					}
					return nil
				},
			},
		},
	})
}
