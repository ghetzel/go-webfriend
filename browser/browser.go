package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/ghetzel/argonaut"
	"github.com/ghetzel/friendscript/utils"
	"github.com/ghetzel/go-stockutil/executil"
	"github.com/ghetzel/go-stockutil/httputil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/pathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/husobee/vestigo"
	"github.com/mafredri/cdp/devtool"
	"github.com/mitchellh/go-ps"
	"github.com/phayes/freeport"
)

var DefaultStartWait = time.Second
var ProcessExitMaxWait = 10 * time.Second
var ProcessExitCheckInterval = 125 * time.Millisecond
var DefaultDebuggingPort = 0 // 0 = allocate an ephemeral port
var DefaultStartURL = `about:blank`
var DefaultContainerMemory = `512m`
var DefaultContainerSharedMemory = `256m`
var DebuggerInnerPort = 9222
var DefaultUserDirPath = `/var/tmp`
var ContainerInspectTimeout = 3 * time.Second
var DefaultContainerTargetAddr = `localhost`
var DefaultContainerRuntime = `docker`

var rpcGlobalTimeout = (60 * time.Second)
var rpcConnectTimeout = (60 * time.Second)
var rpcConnectRetryInterval = (250 * time.Millisecond)

type PathHandlerFunc = func(string) (string, io.Writer, bool)
type PathReaderFunc = func(string) (io.ReadCloser, bool)

var activeBrowserInstances sync.Map
var globalSignal = make(chan os.Signal, 1)
var globalStopping bool

func init() {
	// we keep track of all active & running browser instances.  if we get exit signals,
	// go through and stop them before exiting
	go executil.TrapSignals(func(sig os.Signal) bool {
		globalSignal <- sig
		StopAllActiveBrowsers()
		return false
	}, os.Interrupt, syscall.SIGTERM)
}

func StopAllActiveBrowsers() {
	if globalStopping {
		return
	} else {
		globalStopping = true
	}

	log.Debugf("[browser] Cleaning up active instances")

	activeBrowserInstances.Range(func(id interface{}, b interface{}) bool {
		if browser, ok := b.(*Browser); ok {
			log.Debugf("[%v] Cleaning up instance", id)
			browser.Stop()
		}

		return true
	})

	log.Debugf("[browser] Cleanup complete. Time to die.")
}

type Browser struct {
	Command                     argonaut.CommandName   `argonaut:",joiner=[=]"`
	App                         string                 `argonaut:"app,long"`
	DisableGPU                  bool                   `argonaut:"disable-gpu,long"`
	HideScrollbars              bool                   `argonaut:"hide-scrollbars,long"`
	Headless                    bool                   `argonaut:"headless,long"`
	Kiosk                       bool                   `argonaut:"kiosk,long"`
	ProxyBypassList             []string               `argonaut:"proxy-bypass-list,long,delimiters=[;]"`
	ProxyServer                 string                 `argonaut:"proxy-server,long"`
	RemoteDebuggingPort         int                    `argonaut:"remote-debugging-port,long"`
	RemoteDebuggingAddress      string                 `argonaut:"remote-debugging-address,long"`
	UserDataDirectory           string                 `argonaut:"user-data-dir,long"`
	DefaultBackgroundColor      string                 `argonaut:"default-background-color,long"`
	DisableSessionCrashedBubble bool                   `argonaut:"disable-session-crashed-bubble,long"`
	DisableInfobars             bool                   `argonaut:"disable-infobars,long"`
	SingleProcess               bool                   `argonaut:"single-process,long"`
	DisableSharedMemory         bool                   `argonaut:"disable-dev-shm-usage,long"`
	DisableSetuidSandbox        bool                   `argonaut:"disable-setuid-sandbox,long"`
	NoZygote                    bool                   `argonaut:"no-zygote,long"`
	NoSandbox                   bool                   `argonaut:"no-sandbox,long"`
	UserAgent                   string                 `argonaut:"user-agent,long"`
	URL                         string                 `argonaut:",positional"`
	StartWait                   time.Duration          `argonaut:"-"`
	Environment                 map[string]interface{} `argonaut:"-"`
	Directory                   string                 `argonaut:"-"`
	Preferences                 *Preferences           `argonaut:"-"`
	Container                   Container              `argonaut:"-"`
	ID                          string                 `argonaut:"-"`
	RemoteAddress               string
	cmd                         *exec.Cmd
	exitchan                    chan error
	devtools                    *devtool.DevTools
	router                      *vestigo.Router
	isTempUserDataDir           bool
	activeTabId                 string
	tabs                        map[string]*Tab
	tabLock                     sync.Mutex
	scopeable                   utils.Scopeable
	pathWriters                 []PathHandlerFunc
	pathReaders                 []PathReaderFunc
	stopped                     bool
	stopping                    bool
	dockerClient                *docker.Client
}

