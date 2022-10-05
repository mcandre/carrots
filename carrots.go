package carrots

import (
	"io"
	"fmt"
	"path"
	"path/filepath"
	"os"
	"regexp"
)

// SSHKeyPattern matches SSH key filenames.
var SSHKeyPattern = regexp.MustCompile("^id_.+$")

// SSHPublicKeyPattern matches SSH public key filenames.
var SSHPublicKeyPattern = regexp.MustCompile("^id_.+\\.pub$")

// Scanner collects warnings.
type Scanner struct {
	// Warnings denote an actionable permission discrepancy.
	Warnings []string

	// Home denotes the current user's home directory.
	Home string
}

// NewScanner constructs a scanner.
func NewScanner() (*Scanner, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return nil, err
	}

	return &Scanner{Home: home}, nil
}

// ScanSSH analyzes .ssh directories.
func (o Scanner) ScanSSH(pth string, info os.FileInfo) []string {
	if info.Name() == ".ssh" {
		mode := info.Mode() % 01000

		if mode != 0700 {
			return []string{fmt.Sprintf("%s: expected chmod 0700, got %04o", pth, mode)}
		}
	}

	return []string{}
}

// ScanSSHConfig analyzes .ssh/config files.
func (o Scanner) ScanSSHConfig(pth string, info os.FileInfo) []string {
	if info.Name() == "config" {
		parent := path.Base(filepath.Dir(pth))

		if parent == ".ssh" {
			mode := info.Mode() % 01000

			if mode != 0400 {
				return []string{fmt.Sprintf("%s: expected chmod 0400, got %04o", pth, mode)}
			}
		}
	}

	return []string{}
}

// ScanSSHKeys analyzes .ssh/id_.+(\.pub)? files.
func (o Scanner) ScanSSHKeys(pth string, info os.FileInfo) []string {
	name := info.Name()

	if SSHKeyPattern.MatchString(name) {
		parent := path.Base(filepath.Dir(pth))

		if parent == ".ssh" {
			mode := info.Mode() % 01000

			if SSHPublicKeyPattern.MatchString(name) {
				if mode != 0644 {
					return []string{fmt.Sprintf("%s: expected chmod 0644, got %04o", pth, mode)}
				}
			} else {
				if mode != 0600 {
					return []string{fmt.Sprintf("%s: expected chmod 0600, got %04o", pth, mode)}
				}
			}
		}
	}

	return []string{}
}

// ScanAuthorizedKeys analyzes authorized_keys files.
func (o Scanner) ScanSSHAuthorizedKeys(pth string, info os.FileInfo) []string {
	if info.Name() == "authorized_keys" {
		mode := info.Mode() % 01000

		if mode != 0600 {
			return []string{fmt.Sprintf("%s: expected chmod 0600, got %04o", pth, mode)}
		}
	}

	return []string{}
}

// ScanKnownHosts analyzes known_hosts files.
func (o Scanner) ScanSSHKnownHosts(pth string, info os.FileInfo) []string {
	if info.Name() == "known_hosts" {
		mode := info.Mode() % 01000

		if mode != 0644 {
			return []string{fmt.Sprintf("%s: expected chmod 0644, got %04o", pth, mode)}
		}
	}

	return []string{}
}

// ScanHome analyzes home directories.
func (o Scanner) ScanHome(pth string, info os.FileInfo) []string {
	if info.Name() == o.Home {
		mode := info.Mode() % 01000

		if mode != 0755 {
			return []string{fmt.Sprintf("%s: expected chmod 0755, got %04o", pth, mode)}
		}
	}

	return []string{}
}

// Walk traverses a file path recursively,
// collecting known permission discrepancies.
func (o *Scanner) Walk(pth string, info os.FileInfo, err error) error {
	o.Warnings = append(o.Warnings, o.ScanSSH(pth, info)...)
	o.Warnings = append(o.Warnings, o.ScanSSHConfig(pth, info)...)
	o.Warnings = append(o.Warnings, o.ScanSSHKeys(pth, info)...)
	o.Warnings = append(o.Warnings, o.ScanSSHAuthorizedKeys(pth, info)...)
	o.Warnings = append(o.Warnings, o.ScanSSHKnownHosts(pth, info)...)
	o.Warnings = append(o.Warnings, o.ScanHome(pth, info)...)
	return nil
}

// Scan checks the given root file path recursively
// for known permission discrepancies.
func Scan(root string) ([]string, error) {
	scanner, err := NewScanner()

	if err != nil {
		return []string{}, err
	}

	err = filepath.Walk(root, scanner.Walk)

	if err != nil && err != io.EOF {
		return scanner.Warnings, err
	}

	return scanner.Warnings, nil
}

// Report emits any warnings the console.
// If warnings are present, returns 1.
// Else, returns 0.
func Report(root string) int {
	warnings, err := Scan(root)

	for _, warning := range warnings {
		fmt.Println(warning)
	}

	if len(warnings) != 0 {
		return 1
	}

	if err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}
