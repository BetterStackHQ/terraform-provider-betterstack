package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// The API stores a zone it is sent as its Rails name ("Tokyo" for "Asia/Tokyo") and returns that
// name. A configuration may spell the zone either way: after a create or an import the plan is
// empty, and stays empty.
func TestAccATimeZoneUnderEitherNamePlansNoChange(t *testing.T) {
	cases := []struct{ tfType, record, field, body string }{
		{"betterstack_monitor", "monitor", "maintenance_timezone", "url = \"https://example.com\"\n"},
		{"betterstack_heartbeat", "heartbeat", "maintenance_timezone", "name = \"example\"\n period = 60\n grace = 60\n"},
		{"betterstack_status_page", "status_page", "timezone", "company_name = \"example\"\n subdomain = \"example-page\"\n"},
	}
	for _, tc := range cases {
		for _, zone := range []string{"Asia/Tokyo", "Tokyo"} {
			t.Run(tc.tfType+"/"+zone, func(t *testing.T) {
				mock := newMockAPI(t)
				field := tc.field
				mock.AfterWrite(tc.record, func(attrs map[string]interface{}) {
					if attrs[field] == "Asia/Tokyo" {
						attrs[field] = "Tokyo"
					}
				})
				addr := tc.tfType + ".test"
				res := fmt.Sprintf("resource %q \"test\" {\n %s %s = %q\n}\n", tc.tfType, tc.body, field, zone)
				created := providerConfig(mock.URL()) + res
				forget := providerConfig(mock.URL()) + "removed {\n from = " + addr + "\n lifecycle {\n destroy = false\n }\n}\n"
				imported := providerConfig(mock.URL()) + "import {\n to = " + addr + "\n id = \"1\"\n}\n" + res
				resource.Test(t, resource.TestCase{
					ProtoV6ProviderFactories: testProviderFactories,
					Steps: []resource.TestStep{
						{Config: created},
						{Config: created, PlanOnly: true},
						{Config: forget},
						{
							Config: imported,
							ConfigPlanChecks: resource.ConfigPlanChecks{
								PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionNoop)},
							},
						},
						{Config: imported, PlanOnly: true},
						// A real change still plans as one, and leaves nothing to plan after it.
						{
							Config: providerConfig(mock.URL()) + strings.Replace(res, zone, "Europe/Prague", 1),
							ConfigPlanChecks: resource.ConfigPlanChecks{
								PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
							},
						},
						{Config: providerConfig(mock.URL()) + strings.Replace(res, zone, "Europe/Prague", 1), PlanOnly: true},
						{Config: forget},
					},
				})
			})
		}
	}
}
