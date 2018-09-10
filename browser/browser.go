package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"

	"github.com/ghetzel/argonaut"
	"github.com/ghetzel/friendscript/utils"
	"github.com/ghetzel/go-stockutil/httputil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/pathutil"
	"github.com/husobee/vestigo"
	"github.com/mafredri/cdp/devtool"
	"github.com/phayes/freeport"
)

var rpcGlobalTimeout = (60 * time.Second)
var DefaultStartWait = time.Duration(500) * time.Millisecond
var ProcessExitMaxWait = 10 * time.Second
var ProcessExitCheckInterval = 125 * time.Millisecond

type PathHandlerFunc = func(string) (string, io.Writer, bool)

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
	cmd                         *exec.Cmd
	exitchan                    chan error
	devtools                    *devtool.DevTools
	router                      *vestigo.Router
	isTempUserDataDir           bool
	activeTabId                 string
	tabs                        map[string]*Tab
	tabLock                     sync.Mutex
	scopeable                   utils.Scopeable
	pathHandlers                []PathHandlerFunc
}

func NewBrowser() *Browser {
	return &Browser{
		Command:             argonaut.CommandName(LocateChromeExecutable()),
		URL:                 `about:blank`,
		Headless:            true,
		RemoteDebuggingPort: 0,
		Preferences:         GetDefaultPreferences(),
		StartWait:           DefaultStartWait,
		exitchan:            make(chan error),
		tabs:                make(map[string]*Tab),
		pathHandlers:        make([]PathHandlerFunc, 0),
	}
}

func Start() (*Browser, error) {
	browser := NewBrowser()
	return browser, browser.Launch()
}

func (self *Browser) RegisterPathHandler(handler PathHandlerFunc) {
	self.pathHandlers = append(self.pathHandlers, handler)
}

func (self *Browser) GetWriterForPath(path string) (string, io.Writer, bool) {
	for _, handler := range self.pathHandlers {
		if p, w, ok := handler(path); ok {
			return p, w, true
		}
	}

	return ``, nil, false
}

func (self *Browser) SetScope(scopeable utils.Scopeable) {
	self.scopeable = scopeable
}

func (self *Browser) Launch() error {
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

		self.cmd = cmd

		for k, v := range self.Environment {
			self.cmd.Env = append(self.cmd.Env, fmt.Sprintf("%v=%v", k, v))
		}

		if self.Directory != `` {
			self.cmd.Dir = self.Directory
		}

		self.cmd.Stdout = httputil.NewWritableLogger(httputil.Info, `[PROC] `)
		self.cmd.Stderr = httputil.NewWritableLogger(httputil.Warning, `[PROC] `)

		// launch the browser
		go func() {
			log.Debugf("[browser] Executing: %v", strings.Join(self.cmd.Args, ` `))
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
			log.Debugf("[browser] Process stayed running for %v", self.StartWait)

			if err := self.connectRPC(fmt.Sprintf("127.0.0.1:%d", self.RemoteDebuggingPort)); err == nil {
				return nil
			} else {
				defer self.Stop()
				return err
			}
		}
	} else {
		return err
	}
}

func (self *Browser) ctx() context.Context {
	c := context.Background()
	c, _ = context.WithTimeout(c, rpcGlobalTimeout)
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
	self.tabLock.Lock()
	defer self.tabLock.Unlock()
	defer self.cleanupUserDataDirectory()

	log.Debug("[browser] Stopping process...")

	for _, tab := range self.tabs {
		tab.Disconnect()
	}

	if process := self.cmd.Process; process == nil {
		return fmt.Errorf("Process not running")
	} else {
		log.Debugf("[browser] Killing browser process %d", process.Pid)

		if err := process.Kill(); err == nil {
			started := time.Now()
			deadline := started.Add(ProcessExitMaxWait)

			for t := started; t.Before(deadline); t = time.Now() {
				if proc, err := ps.FindProcess(process.Pid); err == nil && proc == nil {
					log.Debugf("[browser] PID %d is gone", process.Pid)
					return nil
				}

				log.Debugf("[browser] Polling for PID %d to disappear", process.Pid)
				time.Sleep(ProcessExitCheckInterval)
			}

			return fmt.Errorf("Could not confirm process %d exited", process.Pid)
		} else {
			return err
		}
	}
}

func (self *Browser) cleanupUserDataDirectory() error {
	if self.isTempUserDataDir && pathutil.DirExists(self.UserDataDirectory) {
		log.Debugf("[browser] Cleaning up temporary profile %s", self.UserDataDirectory)
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
