# noroot

## Overview
This repository contains a `cloud-init.yaml` file and a Go program to set up VPS hosts with some basic hardening measures.
The cloud-init configuration does the following:

- Disables root SSH login
- Creates a new user with sudo privileges
- Adds your SSH key to the new user

## Usage

### Requirements

- Go (https://golang.org/doc/install)

### Generating `cloud-init.yaml`

The repository includes a Go program, `noroot.go`, which generates the `cloud-init.yaml` file with your SSH key.
Simply run `go run noroot.go`

You can specify a different SSH public key file, username, and output file path using the -keyfile, -username, and -output flags:
./noroot -keyfile /path/to/your/id_rsa.pub -username yourusername -output /path/to/output/cloud-init.yaml

