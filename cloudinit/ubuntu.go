package cloudinit

import (
	"strings"
	"text/template"
)

// UbuntuOS is the Ubuntu OS image (apt-based).
type UbuntuOS struct{}

const ubuntuTemplate = `#cloud-config
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
{{range .AdminSSHKeys}}      - {{.}}
{{end}}
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
  - apt-get install -y docker.io docker-compose-plugin
  - gpasswd -a {{.DockerUsername}} docker
  - systemctl enable fail2ban
  - systemctl start fail2ban
  - systemctl enable ntp
  - systemctl start ntp
  - dpkg-reconfigure -plow unattended-upgrades
  - systemctl enable auditd
  - systemctl start auditd
  - mkdir -p /home/{{.AdminUsername}}/.ssh
{{range .AdminSSHKeys}}  - echo '{{.}}' >> /home/{{$.AdminUsername}}/.ssh/authorized_keys
{{end}}  - chown -R {{.AdminUsername}}:{{.AdminUsername}} /home/{{.AdminUsername}}/.ssh
  - chmod 600 /home/{{.AdminUsername}}/.ssh/authorized_keys
  - mkdir -p /home/{{.DockerUsername}}/.ssh
  - chown -R {{.DockerUsername}}:{{.DockerUsername}} /home/{{.DockerUsername}}/.ssh
  - chmod 700 /home/{{.DockerUsername}}/.ssh
  - touch /home/{{.DockerUsername}}/.hushlogin
  - systemctl enable docker
  - reboot

final_message: "The system is finally up, after $UPTIME seconds"
`

// CloudInit returns cloud-init YAML for Ubuntu.
func (UbuntuOS) CloudInit(d Data) (string, error) {
	tmpl, err := template.New("ubuntu").Parse(ubuntuTemplate)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, d); err != nil {
		return "", err
	}
	return out.String(), nil
}
