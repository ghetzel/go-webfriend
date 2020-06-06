package browser

import (
	"fmt"
	"net"

	"github.com/ghetzel/go-stockutil/typeutil"
)

type Container interface {
	Start() error
	Config() *ContainerConfig
	Validate() error
	Address() string
	IsRunning() bool
	Stop() error
	ID() string
	String() string
}

type ContainerConfig struct {
	Hostname     string
	Namespace    string
	Name         string
	User         string
	Env          []string
	Cmd          []string
	ImageName    string
	Memory       string
	SharedMemory string
	Ports        []string
	Volumes      []string
	Labels       map[string]string
	Privileged   bool
	WorkingDir   string
	UserDirPath  string
	TargetAddr   string
}

func (self *ContainerConfig) Validate() error {
	if self.Name == `` {
		return fmt.Errorf("container config: must specify a name")
	}

	if self.ImageName == `` {
		return fmt.Errorf("container config: must specify a container image")
	}

	if len(self.Cmd) == 0 {
		return fmt.Errorf("container config: must provide a command to run inside the container")
	}

	return nil
}

func (self *ContainerConfig) SetTargetPort(port int) string {
	if h, _, err := net.SplitHostPort(self.TargetAddr); err == nil {
		self.TargetAddr = net.JoinHostPort(h, typeutil.String(port))
	}

	return self.TargetAddr
}

func (self *ContainerConfig) AddPort(outer int, inner int, proto string) {
	if proto == `` {
		proto = `tcp`
	}

	self.Ports = append(self.Ports, fmt.Sprintf("%d:%d/%s", outer, inner, proto))
}
