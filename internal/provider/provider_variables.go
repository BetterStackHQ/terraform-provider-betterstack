//go:build !preview

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

// baseURLVariables are the provider settings that fill a templated base URL, one per variable the
// API description declares (a Cloud availability zone). A provider alias per value reaches several.
type baseURLVariables struct {
}

func baseURLVariableAttributes() map[string]schema.Attribute {
	return nil
}

// values are the configured settings, each falling back to its environment variable.
func (v baseURLVariables) values() map[string]string {
	return nil
}
