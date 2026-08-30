package browser

import (
	"net/url"
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	"go.gary.cool/friendscript/utils"
	"go.gary.cool/go-stockutil/executil"
	"go.gary.cool/go-stockutil/log"
	"go.gary.cool/go-stockutil/sliceutil"
	"go.gary.cool/go-stockutil/stringutil"
	"go.gary.cool/go-stockutil/typeutil"
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
	Accuracy        float64
	AuthOrigin      string
	AuthPassword    string
	AuthUsername    string
	BaseURL         string
	DisableCSP      bool
	DisableNetwork  bool
	DisableScripts  bool
	EmulateTouch    bool
	Engine          string
	Height          int
	IgnoreTLSErrors bool
	Latitude        float64
	Locale          string
	Longitude       float64
	ProxyBypass     []string
	ProxyPassword   string
	ProxyURL        string
	ProxyUsername   string
	Scale           float64
	StartURL        string
	Timezone        string
	UserAgent       string
	Width           int
	playwright      *playwright.Playwright
	browser         playwright.Browser
	pages           []*Page
	activePage      int
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
	if pw, err := playwright.Run(&playwright.RunOptions{}); err == nil {
		self.playwright = pw
	} else {
		return err
	}

	var lerr error

	if opts, err := self.browserLaunchOptions(); err == nil {
		switch self.Engine {
		case ``, `chromium`:
			self.browser, lerr = self.playwright.Chromium.Launch(*opts)
		case `firefox`:
			self.browser, lerr = self.playwright.Firefox.Launch(*opts)
		case `webkit`:
			self.browser, lerr = self.playwright.WebKit.Launch(*opts)
		}
	} else {
		return errors.Wrap(err, "bad options")
	}

	if lerr != nil {
		return errors.Wrap(lerr, "failed launch")
	}

	if self.StartURL == `` {
		self.StartURL = DefaultStartURL
	}

	if p, err := NewPage(self); err == nil {
		if self.StartURL != `` {
			if _, err := p.Goto(self.StartURL, playwright.PageGotoOptions{
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

func (self *Browser) browserLaunchOptions() (*playwright.BrowserTypeLaunchOptions, error) {
	var opts = new(playwright.BrowserTypeLaunchOptions)

	// webfriend will handle the signals itself
	opts.HandleSIGHUP = playwright.Bool(false)
	opts.HandleSIGINT = playwright.Bool(false)
	opts.HandleSIGTERM = playwright.Bool(false)

	return opts, nil
}

func (self *Browser) newPageOptions() []playwright.BrowserNewPageOptions {
	var opt playwright.BrowserNewPageOptions

	// simple options
	opt.BaseURL = playwright.String(self.BaseURL)
	opt.BypassCSP = playwright.Bool(self.DisableCSP)
	opt.HasTouch = playwright.Bool(self.EmulateTouch)
	opt.IgnoreHttpsErrors = playwright.Bool(self.IgnoreTLSErrors)
	opt.JavaScriptEnabled = playwright.Bool(!self.DisableScripts)
	opt.Offline = playwright.Bool(self.DisableNetwork)

	// scale
	if s := self.Scale; s > 0 {
		opt.DeviceScaleFactor = playwright.Float(s)
	}

	// locale
	if locale := self.Locale; locale != `` {
		opt.Locale = playwright.String(locale)
	}

	// geolocation. my apologies to Null Island
	if lat := self.Latitude; lat != 0 {
		if lon := self.Longitude; lon != 0 {
			opt.Geolocation = &playwright.Geolocation{
				Latitude:  lat,
				Longitude: lon,
				Accuracy:  playwright.Float(self.Accuracy),
			}
		}
	}

	// viewport size
	if w := self.Width; w > 0 {
		if h := self.Height; h > 0 {
			opt.Screen = &playwright.Size{
				Width:  w,
				Height: h,
			}

			opt.Viewport = &playwright.Size{
				Width:  w,
				Height: h,
			}
		}
	}

	// user agent
	switch ua := self.UserAgent; ua {
	case ``:
		break
	case `random`:
		opt.UserAgent = playwright.String(stringutil.UUID().Base58())
	default:
		opt.UserAgent = playwright.String(ua)
	}

	// timezone
	if tz := self.Timezone; tz != `` {
		opt.TimezoneId = playwright.String(tz)
	} else if tz := os.Getenv(`TZ`); tz != `` {
		opt.TimezoneId = playwright.String(tz)
	}

	// credentials
	if self.AuthUsername != `` {
		var creds = &playwright.HttpCredentials{
			Username: self.AuthUsername,
			Password: self.AuthPassword,
		}

		if origin := self.AuthOrigin; origin != `` {
			creds.Origin = playwright.String(origin)
		}

		opt.HttpCredentials = creds
	}

	// proxy server
	var proxyURL *url.URL

	if pu := self.ProxyURL; pu != `` {
		proxyURL, _ = url.Parse(pu)
	} else if pu := os.Getenv(`HTTP_PROXY`); pu != `` {
		proxyURL, _ = url.Parse(pu)
	}

	if proxyURL != nil {
		var uu string
		var up string

		if user := proxyURL.User; user != nil {
			uu = user.Username()
			up, _ = user.Password()
		}

		opt.Proxy = &playwright.Proxy{
			Server: proxyURL.String(),
		}

		if np := os.Getenv(`NO_PROXY`); np != `` {
			self.ProxyBypass = sliceutil.SplitTrimSpaceCompact(np, `,`)
		}

		if bypass := self.ProxyBypass; len(bypass) > 0 {
			opt.Proxy.Bypass = playwright.String(strings.Join(bypass, `, `))
		}

		if u := typeutil.OrString(self.ProxyUsername, uu); u != `` {
			opt.Proxy.Username = playwright.String(u)
		}

		if p := typeutil.OrString(self.ProxyPassword, up); p != `` {
			opt.Proxy.Password = playwright.String(p)
		}
	}

	return []playwright.BrowserNewPageOptions{opt}
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

	self.browser = nil
	self.playwright = nil
	self.pages = nil
	self.activePage = 0

	return
}
