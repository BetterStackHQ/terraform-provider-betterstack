package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// deletion_protection lives in Terraform state only. While it says true, the plan refuses to destroy
// or replace the object; once an apply has turned it off, the delete goes out with force=true, so
// the API's guard on a source that holds data lets it through. No request ever carries the field.
func TestAccDeletionProtectionRefusesDestroyAndReplace(t *testing.T) {
	mock := newMockAPI(t)
	config := func(protected bool, region string) string {
		return providerConfig(mock.URL()) + fmt.Sprintf(
			"resource \"betterstack_source\" \"test\" {\n name = \"prod-logs\"\n platform = \"ruby\"\n data_region = %q\n deletion_protection = %t\n}\n",
			region, protected)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: config(true, "us_east")},
			{
				Config:      providerConfig(mock.URL()),
				ExpectError: regexp.MustCompile(`(?s)Deletion protection is on for the source "prod-logs".*This plan destroys`),
			},
			{
				Config:      config(true, "germany"),
				ExpectError: regexp.MustCompile(`(?s)Deletion protection is on for "prod-logs".*Changing data_region replaces it`),
			},
			// Turning it off is its own apply; the replacement and the destroy then go through.
			{Config: config(false, "us_east")},
			{Config: config(false, "germany")},
		},
		CheckDestroy: func(*terraform.State) error {
			deletes := 0
			for _, request := range mock.Requests() {
				if !strings.HasPrefix(request, "DELETE ") {
					continue
				}
				deletes++
				if !strings.Contains(request, "force=true") {
					return fmt.Errorf("delete without force=true: %s", request)
				}
			}
			if deletes != 2 {
				return fmt.Errorf("expected the replacement's and the final destroy's deletes, got %d", deletes)
			}
			if body := string(mock.LastRequestBody("source")); strings.Contains(body, "deletion_protection") {
				return fmt.Errorf("a request carried deletion_protection: %s", body)
			}
			return nil
		},
	})
}

// An imported object starts protected, as a created one does: a config that leaves the attribute out
// plans no change after the import, and a destroy is refused.
func TestAccDeletionProtectionIsOnAfterImport(t *testing.T) {
	mock := newMockAPI(t)
	addr := "betterstack_source.test"
	source := "resource \"betterstack_source\" \"test\" {\n name = \"prod-logs\"\n platform = \"ruby\"\n}\n"
	imported := providerConfig(mock.URL()) + "import {\n to = " + addr + "\n id = \"1\"\n}\n" + source
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			// The mock numbers records from 1, so the source made here is "1".
			{Config: providerConfig(mock.URL()) + source},
			// Forget it without deleting it, as if Terraform had never created it.
			{Config: providerConfig(mock.URL()) + "removed {\n from = " + addr + "\n lifecycle {\n destroy = false\n }\n}\n"},
			{
				Config: imported,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionNoop)},
				},
				Check: resource.TestCheckResourceAttr(addr, "deletion_protection", "true"),
			},
			{Config: imported, PlanOnly: true},
			{
				Config:      providerConfig(mock.URL()),
				ExpectError: regexp.MustCompile(`Deletion protection is on for the source "prod-logs"`),
			},
			{Config: providerConfig(mock.URL()) + "removed {\n from = " + addr + "\n lifecycle {\n destroy = false\n }\n}\n"},
		},
	})
}

// A replacement the plan does not show as one (a tainted object, terraform apply -replace) and a
// pulumi destroy never reach the plan checks, so the delete refuses on its own and sends nothing.
func TestAccDeletionProtectionRefusesATaintedReplace(t *testing.T) {
	mock := newMockAPI(t)
	addr := "betterstack_source.test"
	source := providerConfig(mock.URL()) + "resource \"betterstack_source\" \"test\" {\n name = \"prod-logs\"\n platform = \"ruby\"\n}\n"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{Config: source},
			{
				Config:      source,
				Taint:       []string{addr},
				ExpectError: regexp.MustCompile(`(?s)Deletion protection is on for the source "prod-logs".*so it was not deleted`),
			},
			{Config: providerConfig(mock.URL()) + "removed {\n from = " + addr + "\n lifecycle {\n destroy = false\n }\n}\n"},
		},
		CheckDestroy: func(*terraform.State) error {
			for _, request := range mock.Requests() {
				if strings.HasPrefix(request, "DELETE ") {
					return fmt.Errorf("a protected source was deleted: %s", request)
				}
			}
			return nil
		},
	})
}
