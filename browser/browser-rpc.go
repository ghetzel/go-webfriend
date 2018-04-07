package browser

import (
	"fmt"
	"reflect"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/mafredri/cdp/devtool"
)

var rpcConnectTimeout = (10 * time.Second)
var rpcConnectRetryInterval = (125 * time.Millisecond)
var rpcConnectMaxRetries = 40

func (self *Browser) connectRPC(address string) error {
	self.devtools = devtool.New(fmt.Sprintf("http://%v", address))
	connected := false

	for i := 0; i < rpcConnectMaxRetries; i++ {
		if version, err := self.devtools.Version(self.ctx()); err == nil {
			connected = true
			log.Debugf("Connected to %v; protocol %v", version.Browser, version.Protocol)
			break
		} else {
			time.Sleep(rpcConnectRetryInterval)
		}
	}

	if connected {
		return self.syncState()
	} else {
		return fmt.Errorf("Failed to connect to RPC interface after %d attempts", rpcConnectMaxRetries)
	}
}

func (self *Browser) syncState() error {
	if self.devtools != nil {
		if targets, err := self.devtools.List(self.ctx()); err == nil {
			var ids []string

			for _, target := range targets {
				if target.Type == devtool.Page {
					if tab, err := newTabFromTarget(self, target); err == nil {
						self.tabLock.Lock()
						self.tabs[target.ID] = tab
						ids = append(ids, tab.ID())

						if self.activeTabId == `` {
							log.Debugf("Setting tab %v as active", tab.ID())
							self.activeTabId = tab.ID()
						}

						self.tabLock.Unlock()
					} else {
						log.Warningf("failed to register tab %v: %v", target.ID, err)
					}
				}
			}

			self.tabLock.Lock()
			defer self.tabLock.Unlock()

			// cull tabs on our end that no longer exist in Chrome
			for id, tab := range self.tabs {
				if !sliceutil.ContainsString(ids, id) {
					if err := tab.Disconnect(); err != nil {
						log.Warningf("failed to disconnect tab %v: %v", tab.ID(), err)
					}

					delete(self.tabs, id)
				}
			}

		} else {
			return fmt.Errorf("DevTools error: %v", err)
		}
	} else {
		return fmt.Errorf("DevTools connection unavailable")
	}

	return nil
}

func fnOutputVarToError(in reflect.Value) error {
	if !in.IsValid() {
		return fmt.Errorf("invalid output value")
	}

	if !in.CanInterface() {
		return fmt.Errorf("cannot get output value")
	}

	if rv := in.Interface(); rv == nil {
		return nil
	} else {
		if err, ok := rv.(error); ok {
			return err
		} else {
			return fmt.Errorf("Expected return type to be an error, got %T", rv)
		}
	}
}
