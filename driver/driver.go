package driver

import "context"

// CreateOptions holds common options for creating a cloud server.
type CreateOptions struct {
	Name       string   // Server name
	Image      string   // OS image (e.g. "ubuntu-24.04", "fedora-41")
	ServerType string   // Instance type (e.g. "cpx11", "cax11")
	Location   string   // Region/datacenter (e.g. "hel1", "fsn1")
	UserData   string   // Cloud-init or script content
	SSHKeys    []string // SSH public key strings (optional; some drivers use account keys)
}

// Server holds minimal info about a created server.
type Server struct {
	ID        string
	Name      string
	PublicIP  string
	PrivateIP string
	Status    string
}

// Driver is the interface that cloud providers implement.
type Driver interface {
	// Name returns the driver name (e.g. "hetzner").
	Name() string
	// Create creates a server with the given options and returns server details.
	Create(ctx context.Context, opts CreateOptions) (*Server, error)
	// Get returns a server by ID.
	Get(ctx context.Context, id string) (*Server, error)
	// Delete deletes a server by ID.
	Delete(ctx context.Context, id string) error
}