func NewBrowser() *Browser {
	return &Browser{
		ID:                  strings.ToLower(stringutil.UUID().Base58()),
		Command:             argonaut.CommandName(LocateChromeExecutable()),
		URL:                 DefaultStartURL,
		Headless:            true,
		RemoteDebuggingPort: 0,
		Preferences:         GetDefaultPreferences(),
		StartWait:           DefaultStartWait,
		exitchan:            make(chan error),
		tabs:                make(map[string]*Tab),
		pathWriters:         make([]PathHandlerFunc, 0),
		pathReaders:         make([]PathReaderFunc, 0),
	}
}

func Start() (*Browser, error) {
	var browser = NewBrowser()
	return browser, browser.Launch()
}

func (self *Browser) RegisterPathHandler(handler PathHandlerFunc) {
	self.pathWriters = append(self.pathWriters, handler)
}

func (self *Browser) RegisterPathReader(handler PathReaderFunc) {
	self.pathReaders = append(self.pathReaders, handler)
}

func (self *Browser) GetWriterForPath(path string) (string, io.Writer, bool) {
	for _, handler := range self.pathWriters {
		if p, w, ok := handler(path); ok {
			return p, w, true
		}
	}

	return ``, nil, false
}

func (self *Browser) GetReaderForPath(path string) (io.ReadCloser, error) {
	for _, handler := range self.pathReaders {
		if r, ok := handler(path); ok {
			return r, nil
		}
	}

	return os.Open(path)
}

func (self *Browser) SetScope(scopeable utils.Scopeable) {
	self.scopeable = scopeable
}

