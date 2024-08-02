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
