// Package fixture contains corrected analysis-only examples.
package fixture

import (
	"crypto/tls"
	"errors"
	"os/exec"
)

// Command allows a single constant executable and fixed arguments, without shell.
func Command(input string) (*exec.Cmd, error) {
	if input != "status" {
		return nil, errors.New("unsupported action")
	}
	return exec.Command("/usr/bin/true"), nil
}

// TLS retains certificate verification and requires TLS 1.2 or newer.
func TLS() *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS12}
}