func (self *Browser) Launch() error {
	activeBrowserInstances.Store(self.ID, self)

	var remoteAddr = self.RemoteAddress

	// no remote address, so we're starting our own session
	if remoteAddr == `` {
		if self.UserDataDirectory == `` {
			if userDataDir, err := ioutil.TempDir(``, `webfriend-`); err == nil {
				self.UserDataDirectory = userDataDir
				self.isTempUserDataDir = true
			} else {
				return err
			}
		}

		if self.RemoteDebuggingPort <= 0 {
			if port, err := freeport.GetFreePort(); err == nil {
				self.RemoteDebuggingPort = port
			} else {
				return err
			}
		}

		if err := self.preparePaths(); err != nil {
			return err
		}

		// force disable sandboxing if the effective uid is 0 (root)
		if u, err := user.Current(); err == nil && typeutil.Int(u.Uid) == 0 {
			log.Debugf("Sandboxing is force disabled when running as root")
			self.NoSandbox = true
		}

		// manipulate various settings we'll need because we're going to run inside of a container
		// do this here because it'll be quite annoying to change them after argonaut.Command(self)
		if self.Container != nil {
			// disable sandboxing inside of container environment
			self.NoSandbox = true

			// allow remote debugging (port forwarding into the container requires this)
			self.RemoteDebuggingAddress = `0.0.0.0`
		}

		if cmd, err := argonaut.Command(self); err == nil {
			if self.isTempUserDataDir {
				if err := self.createFirstRunPreferences(); err != nil {
					return err
				}

				if err := self.preparePreferencesPrelaunch(); err != nil {
					return err
				}
			}

			if args := os.Getenv(`WEBFRIEND_BROWSER_ARGS`); args != `` {
				cmd.Args = append(cmd.Args, strings.Split(args, ` `)...)
			}

			for k, v := range self.Environment {
				self.cmd.Env = append(self.cmd.Env, fmt.Sprintf("%v=%v", k, v))
			}

			if self.Directory != `` {
				self.cmd.Dir = self.Directory
			}

			self.cmd = cmd

			// if requested, shove everything inside of a docker container
			if container := self.Container; container != nil {
				cfg := container.Config()
				cfg.Name = fmt.Sprintf("webfriend-%s", self.ID)
				cfg.Env = self.cmd.Env
				cfg.Cmd = append(self.cmd.Args)

				if err := cfg.Validate(); err != nil {
					return fmt.Errorf("invalid container config: %v", err)
				}

				if err := container.Validate(); err != nil {
					return fmt.Errorf("invalid container config: %v", err)
				}

				// rework the command we're running inside the container
				for i, arg := range cfg.Cmd {
					switch k, v := stringutil.SplitPair(arg, `=`); k {
					case `--user-data-dir`:
						cfg.Volumes = append(cfg.Volumes, fmt.Sprintf("%s:%s", v, cfg.UserDirPath))
						cfg.Cmd[i] = k + `=` + cfg.UserDirPath

					case `--remote-debugging-port`:
						cfg.AddPort(self.RemoteDebuggingPort, DebuggerInnerPort, ``)
						cfg.Cmd[i] = k + `=` + typeutil.String(DebuggerInnerPort)
						cfg.SetTargetPort(self.RemoteDebuggingPort)
					}
				}

				if err := container.Start(); err == nil {
					self.stopped = false
					log.Debugf("[%s] Container %v starting...", self.ID, container)

					select {
					case sig := <-globalSignal:
						if err := container.Stop(); err == nil {
							return fmt.Errorf("Container startup interrupted with signal %v", sig)
						} else {
							return err
						}

					case <-time.After(self.StartWait):
						if addr := container.Address(); addr != `` {
							log.Debugf("[%s] Container stayed running for %v", self.ID, self.StartWait)
							remoteAddr = addr
						} else {
							return fmt.Errorf("No debugger available after %v", self.StartWait)
						}
					}
				} else {
					return err
				}
			} else {
				self.cmd.Stdout = httputil.NewWritableLogger(httputil.Info, `[PROC] `)
				self.cmd.Stderr = httputil.NewWritableLogger(httputil.Warning, `[PROC] `)

				// launch the browser
				go func() {
					log.Debugf("[%s] Executing: %v", self.ID, strings.Join(self.cmd.Args, ` `))
					self.stopped = false
					self.exitchan <- self.cmd.Run()
				}()

				select {
				case err := <-self.exitchan:
					if err == nil {
						if eerr, ok := err.(*exec.ExitError); ok {
							if status, ok := eerr.Sys().(syscall.WaitStatus); ok {
								err = fmt.Errorf("Process exited prematurely with status %d", status.ExitStatus())
							} else if eerr.Success() {
								err = fmt.Errorf("Process exited prematurely without error")
							}
						}

						if err == nil {
							err = fmt.Errorf("Process exited prematurely with non-zero status")
						}
					}

					return err
				case <-time.After(self.StartWait):
					log.Debugf("[%s] Process stayed running for %v", self.ID, self.StartWait)
					remoteAddr = `localhost:` + typeutil.String(self.RemoteDebuggingPort)
					self.stopped = false
				}
			}
		} else {
			return err
		}
	}

	if err := self.connectRPC(remoteAddr); err == nil {
		return nil
	} else {
		return self.stopWithError(err)
	}
}

