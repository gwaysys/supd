package main

import (
	"github.com/gwaysys/supd/internal/cmd"
)

// Entry point for the supd binary. The repository previously kept this
// under cmd/supd/main.go (with a near-identical copy under
// cmd/supervisord/main.go); after the cmd/ removal it lives at the repo
// root so the canonical build command is `go build .` from the repo
// root.
func main() {
	cmd.Run()
}
