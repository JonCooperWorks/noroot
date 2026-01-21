package main

import (
	"bufio"
	"errors"
	"flag"
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
  - apt-transport-https
  - ca-certificates
  - curl
  - software-properties-common

users:
  - name: {{.AdminUsername}}
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash
    ssh-authorized-keys:
      - {{.AdminSSHKey}}
  - name: {{.DockerUsername}}
    groups: docker
    shell: /bin/bash
    ssh_pwauth: false
    lock_passwd: true

ssh_pwauth: false

disable_root: true

chpasswd:
  list: |
     {{.AdminUsername}}:password
  expire: false

runcmd:
  - curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
  - add-apt-repository "deb [arch=arm64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
  - apt-get update
  - apt-get install -y docker.io docker-compose
  - gpasswd -a {{.DockerUsername}} docker
  - systemctl enable fail2ban
  - systemctl start fail2ban
  - systemctl enable ntp
  - systemctl start ntp
  - dpkg-reconfigure -plow unattended-upgrades
  - systemctl enable auditd
  - systemctl start auditd
  - mkdir -p /home/{{.AdminUsername}}/.ssh
  - echo '{{.AdminSSHKey}}' > /home/{{.AdminUsername}}/.ssh/authorized_keys
  - chown -R {{.AdminUsername}}:{{.AdminUsername}} /home/{{.AdminUsername}}/.ssh
  - chmod 600 /home/{{.AdminUsername}}/.ssh/authorized_keys
  - mkdir -p /home/{{.DockerUsername}}/.ssh
  - chown -R {{.DockerUsername}}:{{.DockerUsername}} /home/{{.DockerUsername}}/.ssh
  - chmod 700 /home/{{.DockerUsername}}/.ssh
  - touch /home/{{.DockerUsername}}/.hushlogin
  - systemctl enable docker
  - reboot

final_message: "The system is finally up, after $UPTIME seconds"
`

type CloudInitData struct {
	AdminUsername  string
	AdminSSHKey    string
	DockerUsername string
}

func main() {
	// Parse command line arguments
	adminKeyFile := flag.String("adminkeyfile", os.Getenv("HOME")+"/.ssh/id_rsa.pub", "Path to the admin SSH public key file")
	outputFile := flag.String("output", "cloud-init.yml", "Path to the output YAML file")
	adminUsername := flag.String("adminusername", "topman", "Username for the admin user")
	dockerUsername := flag.String("dockerusername", "dockeruser", "Username for the docker user")
	flag.Parse()

	// Validate the usernames
	if !isValidUsername(*adminUsername) {
		log.Fatalf("Invalid admin username: %s. Must be 1-32 characters long and contain only lowercase letters, numbers, and underscores.\n", *adminUsername)
	}
	if !isValidUsername(*dockerUsername) {
		log.Fatalf("Invalid docker username: %s. Must be 1-32 characters long and contain only lowercase letters, numbers, and underscores.\n", *dockerUsername)
	}

	// Read and validate the SSH keys from the specified files
	adminKey, err := readSSHKey(*adminKeyFile)
	if err != nil {
		log.Fatalf("Error reading admin SSH key: %v\n", err)
	}

	// Prepare data for the template
	data := CloudInitData{
		AdminUsername:  *adminUsername,
		AdminSSHKey:    adminKey,
		DockerUsername: *dockerUsername,
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
	var out map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlContent.String()), &out)
	if err != nil {
		log.Printf("Generated YAML content is invalid: %v\n", err)
	}

	// Write the YAML content to the output file
	if err := os.WriteFile(*outputFile, []byte(yamlContent.String()), 0644); err != nil {
		log.Fatalf("Error writing YAML file: %v\n", err)
	}

	log.Printf("Successfully wrote cloud-init config to %s\n", *outputFile)
}

// readSSHKey reads the SSH key from the given file path and validates its structure
func readSSHKey(filePath string) (string, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	key := getFirstLine(string(fileContent))
	if !isValidSSHKey(key) {
		return "", errors.New("invalid SSH key format")
	}
	return key, nil
}

// getFirstLine returns the first line of the given string.
func getFirstLine(input string) string {
	scanner := bufio.NewScanner(strings.NewReader(input))
	if scanner.Scan() {
		return scanner.Text()
	}
	return ""
}

// isValidUsername validates the username
func isValidUsername(username string) bool {
	validUsername := regexp.MustCompile(`^[a-z0-9_]{1,32}$`)
	return validUsername.MatchString(username)
}

// isValidSSHKey validates the SSH key
func isValidSSHKey(key string) bool {
	_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	return err == nil
}
