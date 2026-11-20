package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// The API returns a source's data region the way the source picker names it ("europe"), whichever
// accepted spelling created it. Every step here also runs the framework's empty-plan check after apply.
func dataRegionSource(name, region string, protected bool) string {
	return fmt.Sprintf("resource \"betterstack_source\" %q {\n name = %q\n platform = \"ruby\"\n data_region = %q\n deletion_protection = %t\n}\n", name, name, region, protected)
}

// Written with an older region id or a cluster name, the source keeps planning clean: the provider keeps
// what was written. A source moved from logtail_source is the same case: its state holds whatever the old
// provider stored, often a cluster name such as eu-nbg-2, and the migration writes that into the config.
func TestAccDataRegionCreatedWithAnOlderIdPlansClean(t *testing.T) {
	for _, written := range []string{"germany", "eu-nbg-2"} {
		t.Run(written, func(t *testing.T) {
			mock := newMockAPI(t)
			mock.InjectOnCreate("source", map[string]interface{}{"data_region": "europe"})
			config := providerConfig(mock.URL()) + dataRegionSource("logs", written, false)
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testProviderFactories,
				Steps: []resource.TestStep{
					{Config: config, Check: resource.TestCheckResourceAttr("betterstack_source.logs", "data_region", written)},
					{Config: config, PlanOnly: true},
				},
			})
		})
	}
}

// A source made elsewhere (the web UI) and imported. Written with the picker's name, the import plans no
// change; written with an older id of the same region, one in-place update that sends nothing for the
// region, and a clean plan after it. A different region still replaces the source, once deletion
// protection, on for an imported source, is turned off.
func TestAccDataRegionImportPlansNoReplacement(t *testing.T) {
	for _, tc := range []struct {
		written string
		action  plancheck.ResourceActionType
	}{
		{"europe", plancheck.ResourceActionNoop},
		{"germany", plancheck.ResourceActionUpdate},
	} {
		t.Run(tc.written, func(t *testing.T) {
			mock := newMockAPI(t)
			mock.InjectOnCreate("source", map[string]interface{}{"data_region": "europe"})
			addr := "betterstack_source.logs"
			imported := providerConfig(mock.URL()) + "import {\n to = " + addr + "\n id = \"1\"\n}\n" + dataRegionSource("logs", tc.written, true)
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testProviderFactories,
				Steps: []resource.TestStep{
					// The mock numbers records from 1, so the source made here is "1".
					{
						Config: providerConfig(mock.URL()) + dataRegionSource("logs", "europe", false),
						Check: func(s *terraform.State) error {
							if id := s.RootModule().Resources[addr].Primary.ID; id != "1" {
								return fmt.Errorf("expected the first mock record, got id %q", id)
							}
							return nil
						},
					},
					// Forget it without deleting it, as if Terraform had never created it.
					{Config: providerConfig(mock.URL()) + "removed {\n from = " + addr + "\n lifecycle {\n destroy = false\n }\n}\n"},
					{
						Config: imported,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, tc.action)},
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(addr, "data_region", tc.written),
							func(*terraform.State) error {
								if tc.action == plancheck.ResourceActionUpdate && strings.Contains(string(mock.LastRequestBody("source")), "data_region") {
									return fmt.Errorf("the in-place update sent data_region: %s", mock.LastRequestBody("source"))
								}
								return nil
							},
						),
					},
					{Config: imported, PlanOnly: true},
					{Config: providerConfig(mock.URL()) + dataRegionSource("logs", tc.written, false)},
					{
						Config: providerConfig(mock.URL()) + dataRegionSource("logs", "us", false),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionReplace)},
						},
					},
				},
			})
		})
	}
}
