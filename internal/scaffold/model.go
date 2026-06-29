// Package scaffold renders embedded service templates into a destination
// directory tree. It is the core of the forge "new" command.
package scaffold

import (
	"fmt"
	"regexp"
)

// dnsLabel matches an RFC 1123 DNS label: lowercase alphanumerics and hyphens,
// neither starting nor ending with a hyphen.
var dnsLabel = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// ServiceSpec describes a single service to be scaffolded.
type ServiceSpec struct {
	// Name is the service name. It must be a valid lowercase DNS label because it
	// is reused for the repository name, Kubernetes objects, and ingress host.
	Name string
	// Lang is the implementation language. Only "go" is supported in v1.
	Lang string
	// Port is the TCP port the service listens on inside the container.
	Port int
	// Image is the container image repository (without tag) used in Helm values.
	Image string
}

// Default fills any zero-valued optional fields with their defaults. It doesnot
// validate; call Validate after defaulting.
func (s *ServiceSpec) Default() {
	if s.Lang == "" {
		s.Lang = "go"
	}
	if s.Port == 0 {
		s.Port = 8080
	}
	if s.Image == "" {
		s.Image = s.Name
	}
}

// Validate reports whether the spec is usable. It enforces a DNS-safe lowercase
// name, a supported language, and a valid port range.
func (s ServiceSpec) Validate() error {
	if !dnsLabel.MatchString(s.Name) {
		return fmt.Errorf("name %q must be a lowercase DNS label (a-z, 0-9, hyphen)", s.Name)
	}
	if len(s.Name) > 63 {
		return fmt.Errorf("name %q must be 63 characters or fewer", s.Name)
	}
	if s.Lang != "go" {
		return fmt.Errorf("lang %q is not supported; only go is supported in v1", s.Lang)
	}
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("port %d is out of range (1-65535)", s.Port)
	}
	return nil
}
