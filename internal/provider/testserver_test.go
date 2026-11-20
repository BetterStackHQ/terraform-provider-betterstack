package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testProviderFactories serves the provider in-process over protocol 6, so acceptance tests never
// download anything.
var testProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"betterstack": providerserver.NewProtocol6WithError(New("test")()),
}

func providerConfig(endpoint string) string {
	return fmt.Sprintf("provider \"betterstack\" {\n  api_token = \"test-token\"\n  endpoint  = %q\n}\n\n", endpoint)
}

// newMockAPI ties a MockAPI (see mockapi.go) to the test's lifetime.
func newMockAPI(t *testing.T) *MockAPI {
	t.Helper()
	m := NewMockAPI()
	t.Cleanup(m.Close)
	return m
}
