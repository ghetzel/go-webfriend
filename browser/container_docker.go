package browser

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/dustin/go-humanize"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type DockerContainer struct {
	*ContainerConfig
	validated bool
	memory    int64
	shmSize   int64
	client    *docker.Client
	endpoint  string
	id        string
}

func NewDockerContainer(url string) *DockerContainer {
	return &DockerContainer{
		ContainerConfig: &ContainerConfig{
			TargetAddr: DefaultContainerTargetAddr + `:` + typeutil.String(DebuggerInnerPort),
		},
		endpoint: url,
	}
}

func (self *DockerContainer) ID() string {
	return self.id
}

func (self *DockerContainer) String() string {
	return self.Name
}

func (self *DockerContainer) Config() *ContainerConfig {
	return self.ContainerConfig
}

func (self *DockerContainer) Start() error {
	if self.client == nil {
		var perr error

		if self.endpoint == `` {
			self.client, perr = docker.NewEnvClient()
		} else {
			self.client, perr = docker.NewClient(
				self.endpoint,
				``,
				nil,
				nil,
			)
		}

		if perr != nil {
			return fmt.Errorf("docker: %v", perr)
		}
	} else if !self.validated {
		if err := self.Validate(); err != nil {
			return err
		}
	}

	if res, err := self.client.ContainerCreate(
		context.Background(),
		&container.Config{
			Hostname:     self.Hostname,
			User:         self.User,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Env:          self.Env,
			Cmd:          strslice.StrSlice(self.Cmd),
			Image:        self.ImageName,
			Volumes:      self.volmap(),
			WorkingDir:   self.WorkingDir,
			Labels:       self.Labels,
			ExposedPorts: self.portset(),
		},
		&container.HostConfig{
			AutoRemove:   true,
			PortBindings: self.portmap(),
			Privileged:   self.Privileged,
			ShmSize:      self.shmSize,
			IpcMode:      container.IpcMode(`private`),
			Resources: container.Resources{
				Memory: self.memory,
			},
		},
		nil,
		nil,
		self.Name,
	); err == nil {
		// for _, warn := range res.Warnings {
		// 	log.Warningf("docker: %s", warn)
		// }

		if res.ID != `` {
			self.id = res.ID

			if err := self.client.ContainerStart(
				context.Background(),
				self.id,
				types.ContainerStartOptions{},
			); err == nil {
				go self.logtail()
				return nil
			} else {
				return fmt.Errorf("container start failed: %v", err)
			}
		} else {
			return fmt.Errorf("container start failed: no ID returned")
		}
	} else {
		return fmt.Errorf("container create failed: %v", err)
	}
}

func (self *DockerContainer) outerPort(innerPort int) string {
	for _, portspec := range self.Ports {
		outer, inner := stringutil.SplitPair(portspec, `:`)

		if inner == typeutil.String(innerPort) {
			return outer
		}
	}

	return ``
}

func (self *DockerContainer) volmap() map[string]struct{} {
	var out = make(map[string]struct{})

	for _, volspec := range self.Volumes {
		out[volspec] = struct{}{}
	}

	return out
}

func (self *DockerContainer) portset() nat.PortSet {
	var out = make(nat.PortSet)

	for port := range self.portmap() {
		out[port] = struct{}{}
	}

	return out
}

func (self *DockerContainer) portmap() map[nat.Port][]nat.PortBinding {
	var out = make(map[nat.Port][]nat.PortBinding)

	for _, portspec := range self.Ports {
		outer, inner := stringutil.SplitPair(portspec, `:`)

		if inner == `` {
			inner = outer
		}

		out[nat.Port(inner)] = []nat.PortBinding{
			{
				HostPort: outer,
			},
		}
	}

	return out
}

func (self *DockerContainer) logtail() {
	if self.IsRunning() {
		if rc, err := self.client.ContainerLogs(
			context.Background(),
			self.id,
			types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Timestamps: true,
				Follow:     true,
			},
		); err == nil {
			defer rc.Close()

			var linescan = bufio.NewScanner(rc)

			for linescan.Scan() {
				log.Debugf("[%s] %s", self.id, linescan.Text())
			}
		}
	}
}

func (self *DockerContainer) Address() string {
	if self.IsRunning() {
		return self.TargetAddr
	} else {
		return ``
	}
}

func (self *DockerContainer) IsRunning() bool {
	if self.client != nil {
		if self.id != `` {
			ctx, cn := context.WithTimeout(context.Background(), time.Second)
			defer cn()

			if s, err := self.client.ContainerStatPath(ctx, self.id, `/`); err == nil {
				return s.Mode.IsDir()
			}
		}
	}

	return false
}

func (self *DockerContainer) Stop() error {
	if self.IsRunning() {
		ctx, cn := context.WithTimeout(context.Background(), ProcessExitMaxWait)
		defer cn()

		if err := self.client.ContainerStop(ctx, self.id, &ProcessExitMaxWait); err == nil {
			ctx, cn := context.WithTimeout(context.Background(), ProcessExitMaxWait)
			defer cn()

			if err := self.client.ContainerRemove(ctx, self.id, types.ContainerRemoveOptions{}); err == nil {
				return nil
			} else if log.ErrContains(err, `already in progress`) {
				return nil
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return nil
	}

}

func (self *DockerContainer) Validate() error {
	if self.ImageName == `` {
		return fmt.Errorf("container: must specify an image name")
	}

	if self.Name == `` {
		return fmt.Errorf("container: must be given a name")
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

	if self.UserDirPath == `` {
		self.UserDirPath = DefaultUserDirPath
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
