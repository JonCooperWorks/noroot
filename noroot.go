package main

import (
	"bufio"
	"flag"
    "fmt"
    "log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

const cloudInitTemplate = `#cloud-config
package_update: true
package_upgrade: true
packages:
  - fail2ban
  - ntp
  - unattended-upgrades
  - auditd

users:
  - name: {{.Username}}
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash
    ssh-authorized-keys:
      - {{.SSHKey}}

ssh_pwauth: false

disable_root: true

chpasswd:
  list: |
     {{.Username}}:password
  expire: false

runcmd:
  - systemctl enable fail2ban
  - systemctl start fail2ban
  - systemctl enable ntp
  - systemctl start ntp
  - dpkg-reconfigure -plow unattended-upgrades
  - systemctl enable auditd
  - systemctl start auditd
  - mkdir -p /home/{{.Username}}/.ssh
  - echo '{{.SSHKey}}' > /home/{{.Username}}/.ssh/authorized_keys
  - chown -R {{.Username}}:{{.Username}} /home/{{.Username}}/.ssh
  - chmod 600 /home/{{.Username}}/.ssh/authorized_keys

final_message: "The system is finally up, after $UPTIME seconds"
`

type CloudInitData struct {
	Username string
	SSHKey   string
}

func main() {
	// Parse command line arguments
	keyFile := flag.String("keyfile", os.Getenv("HOME")+"/.ssh/id_rsa.pub", "Path to the SSH public key file")
	outputFile := flag.String("output", "cloud-init.yml", "Path to the output YAML file")
	username := flag.String("username", "topman", "Username for the new user")
	flag.Parse()

	// Validate the username
	if !isValidUsername(*username) {
		log.Fatalf("Invalid username: %s. Must be 1-32 characters long and contain only lowercase letters, numbers, and underscores.\n", *username)
	}

	// Read and validate the SSH key from the specified file
	key, err := readSSHKey(*keyFile)
	if err != nil {
		log.Fatalf("Error reading SSH key: %v\n", err)
	}

	// Prepare data for the template
	data := CloudInitData{
		Username: *username,
		SSHKey:   key,
	}

	// Generate the cloud-init YAML content using the template
	tmpl, err := template.New("cloudInit").Parse(cloudInitTemplate)
	if err != nil {
		log.Fatalf("Error parsing template: %v\n", err)
	}

	var yamlContent strings.Builder
	if err := tmpl.Execute(&yamlContent, data); err != nil {
		log.Fatalf("Error executing template: %v\n", err)
	}

	// Validate the YAML content
	if !isValidYAML(yamlContent.String()) {
		log.Fatalf("Generated YAML content is invalid\n")
	}

	// Write the YAML content to the output file
	if err := os.WriteFile(*outputFile, []byte(yamlContent.String()), 0644); err != nil {
		log.Fatalf("Error writing YAML file: %v\n", err)
	}

	log.Printf("Successfully wrote cloud-init config to %s\n", *outputFile)
}

// readSSHKey reads the SSH key from the given file path and validates its structure
func readSSHKey(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		key := scanner.Text()
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
		if err != nil {
			return "", fmt.Errorf("invalid SSH key format: %v", err)
		}
		return key, nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no SSH key found in file %s", filePath)
}

// isValidUsername validates the username
func isValidUsername(username string) bool {
	validUsername := regexp.MustCompile(`^[a-z0-9_]{1,32}$`)
	return validUsername.MatchString(username)
}

// isValidYAML validates the YAML content
func isValidYAML(content string) bool {
	var out map[string]interface{}
	return yaml.Unmarshal([]byte(content), &out) == nil
}