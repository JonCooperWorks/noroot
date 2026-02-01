package cloudinit

// Data holds the common inputs for generating cloud-init config for any distro.
type Data struct {
	AdminUsername   string
	AdminSSHKeys    []string
	DockerUsername  string
}
