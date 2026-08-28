package supd

import (
	"github.com/gwaysys/supd/setup"
)

// SetupCommand is a type alias for setup.Command so that go-flags reflection
// sees the same positional/flag tags and Execute method, while keeping the
// `*SetupCommand` symbol used elsewhere in this package unchanged. The actual
// implementation lives in the github.com/gwaysys/supd/setup package.
type SetupCommand = setup.Command

func init() {
	setup.Register(parser)
}