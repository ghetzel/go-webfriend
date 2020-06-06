package browser

import (
	"fmt"
	"strings"
	"time"

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

type KubernetesContainer struct {
	*ContainerConfig
	k8s        *kubernetes.Clientset
	id         string
	kubeConfig string
	memory     resource.Quantity
	pod        *v1.Pod
}

func NewKubernetesContainer() *KubernetesContainer {
	return &KubernetesContainer{
		ContainerConfig: &ContainerConfig{},
		kubeConfig:      DefaultKubeConfigFile,
	}
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

	if !self.memory.IsZero() {
		resources.Requests[v1.ResourceMemory] = self.memory
		resources.Limits[v1.ResourceMemory] = self.memory
	}

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

		for !self.IsRunning() {
			log.Infof("not running")
			time.Sleep(time.Second)
		}

		log.Noticef("running")
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
		inner, p := stringutil.SplitPair(inner, `/`)
		var port = v1.ContainerPort{
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
		return self.TargetAddr
	} else {
		return ``
	}
}

func (self *KubernetesContainer) IsRunning() bool {
	if self.pod == nil {
		return false
	}

	if p, err := self.podapi().UpdateStatus(self.pod); err == nil {
		self.pod = p
	}

	return (self.pod.Status.Phase == v1.PodRunning)
}

func (self *KubernetesContainer) Stop() error {
	return self.podapi().Delete(self.pod.Name, nil)
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
