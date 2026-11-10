// Package fixture contains analysis-only examples. No main or init functions.
package fixture

import (
	"crypto/tls"
	"os/exec"
)

// Command constructs a shell invocation but NEVER executes it.
func Command(input string) *exec.Cmd {
	return exec.Command("sh", "-c", input)
}

// TLS deliberately disables server identity verification for SAST demonstration.
func TLS() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}
