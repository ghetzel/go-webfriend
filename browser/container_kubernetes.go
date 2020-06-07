package browser

import (
	"fmt"
	"strings"

	"github.com/ghetzel/go-stockutil/fileutil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1typed "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var DefaultKubeConfigFile = `~/.kube/config`
var KubeContainerInstanceName = `browser`

type k8sReady error

type KubernetesContainer struct {
	*ContainerConfig
	k8s            *kubernetes.Clientset
	id             string
	kubeConfig     string
	memory         resource.Quantity
	pod            *v1.Pod
	stopped        bool
	errchan        chan error
	firstOuterPort string
	lastError      error
}

func NewKubernetesContainer() *KubernetesContainer {
	return &KubernetesContainer{
		ContainerConfig: &ContainerConfig{},
		kubeConfig:      DefaultKubeConfigFile,
		errchan:         make(chan error, 1),
	}
}

func (self *KubernetesContainer) StartErr() <-chan error {
	return self.errchan
}

func (self *KubernetesContainer) ID() string {
	return self.id
}

func (self *KubernetesContainer) String() string {
	return self.Name
}

func (self *KubernetesContainer) Config() *ContainerConfig {
	return self.ContainerConfig
}

func (self *KubernetesContainer) Start() error {
	var kubeconf *rest.Config

	// work out config details
	if self.kubeConfig != `` {
		if kc, err := fileutil.ExpandUser(self.kubeConfig); err == nil {
			if cfg, err := clientcmd.BuildConfigFromFlags("", kc); err == nil {
				kubeconf = cfg
			} else {
				return fmt.Errorf("k8s: config: %v", err)
			}
		} else {
			return fmt.Errorf("k8s: config: %v", err)
		}
	} else if cfg, err := rest.InClusterConfig(); err == nil {
		kubeconf = cfg
	} else {
		return fmt.Errorf("k8s: config: %v", err)
	}

	// initialize k8s client using config
	if clientset, err := kubernetes.NewForConfig(kubeconf); err == nil {
		self.k8s = clientset
	} else {
		return fmt.Errorf("k8s: client: %v", err)
	}

	// define and launch pod

	var resources = v1.ResourceRequirements{
		Requests: make(map[v1.ResourceName]resource.Quantity),
		Limits:   make(map[v1.ResourceName]resource.Quantity),
	}

	// if !self.memory.IsZero() {
	// 	resources.Requests[v1.ResourceMemory] = self.memory
	// 	resources.Limits[v1.ResourceMemory] = self.memory
	// }

	if pod, err := self.podapi().Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      self.Name,
			Namespace: self.Namespace,
			Labels:    self.Labels,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Hostname:      self.Hostname,
			Containers: []v1.Container{
				{
					Name:       KubeContainerInstanceName,
					Image:      self.ImageName,
					Command:    self.Cmd,
					WorkingDir: self.WorkingDir,
					Ports:      self.containerPorts(),
					Env:        self.envVars(),
					Resources:  resources,
				},
			},
		},
	}); err == nil {
		self.pod = pod
		self.id = pod.Name
		self.stopped = false

		return nil
	} else {
		return err
	}
}

func (self *KubernetesContainer) podapi() v1typed.PodInterface {
	return self.k8s.CoreV1().Pods(self.Namespace)
}

func (self *KubernetesContainer) containerPorts() (ports []v1.ContainerPort) {
	for _, port := range self.Ports {
		outer, inner := stringutil.SplitPair(port, `:`)

		if self.firstOuterPort == `` {
			self.firstOuterPort = outer
		}

		inner, p := stringutil.SplitPair(inner, `/`)
		var port = v1.ContainerPort{
			Name:          `cdp-debugger`,
			HostPort:      int32(typeutil.Int(outer)),
			ContainerPort: int32(typeutil.Int(inner)),
		}

		switch strings.ToLower(p) {
		case `udp`:
			port.Protocol = v1.ProtocolUDP
		case `sctp`:
			port.Protocol = v1.ProtocolSCTP
		default:
			port.Protocol = v1.ProtocolTCP
		}

		ports = append(ports, port)
	}

	return
}

func (self *KubernetesContainer) envVars() (envs []v1.EnvVar) {
	for _, env := range self.Env {
		k, v := stringutil.SplitPair(env, `=`)

		envs = append(envs, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	return
}

func (self *KubernetesContainer) Address() string {
	if self.IsRunning() {
		// TODO: need to work out the correct method of ascertaining the IP;
		// best guess right now: detect if we're "in cluster", thus can use the PodIP,
		// else, use the HostIP.
		return self.pod.Status.HostIP + `:` + self.firstOuterPort
	} else {
		return ``
	}
}

func (self *KubernetesContainer) IsRunning() bool {
	if self.pod == nil {
		return false
	}

	if p, err := self.podapi().Get(self.pod.Name, metav1.GetOptions{}); err == nil {
		self.pod = p
		switch phase := p.Status.Phase; phase {
		case v1.PodRunning:
			return true
		case v1.PodPending:
			return false
		default:
			var merr error

			if r := p.Status.Reason; r != `` {
				merr = log.AppendError(merr, fmt.Errorf(r))
			}

			for _, status := range p.Status.ContainerStatuses {
				if wait := status.State.Waiting; wait != nil {
					if strings.Contains(wait.Reason, `Err`) || strings.Contains(wait.Reason, `Backoff`) {
						merr = log.AppendError(merr, fmt.Errorf("%s: %s", wait.Reason, wait.Message))
					}
				} else if term := status.State.Terminated; term != nil {
					merr = log.AppendError(merr, fmt.Errorf(
						"exited with code %d: %s - %s",
						term.ExitCode,
						term.Reason,
						term.Message,
					))
				}
			}

			if merr != nil {
				self.lastError = merr
			}
		}
	}

	return false
}

func (self *KubernetesContainer) Stop() error {
	if self.pod != nil {
		var podname = self.pod.Name
		self.pod = nil
		self.stopped = true

		return self.podapi().Delete(podname, nil)
	} else {
		return nil
	}
}

func (self *KubernetesContainer) Validate() error {
	if self.Namespace == `` {
		self.Namespace = `default`
	}

	if self.Memory != `` {
		if v, err := resource.ParseQuantity(self.Memory); err == nil {
			self.memory = v
		} else {
			return fmt.Errorf("container-memory: %v", err)
		}
	}

	return nil
}
