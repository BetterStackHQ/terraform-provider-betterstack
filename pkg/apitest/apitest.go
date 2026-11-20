// Package apitest is test support for the bridges built on this provider (notably Pulumi): the
// in-memory mock API the provider's own acceptance tests run against, and the per-resource smoke
// fixtures they use, re-exported from the internal package the way pkg/shim re-exports the
// provider.
package apitest

import (
	"github.com/betterstackhq/terraform-provider-betterstack/internal/provider"
)

// MockAPI serves every described endpoint in memory, speaking each API's declared envelope.
type MockAPI = provider.MockAPI

// SmokeFixture is one managed resource's sample create, update and replace inputs.
type SmokeFixture = provider.SmokeFixture

// SmokeSeed is the pre-existing record an adopting resource expects.
type SmokeSeed = provider.SmokeSeed

// NewMockAPI starts a mock server; the caller must Close it.
func NewMockAPI() *MockAPI { return provider.NewMockAPI() }

// Fixtures returns the smoke fixtures, one per managed resource, as a fresh slice.
func Fixtures() []SmokeFixture { return append([]SmokeFixture(nil), provider.SmokeFixtures...) }
