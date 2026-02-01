package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joncooperworks/noroot/cloudinit"
	"github.com/joncooperworks/noroot/driver"
	"github.com/joncooperworks/noroot/driver/hetzner"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

func main() {
	// Cloud-init generation flags
	adminKeyFile := flag.String("adminkeyfile", os.Getenv("HOME")+"/.ssh/id_ecdsa.pub", "Path to the admin SSH public key file")
	outputFile := flag.String("output", "cloud-init.yml", "Path to the output YAML file")
	adminUsername := flag.String("adminusername", "topman", "Username for the admin user")
	dockerUsername := flag.String("dockerusername", "dockeruser", "Username for the docker user")
	distro := flag.String("distro", cloudinit.DistroUbuntu, "Distro for cloud-init: ubuntu, fedora")

	// Optional: create server via cloud driver
	driverName := flag.String("driver", "", "Cloud driver to use when creating a server (e.g. hetzner). If empty, only generates cloud-init.")
	token := flag.String("token", "", "API token for the cloud driver (defaults to HETZNER_TOKEN env var)")
	serverName := flag.String("name", "noroot-server", "Server name when using -driver")
	image := flag.String("image", "ubuntu-24.04", "OS image (e.g. ubuntu-24.04, fedora-41)")
	serverType := flag.String("type", "cpx11", "Server type (e.g. cpx11, cax11)")
	location := flag.String("location", "hel1", "Location/datacenter (e.g. hel1, fsn1)")

	flag.Parse()

	// Fall back to environment variable if token not provided via flag
	if *token == "" {
		*token = os.Getenv("HETZNER_TOKEN")
	}

	// Validate usernames
	if !isValidUsername(*adminUsername) {
		log.Fatalf("Invalid admin username: %s. Must be 1-32 characters long and contain only lowercase letters, numbers, and underscores.\n", *adminUsername)
	}
	if !isValidUsername(*dockerUsername) {
		log.Fatalf("Invalid docker username: %s. Must be 1-32 characters long and contain only lowercase letters, numbers, and underscores.\n", *dockerUsername)
	}

	// Read SSH keys
	adminKeys, err := readSSHKeys(*adminKeyFile)
	if err != nil {
		log.Fatalf("Error reading admin SSH keys: %v\n", err)
	}

	data := cloudinit.Data{
		AdminUsername:  *adminUsername,
		AdminSSHKeys:   adminKeys,
		DockerUsername: *dockerUsername,
	}

	// Generate cloud-init YAML
	yamlContent, err := cloudinit.Render(data, *distro)
	if err != nil {
		log.Fatalf("Error generating cloud-init: %v\n", err)
	}

	// Validate YAML
	var out map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &out); err != nil {
		log.Printf("Generated YAML validation warning: %v\n", err)
	}

	// Write cloud-init to file
	if err := os.WriteFile(*outputFile, []byte(yamlContent), 0644); err != nil {
		log.Fatalf("Error writing cloud-init file: %v\n", err)
	}
	log.Printf("Wrote cloud-init config to %s (distro=%s)\n", *outputFile, *distro)

	// Optionally create server via driver
	if *driverName != "" {
		var drv driver.Driver
		switch *driverName {
		case "hetzner":
			if *token == "" {
				log.Fatal("Hetzner driver requires -token or HETZNER_TOKEN environment variable")
			}
			drv = hetzner.New(*token)
		default:
			log.Fatalf("Unknown driver %q. Supported: hetzner\n", *driverName)
		}

		opts := driver.CreateOptions{
			Name:       *serverName,
			Image:      *image,
			ServerType: *serverType,
			Location:   *location,
			UserData:   yamlContent,
			SSHKeys:    adminKeys,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		server, err := drv.Create(ctx, opts)
		if err != nil {
			log.Fatalf("Error creating server: %v\n", err)
		}
		log.Printf("Created server: id=%s name=%s public_ip=%s status=%s\n",
			server.ID, server.Name, server.PublicIP, server.Status)
	}
}

func readSSHKeys(filePath string) ([]string, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var keys []string
	scanner := bufio.NewScanner(strings.NewReader(string(fileContent)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if isValidSSHKey(line) {
			keys = append(keys, line)
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("no valid SSH keys found in file")
	}
	return keys, nil
}

func isValidUsername(username string) bool {
	validUsername := regexp.MustCompile(`^[a-z0-9_]{1,32}$`)
	return validUsername.MatchString(username)
}

func isValidSSHKey(key string) bool {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	return err == nil
}
