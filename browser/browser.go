package browser

import (
	"os"
	"sync"
	"syscall"

	"github.com/ghetzel/friendscript/utils"
	"github.com/ghetzel/go-stockutil/executil"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
)

var DefaultStartURL = `about:blank`
var DebuggerInnerPort = 9222
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

	activeBrowserInstances.Range(func(id any, b any) bool {
		if browser, ok := b.(*Browser); ok {
			browser.Stop()
		}

		return true
	})

	log.Debugf("[browser] Cleanup complete. Time to die.")
}

type Browser struct {
	utils.Runtime
	Engine     string
	URL        string
	playwright *playwright.Playwright
	browser    playwright.Browser
	pages      []*Page
	activePage int
}

func NewBrowser() *Browser {
	return &Browser{}
}

func Start() (*Browser, error) {
	var browser = NewBrowser()
	return browser, browser.Launch()
}

func (self *Browser) SetScope(fsenv utils.Runtime) {
	self.Runtime = fsenv
}

func (self *Browser) Launch() error {
	if pw, err := playwright.Run(); err == nil {
		self.playwright = pw
	} else {
		return err
	}

	var err error

	switch self.Engine {
	case ``, `chromium`:
		self.browser, err = self.playwright.Chromium.Launch()
	case `firefox`:
		self.browser, err = self.playwright.Firefox.Launch()
	case `webkit`:
		self.browser, err = self.playwright.WebKit.Launch()
	}

	if err != nil {
		return errors.Wrap(err, "failed launch")
	}

	if self.URL == `` {
		self.URL = DefaultStartURL
	}

	if p, err := NewPage(self); err == nil {
		if self.URL != `` {
			if _, err := p.Goto(self.URL, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
			}); err == nil {
				self.pages = append(self.pages, p)
			} else {
				return errors.Wrap(err, "failed initial navigation")
			}
		}
	} else {
		return errors.Wrap(err, "failed initial navigation")
	}

	return nil
}

func (self *Browser) Page(index ...int) *Page {
	var i int

	if len(index) > 0 {
		i = index[0]
	} else {
		i = self.activePage
	}

	if i < len(self.pages) {
		return self.pages[i]
	}

	return nil
}

func (self *Browser) Stop() (merr error) {
	if b := self.browser; b != nil {
		merr = log.AppendError(merr, b.Close())
	}

	if pw := self.playwright; pw != nil {
		merr = log.AppendError(merr, pw.Stop())
	}

	return
}
