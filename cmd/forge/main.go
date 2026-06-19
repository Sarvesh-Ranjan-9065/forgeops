// Command forge is the ForgeOps command-line entry point. It delegates all
// behavior to the internal/cli package.
package main

import "github.com/sarvesh-ranjan-9065/forgeops/internal/cli"

func main() {
	cli.Execute()
}
