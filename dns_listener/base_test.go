package dns_listener

import (
	"os"
	"testing"

	_ "github.com/exiguus/ns-checker/internal/testinit"
)

func TestMain(m *testing.M) {
	// Setup will be called automatically by testing package
	os.Exit(m.Run())
}
