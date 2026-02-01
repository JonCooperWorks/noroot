package cloudinit

import "fmt"

// OSImage is an OS image that can produce cloud-init YAML.
type OSImage interface {
	// CloudInit returns cloud-init config as YAML for the given data.
	CloudInit(Data) (string, error)
}

// Built-in OS image names.
const (
	OSUbuntu = "ubuntu"
)

var registry = map[string]OSImage{
	OSUbuntu: UbuntuOS{},
}

// OSImageByName returns a registered OS image by name (e.g. "ubuntu").
func OSImageByName(name string) (OSImage, error) {
	img, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown OS image %q (supported: ubuntu)", name)
	}
	return img, nil
}

// Register adds an OS image implementation under the given name.
// Call from init() of packages that provide additional images.
func Register(name string, img OSImage) {
	registry[name] = img
}
