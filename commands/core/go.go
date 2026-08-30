package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/playwright-community/playwright-go"
	defaults "go.gary.cool/go-defaults"
	"go.gary.cool/go-stockutil/stringutil"
	"go.gary.cool/go-webfriend/browser"
	"go.gary.cool/go-webfriend/utils"
)

var RandomReferrerPrefix = `https://go.gary.cool/go-webfriend`

type GoArgs struct {
	// If a URL is specified, it will be used as the HTTP Referer [sic] header
	// field when going to the given page. If the URL of the currently-loaded
	// page and the referrer are the same, the page will no change.
	//
	// For this reason, you may specify the special value 'random', which will
	// generate a URL with a randomly generated path component to ensure that
	// it is always different from the current page. Specifying None will omit
	// the field from the request.
	Referrer string `json:"referrer"`

	// Whether to block until the page has finished loading.
	WaitForLoad bool `json:"wait_for_load" default:"true"`

	// The amount of time to wait for the page to load.
	Timeout time.Duration `json:"timeout" default:"30s"`

	// Whether the resources stack that is queried in page::resources and
	// page::resource is cleared before navigating. Set this to false to
	// preserve the ability to retrieve data that was loaded on previous pages.
	ClearRequests bool `json:"clear_requests" default:"false"`

	// Whether the originating network request is required in the return value.  If this is
	// false, the response may be missing status, headers, and timing details.
	RequireOriginatingRequest bool `json:"require_originating_request" default:"true"`

	// Whether to continue execution if an error is encountered during page
	// load (e.g.: HTTP 4xx/5xx, SSL, TCP connection errors).
	ContinueOnError bool `json:"continue_on_error"`

	// These HTTP status codes are not considered errors.
	ContinueStatuses []int `json:"continue_statuses"`

	// The RPC event to wait for before proceeding to the next command.
	LoadEventName string `json:"load_event_name" default:"load"`

	// Provide a username if one is requested via HTTP Basic authentication.
	Username string `json:"username"`

	// Provide a password if one is requested via HTTP Basic authentication.
	Password string `json:"password"`

	// Only provide credentials if the HTTP Basic Authentication Realm matches this one.
	Realm string `json:"realm"`
}

type GoResponse struct {
	// The final URL of the page that was loaded.
	URL string `json:"url"`

	// The HTTP status code of the loaded page.
	Status int `json:"status"`

	// A map of durations (in milliseconds) that various phases of the page load took.
	TimingDetails map[string]float64 `json:"timing"`

	// Map of HTTP response headers.
	Headers map[string]string `json:"headers"`

	// The MIME type of the response content.
	MimeType string `json:"mimetype"`

	// The remote address of the loaded page.
	RemoteAddress string `json:"remoteAddress"`

	// The protocol that was negotiated and used to load the page.
	Protocol string `json:"protocol"`
}

// Navigate to a URL.
//
// #### Examples
//
// ##### Go to Google.
// ```
// go "google.com"
// ```
//
// ##### Go to www.example.com, only wait for the first network response, and don't fail if the request times out.
// ```
//
//	go "https://www.example.com" {
//	  timeout:             '10s',
//	  continue_on_timeout: true,
//	  load_event_name:     'Network.responseReceived',
//	}
//
// ```
func (self *Commands) Go(uri string, args *GoArgs) (*GoResponse, error) {
	if args == nil {
		args = &GoArgs{}
	}

	defaults.SetDefaults(args)

	args.Timeout = utils.FudgeDuration(args.Timeout)

	// if specified as random, generate a referrer with a UUID in the url
	switch args.Referrer {
	case `random`:
		args.Referrer = stringutil.UUID().String()
	case ``:
		args.Referrer = RandomReferrerPrefix
	}

	if pg := self.browser.Page(); pg != nil {
		var pwWaitState *playwright.WaitUntilState

		if args.WaitForLoad {
			switch args.LoadEventName {
			case ``, `load`:
				pwWaitState = playwright.WaitUntilStateLoad
			case `content`:
				pwWaitState = playwright.WaitUntilStateDomcontentloaded
			case `idle`:
				pwWaitState = playwright.WaitUntilStateNetworkidle
			case `commit`:
				pwWaitState = playwright.WaitUntilStateCommit
			default:
				return nil, fmt.Errorf("invalid load event name %q", args.LoadEventName)
			}
		}

		if res, err := pg.Goto(uri, playwright.PageGotoOptions{
			Referer:   playwright.String(args.Referrer),
			WaitUntil: pwWaitState,
			Timeout:   playwright.Float(float64(args.Timeout.Milliseconds())),
		}); err == nil {
			res.Finished()

			var fsres = &GoResponse{
				URL:    res.URL(),
				Status: res.Status(),
			}

			if timings := res.Request().Timing(); timings != nil {
				fsres.TimingDetails = make(map[string]float64)

				fsres.TimingDetails[`startTime`] = timings.StartTime
				fsres.TimingDetails[`domainLookupStart`] = timings.DomainLookupStart
				fsres.TimingDetails[`domainLookupEnd`] = timings.DomainLookupEnd
				fsres.TimingDetails[`connectStart`] = timings.ConnectStart
				fsres.TimingDetails[`secureConnectionStart`] = timings.SecureConnectionStart
				fsres.TimingDetails[`connectEnd`] = timings.ConnectEnd
				fsres.TimingDetails[`requestStart`] = timings.RequestStart
				fsres.TimingDetails[`responseStart`] = timings.ResponseStart
				fsres.TimingDetails[`responseEnd`] = timings.ResponseEnd
			}

			if addr, err := res.ServerAddr(); err == nil {
				fsres.RemoteAddress = fmt.Sprintf("%v:%d", addr.IpAddress, addr.Port)
			}

			if hdr, err := res.AllHeaders(); err == nil {
				fsres.Headers = hdr
			}

			if !res.Ok() {
				if !args.ContinueOnError {
					return fsres, fmt.Errorf("HTTP %d: %s", res.Status(), res.StatusText())
				} else if !slices.Contains(args.ContinueStatuses, fsres.Status) {
					return fsres, fmt.Errorf("HTTP %d: %s", res.Status(), res.StatusText())
				}
			}

			return fsres, nil
		} else {
			return nil, err
		}
	} else {
		return nil, browser.NoActivePage
	}
}
