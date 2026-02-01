package hetzner

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/joncooperworks/noroot/driver"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/actionutil"
)

const driverName = "hetzner"

// Driver implements driver.Driver for Hetzner Cloud.
type Driver struct {
	client *hcloud.Client
}

// New creates a Hetzner cloud driver using the given API token.
func New(token string) *Driver {
	client := hcloud.NewClient(
		hcloud.WithToken(token),
		hcloud.WithApplication("noroot", "1.0"),
	)
	return &Driver{client: client}
}

// Name returns the driver name.
func (d *Driver) Name() string {
	return driverName
}

// Create creates a server on Hetzner Cloud with the given options.
func (d *Driver) Create(ctx context.Context, opts driver.CreateOptions) (*driver.Server, error) {
	createOpts := hcloud.ServerCreateOpts{
		Name:       opts.Name,
		ServerType: &hcloud.ServerType{Name: opts.ServerType},
		Location:   &hcloud.Location{Name: opts.Location},
		Image:      &hcloud.Image{Name: opts.Image},
		UserData:   opts.UserData,
	}

	// Resolve SSH keys: Hetzner accepts key names or fingerprints (existing keys in the account).
	// Raw public key strings can be added to Hetzner first via the console or API.
	if len(opts.SSHKeys) > 0 {
		keys := make([]*hcloud.SSHKey, 0, len(opts.SSHKeys))
		for i, s := range opts.SSHKeys {
			key, _, err := d.client.SSHKey.Get(ctx, s)
			if err != nil {
				// If it looks like a public key, create a one-off key for this server
				if strings.HasPrefix(s, "ssh-") || strings.HasPrefix(s, "ecdsa-") {
					created, _, createErr := d.client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
						Name:      fmt.Sprintf("%s-key-%d", opts.Name, i),
						PublicKey: s,
					})
					if createErr != nil {
						return nil, fmt.Errorf("create ssh key: %w", createErr)
					}
					if created != nil {
						keys = append(keys, created)
					}
					continue
				}
				return nil, fmt.Errorf("ssh key %q not found in Hetzner account (add it in Security → SSH Keys)", s)
			}
			if key != nil {
				keys = append(keys, key)
			}
		}
		createOpts.SSHKeys = keys
	}

	result, _, err := d.client.Server.Create(ctx, createOpts)
	if err != nil {
		return nil, err
	}

	// Wait for create action to complete
	allActions := actionutil.AppendNext(result.Action, result.NextActions)
	if err := d.client.Action.WaitFor(ctx, allActions...); err != nil {
		_, _, _ = d.client.Server.DeleteWithResult(ctx, result.Server)
		return nil, err
	}

	// Fetch server to get public IP
	server, _, err := d.client.Server.GetByID(ctx, result.Server.ID)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, fmt.Errorf("server not found after create")
	}

	return hcloudServerToDriver(server), nil
}

// Get returns a server by ID.
func (d *Driver) Get(ctx context.Context, id string) (*driver.Server, error) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid server id: %w", err)
	}
	server, _, err := d.client.Server.GetByID(ctx, n)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, fmt.Errorf("server %s not found", id)
	}
	return hcloudServerToDriver(server), nil
}

// Delete deletes a server by ID.
func (d *Driver) Delete(ctx context.Context, id string) error {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid server id: %w", err)
	}
	server, _, err := d.client.Server.GetByID(ctx, n)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("server %s not found", id)
	}
	_, _, err = d.client.Server.DeleteWithResult(ctx, server)
	return err
}

func hcloudServerToDriver(s *hcloud.Server) *driver.Server {
	out := &driver.Server{
		ID:     strconv.FormatInt(s.ID, 10),
		Name:   s.Name,
		Status: string(s.Status),
	}
	if s.PublicNet.IPv4.IP != nil {
		out.PublicIP = s.PublicNet.IPv4.IP.String()
	}
	for _, n := range s.PrivateNet {
		if n.IP != nil {
			out.PrivateIP = n.IP.String()
			break
		}
	}
	return out
}
