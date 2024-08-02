# noroot

## Overview
This repository contains a Go program to set up my VPS hosts with some basic hardening measures.
I'm open sourcing it in case someone else could find this useful.
The cloud-init configuration does the following:

- Disables root SSH login
- Creates a new user with sudo privileges
- Adds your SSH key to the new user
- Updates and upgrades the system packages
- Installs and configures Fail2Ban to protect against brute force attacks
- Sets up NTP to synchronize the system clock
- Configures unattended upgrades for automatic security updates
- Installs and configures audit logging with auditd
- Installs Docker and docker-compose

## Usage

### Requirements

- Go (https://golang.org/doc/install)

### Generating `cloud-init.yml`

The repository includes a Go program, `noroot.go`, which generates the `cloud-init.yml` file with your SSH key.
Simply run `go run noroot.go`

You can specify a different SSH public key file, username, and output file path using the -keyfile, -username, and -output flags:

 `./noroot -keyfile /path/to/your/id_rsa.pub -username yourusername -output /path/to/output/cloud-init.yml`

### Security Model

The security model restricts SSH access exclusively to the admin user using SSH key authentication, while the Docker user cannot access SSH and lacks sudo privileges.
This ensures only the admin can manage the server and switch to the Docker user for container operations.
Additional controls include `fail2ban` to protect against brute force attacks, `auditd` for detailed logging of system activities, and `unattended-upgrades` to automatically apply security updates.
