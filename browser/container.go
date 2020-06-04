package browser

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var DefaultContainerMemory = `512m`
var DefaultContainerSharedMemory = `256m`

type Container struct {
	dockertest.RunOptions
	ImageName    string
	Memory       string
	SharedMemory string
	Volumes      []string
	Publish      []string
	Privileged   bool
	Endpoint     string
	TlsCertPath  string
	validated    bool
	memory       int64
	shmSize      int64
	pool         *dockertest.Pool
	resource     *dockertest.Resource
}

func (self *Container) Start() error {
	if self.pool == nil {
		return fmt.Errorf("invalid endpoint")
	} else if !self.validated {
		if err := self.Validate(self.pool); err != nil {
			return err
		}
	}

	self.Repository, self.Tag = stringutil.SplitPair(self.ImageName, `:`)

	if res, err := self.pool.RunWithOptions(&self.RunOptions, self.dcHostConfig); err == nil {
		self.resource = res
		return nil
	} else {
		return fmt.Errorf("container start failed: %v", err)
	}
}

func (self *Container) IsRunning() bool {
	if res := self.resource; res != nil {
		if c := res.Container; c != nil {
			if c.State.Running {
				return true
			}
		}
	}

	return false
}

func (self *Container) Stop() error {
	if !self.IsRunning() {
		return nil
	}

	return self.resource.Close()
}

func (self *Container) dcHostConfig(cfg *docker.HostConfig) {
	if !self.validated {
		panic("cannot start container: config not validated")
	}

	if cfg != nil {
		cfg.AutoRemove = true
		cfg.Memory = self.memory
		cfg.ShmSize = self.shmSize
		cfg.Privileged = self.Privileged
		cfg.PortBindings = make(map[docker.Port][]docker.PortBinding)

		for _, portspec := range self.Publish {
			outer, inner := stringutil.SplitPair(portspec, `:`)

			cfg.PortBindings[docker.Port(inner)] = []docker.PortBinding{
				{
					HostPort: outer,
				},
			}
		}
	}
}

func (self *Container) Validate(pool *dockertest.Pool) error {
	if pool != nil {
		self.pool = pool
	} else {
		return fmt.Errorf("container: must configure a Docker endpoint")
	}

	if self.ImageName == `` {
		return fmt.Errorf("container: must specify an image name")
	}

	if self.Name == `` {
		self.Name = `webfriend-` + stringutil.UUID().Base58()
	}

	if self.Hostname == `` {
		self.Hostname = self.Name
	}

	if self.Memory == `` {
		self.Memory = DefaultContainerMemory
	}

	if self.SharedMemory == `` {
		self.SharedMemory = DefaultContainerSharedMemory
	}

	if v, err := humanize.ParseBytes(self.Memory); err == nil {
		self.memory = int64(v)
	} else {
		return fmt.Errorf("container-memory: %v", err)
	}

	if v, err := humanize.ParseBytes(self.SharedMemory); err == nil {
		self.shmSize = int64(v)
	} else {
		return fmt.Errorf("container-shm-size: %v", err)
	}

	self.validated = true
	return nil
}
