package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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
  - name: %s
    sudo: ['ALL=(ALL) NOPASSWD:ALL']
    shell: /bin/bash
    ssh-authorized-keys:
      - %s

ssh_pwauth: false

disable_root: true

chpasswd:
  list: |
     %s:password
  expire: false

runcmd:
  - systemctl enable fail2ban
  - systemctl start fail2ban
  - systemctl enable ntp
  - systemctl start ntp
  - dpkg-reconfigure -plow unattended-upgrades
  - systemctl enable auditd
  - systemctl start auditd
  - mkdir -p /home/%s/.ssh
  - echo '%s' > /home/%s/.ssh/authorized_keys
  - chown -R %s:%s /home/%s/.ssh
  - chmod 600 /home/%s/.ssh/authorized_keys

final_message: "The system is finally up, after $UPTIME seconds"
`

func main() {
	// Parse command line arguments
	keyFile := flag.String("keyfile", os.Getenv("HOME")+"/.ssh/id_rsa.pub", "Path to the SSH public key file")
	outputFile := flag.String("output", "cloud-init.yml", "Path to the output YAML file")
	username := flag.String("username", "topman", "Username for the new user")
	flag.Parse()

	// Read the SSH key from the specified file
	key, err := readSSHKey(*keyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SSH key: %v\n", err)
		os.Exit(1)
	}

	// Generate the cloud-init YAML content
	yamlContent := fmt.Sprintf(cloudInitTemplate, *username, key, *username, *username, key, *username, *username, *username, *username, *username)

	// Write the YAML content to the output file
	if err := ioutil.WriteFile(*outputFile, []byte(yamlContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing YAML file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully wrote cloud-init config to %s\n", *outputFile)
}

// readSSHKey reads the SSH key from the given file path
func readSSHKey(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no SSH key found in file %s", filePath)
}