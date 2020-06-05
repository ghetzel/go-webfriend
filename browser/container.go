package browser

import (
	"bufio"
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/dustin/go-humanize"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type Container struct {
	ID           string
	Hostname     string
	Name         string
	User         string
	Env          []string
	Cmd          []string
	ImageName    string
	Memory       string
	SharedMemory string
	Publish      []string
	Volumes      []string
	Labels       map[string]string
	Privileged   bool
	WorkingDir   string
	Endpoint     string
	TlsCertPath  string
	UserDirPath  string
	TargetAddr   string
	OuterPort    int
	validated    bool
	memory       int64
	shmSize      int64
	client       *docker.Client
}

func (self *Container) Start() error {
	if self.client == nil {
		return fmt.Errorf("invalid endpoint")
	} else if !self.validated {
		if err := self.Validate(self.client); err != nil {
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
		},
		&container.HostConfig{
			AutoRemove:   true,
			PortBindings: self.portmap(),
			Privileged:   self.Privileged,
			ShmSize:      self.shmSize,
			Resources: container.Resources{
				Memory: self.memory,
			},
		},
		nil,
		self.Name,
	); err == nil {
		// for _, warn := range res.Warnings {
		// 	log.Warningf("docker: %s", warn)
		// }

		if res.ID != `` {
			self.ID = res.ID
			if err := self.client.ContainerStart(
				context.Background(),
				self.ID,
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

func (self *Container) outerPort(innerPort int) string {
	for _, portspec := range self.Publish {
		outer, inner := stringutil.SplitPair(portspec, `:`)

		if inner == typeutil.String(innerPort) {
			return outer
		}
	}

	return ``
}

func (self *Container) volmap() map[string]struct{} {
	var out = make(map[string]struct{})

	for _, volspec := range self.Volumes {
		out[volspec] = struct{}{}
	}

	return out
}

func (self *Container) portmap() map[nat.Port][]nat.PortBinding {
	var out = make(map[nat.Port][]nat.PortBinding)

	for _, portspec := range self.Publish {
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

func (self *Container) logtail() {
	if self.IsRunning() {
		if rc, err := self.client.ContainerLogs(
			context.Background(),
			self.ID,
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
				fmt.Println(linescan.Text())
			}
		}
	}
}

func (self *Container) DebuggerAddr() string {
	if self.IsRunning() {
		return sliceutil.OrString(self.TargetAddr, DefaultContainerTargetAddr) + `:` + typeutil.String(self.OuterPort)
	}

	return ``
}

func (self *Container) IsRunning() bool {
	if self.client != nil {
		if self.ID != `` {
			ctx, cn := context.WithTimeout(context.Background(), time.Second)
			defer cn()

			if s, err := self.client.ContainerStatPath(ctx, self.ID, `/`); err == nil {
				return s.Mode.IsDir()
			}
		}
	}

	return false
}

func (self *Container) Stop() error {
	if self.IsRunning() {
		ctx, cn := context.WithTimeout(context.Background(), ProcessExitMaxWait)
		defer cn()

		return self.client.ContainerKill(
			ctx,
			self.ID,
			syscall.SIGTERM.String(),
		)
	} else {
		return nil
	}

}

func (self *Container) Validate(client *docker.Client) error {
	if client != nil {
		self.client = client
	} else {
		return fmt.Errorf("container: must configure a Docker endpoint")
	}

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
