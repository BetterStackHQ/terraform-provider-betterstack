package main

import (
	"context"
	"flag"
	"log"

	"github.com/betterstackhq/terraform-provider-betterstack/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set by the release build (goreleaser ldflags); "dev" identifies local builds.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with debugger support (see Terraform plugin debugging docs)")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/BetterStackHQ/betterstack",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