func (self *Browser) ctx() context.Context {
	c := context.Background()
	// c, _ = context.WithTimeout(c, rpcGlobalTimeout)
	return c
}

func (self *Browser) Tab() *Tab {
	if self.activeTabId != `` {
		self.tabLock.Lock()
		defer self.tabLock.Unlock()

		if tab, ok := self.tabs[self.activeTabId]; ok {
			return tab
		}
	}

	panic("No active tab")
}

func (self *Browser) Wait() error {
	return <-self.exitchan
}

func (self *Browser) Stop() error {
	return self.stopWithError(nil)
}

func (self *Browser) stopWithError(xerr error) error {
	if self.stopping {
		return nil
	}

	self.stopped = true
	self.stopping = true

	defer func() {
		self.cleanupUserDataDirectory()
		self.cleanupActiveReference()
		self.stopping = false
	}()

	if self.cmd == nil {
		return nil
	}

	self.tabLock.Lock()
	defer self.tabLock.Unlock()

	log.Debugf("[%v] Stopping Webfriend...", self.ID)

	for _, tab := range self.tabs {
		tab.Disconnect()
	}

	// take different paths depending on whether we're containerized or not
	if container := self.Container; container != nil {
		log.Debugf("[%s] Stopping container %v", container.ID(), container)

		if err := container.Stop(); err == nil {
			return xerr
		} else {
			log.Error(err)
			return err
		}
	} else {
		if process := self.cmd.Process; process == nil {
			return fmt.Errorf("Process not running")
		} else {
			log.Debugf("[%s] Killing browser process %d", self.ID, process.Pid)

			if err := process.Kill(); err == nil {
				var started = time.Now()
				var deadline = started.Add(ProcessExitMaxWait)

				for t := started; t.Before(deadline); t = time.Now() {
					if proc, err := ps.FindProcess(process.Pid); err == nil && proc == nil {
						log.Debugf("[%s] PID %d is gone", self.ID, process.Pid)
						return xerr
					}

					log.Debugf("[%s] Polling for PID %d to disappear", self.ID, process.Pid)
					time.Sleep(ProcessExitCheckInterval)
				}

				return fmt.Errorf("[%s] Could not confirm process %d exited", self.ID, process.Pid)
			} else {
				return err
			}
		}
	}
}

func (self *Browser) cleanupUserDataDirectory() error {
	if self.isTempUserDataDir && pathutil.DirExists(self.UserDataDirectory) {
		log.Debugf("[%s] Cleaning up temporary profile %s", self.ID, self.UserDataDirectory)
		return os.RemoveAll(self.UserDataDirectory)
	}

	return nil
}

func (self *Browser) preparePaths() error {
	if self.UserDataDirectory != `` {
		if dir, err := pathutil.ExpandUser(self.UserDataDirectory); err == nil {
			self.UserDataDirectory = dir
		} else {
			return err
		}
	}

	return nil
}

func (self *Browser) cleanupActiveReference() {
	activeBrowserInstances.Delete(self.ID)
}

func (self *Browser) createFirstRunPreferences() error {
	if self.UserDataDirectory != `` && self.Preferences != nil {
		if err := os.MkdirAll(self.UserDataDirectory, 0700); err == nil {
			if file, err := os.Create(path.Join(self.UserDataDirectory, `First Run`)); err == nil {
				defer file.Close()

				if err := json.NewEncoder(file).Encode(self.Preferences); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (self *Browser) preparePreferencesPrelaunch() error {
	prefsPath := path.Join(self.UserDataDirectory, `Default`, `Preferences`)

	if err := os.MkdirAll(path.Dir(prefsPath), 0700); err != nil {
		return err
	}

	if prefs, err := os.Create(prefsPath); err == nil {
		defer prefs.Close()

		if err := json.NewEncoder(prefs).Encode(self.Preferences); err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}
