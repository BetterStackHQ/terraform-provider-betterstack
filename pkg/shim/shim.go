// Package shim re-exports the plugin-framework provider so external modules (notably the Pulumi
// bridge) can construct it; the implementation itself lives in an internal package.
package shim

import (
	"github.com/betterstackhq/terraform-provider-betterstack/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	pfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
)

// New returns a fresh instance of the Better Stack Terraform plugin-framework provider, sending
// product/version as its User-Agent so the bridge's traffic is distinguishable from the provider's.
func New(product, version string) pfprovider.Provider {
	return provider.NewWithUserAgent(version, product+"/"+version)()
}

// NewForPulumi is New for the Pulumi bridge: the plan's replacement and deletion-protection
// messages then name properties as names spells them and give Pulumi's fixes.
func NewForPulumi(product, version string, names func(resourceType string, p path.Path) string) pfprovider.Provider {
	provider.UsePulumiMessages(names)
	return New(product, version)
}

// QuoteAPIValues writes API values in Markdown so the Pulumi bridge keeps them as written: a word
// that isn't one of names goes in single quotes, and true, false and null go bare.
func QuoteAPIValues(md string, names map[string]bool) string {
	return provider.QuoteAPIValues(md, names)
}

// SetPulumiDescriptionEdit rewrites the Terraform-only sentences of the schema's descriptions under
// Pulumi, before the bridge reads them.
func SetPulumiDescriptionEdit(edit func(string) string) { provider.SetPulumiDescriptionEdit(edit) }

// IDInput names a resource's id input for Pulumi, which reserves id: the calendar_id of a rotation.
type IDInput struct{ Resource, Name string }

// IDInputs lists every resource whose id is an input, with the name Pulumi users give it.
func IDInputs() []IDInput {
	return []IDInput{
		{"betterstack_errors_jira_integration", "integration_id"},
		{"betterstack_errors_linear_integration", "integration_id"},
		{"betterstack_errors_microsoft_teams_integration", "integration_id"},
		{"betterstack_errors_slack_integration", "integration_id"},
		{"betterstack_jira_integration", "integration_id"},
		{"betterstack_on_call_rotation", "calendar_id"},
	}
}
