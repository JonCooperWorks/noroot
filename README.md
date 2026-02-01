# noroot

## Overview

This repository provides a Go **library** and CLI to generate cloud-init configs and optionally create servers via cloud APIs. It sets up VPS hosts with basic hardening:

- Disables root SSH login
- Creates an admin user with sudo and your SSH keys
- Creates a separate Docker user (no SSH, no sudo)
- Installs and configures Fail2ban, NTP/chrony, auditd
- Installs Docker (and unattended-upgrades on Ubuntu)

**OS images:** An `OSImage` interface returns cloud-init YAML; Ubuntu is built-in, and more can be registered.

**Cloud drivers:** Pluggable driver interface with [Hetzner Cloud](https://hetzner.cloud/?ref=rXrWWQ3PDimB) support out of the box.

## Library structure

- **`cloudinit`** – `OSImage` interface returns cloud-init YAML; `OSImageByName("ubuntu")` and `Register` for custom images.
- **`driver`** – Cloud provider interface: `Create`, `Get`, `Delete` with `CreateOptions` (name, image, type, location, user data, SSH keys).
- **`driver/hetzner`** – Hetzner Cloud API driver using [hcloud-go](https://github.com/hetznercloud/hcloud-go).

Use the library in your own code:

```go
import (
    "github.com/joncooperworks/noroot/cloudinit"
    "github.com/joncooperworks/noroot/driver"
    "github.com/joncooperworks/noroot/driver/hetzner"
)

// Get OS image and generate cloud-init YAML
data := cloudinit.Data{
    AdminUsername:  "admin",
    AdminSSHKeys:   []string{"ssh-ed25519 AAAA..."},
    DockerUsername: "dockeruser",
}
img, _ := cloudinit.OSImageByName(cloudinit.OSUbuntu)
yaml, _ := img.CloudInit(data)

// Create a server on Hetzner
drv := hetzner.New(os.Getenv("HETZNER_TOKEN"))
server, _ := drv.Create(ctx, driver.CreateOptions{
    Name:       "my-server",
    Image:      "ubuntu-24.04",
    ServerType: "cpx11",
    Location:   "hel1",
    UserData:   yaml,
    SSHKeys:    data.AdminSSHKeys,
})
```

## Usage

### Requirements

- Go 1.22+

### Generating cloud-init only

Generate a cloud-init YAML file (no API calls):

```bash
go run ./cmd/noroot                          # Ubuntu, writes cloud-init.yml
go build -o noroot ./cmd/noroot && ./noroot -os ubuntu -output cloud-init.yml
```

**Flags:**

| Flag | Default | Description |
|------|--------|-------------|
| `-adminkeyfile` | `$HOME/.ssh/id_rsa.pub` | Admin SSH public key file |
| `-output` | `cloud-init.yml` | Output path |
| `-adminusername` | `topman` | Admin user name |
| `-dockerusername` | `dockeruser` | Docker user name |
| `-os` | `ubuntu` | OS image (e.g. `ubuntu`) |

### Creating a server (Hetzner)

Generate cloud-init **and** create a server on [Hetzner Cloud](https://hetzner.cloud/?ref=rXrWWQ3PDimB). Get an API token from the [Console](https://hetzner.cloud/?ref=rXrWWQ3PDimB) → Security → API Tokens.

**Passing the token:** Prefer not putting the token in your shell history or env. Recommended:

- **Token file** – Store the token in a file outside the repo, restrict permissions, and pass it in:
  ```bash
  echo "your-api-token" > ~/.config/hetzner-token   # or another path
  chmod 600 ~/.config/hetzner-token
  ./noroot -driver hetzner -token "$(cat ~/.config/hetzner-token)" -name my-vps ...
  ```
- **Secret manager** – Load the token from your secret manager and pass it to `-token` or `HETZNER_TOKEN`:
  ```bash
  # Examples (adjust to your tool):
  export HETZNER_TOKEN=$(op read "op://Vault/Hetzner/credential" 2>/dev/null)   # 1Password CLI
  export HETZNER_TOKEN=$(pass show hetzner/api-token 2>/dev/null)               # pass
  ./noroot -driver hetzner -name my-vps ...
  ```

**Example – Ubuntu 24 in Nuremberg, 4GB AMD (cax11):**

```bash
go build -o noroot ./cmd/noroot
./noroot -driver hetzner -token "$(cat ~/.config/hetzner-token)" -os ubuntu -name my-server -image ubuntu-24.04 -type cax11 -location nbg1
```

**Additional flags when using `-driver hetzner`:**

| Flag | Default | Description |
|------|--------|-------------|
| `-driver` | (none) | Cloud driver, e.g. `hetzner` |
| `-token` | `$HETZNER_TOKEN` | API token |
| `-name` | `noroot-server` | Server name |
| `-image` | `ubuntu-24.04` | Image name (e.g. `ubuntu-24.04`) |
| `-type` | `cpx11` | Server type (e.g. `cpx11`, `cax11`) |
| `-location` | `hel1` | Location (e.g. `hel1`, `fsn1`, `nbg1` Nuremberg) |

### Security model

SSH is restricted to the admin user (key-only). The Docker user has no SSH and no sudo; use the admin user to manage the host and run containers as the Docker user. Fail2ban, auditd, and (on Ubuntu) unattended-upgrades add extra hardening.
