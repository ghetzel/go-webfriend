package core

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/utils"
	defaults "github.com/mcuadros/go-defaults"
)

var RandomReferrerPrefix = `https://github.com/ghetzel/go-webfriend`

type GoArgs struct {
	// If a URL is specified, it will be used as the HTTP Referer [sic] header
	// field when going to the given page. If the URL of the currently-loaded
	// page and the referrer are the same, the page will no change.
	//
	// For this reason, you may specify the special value 'random', which will
	// generate a URL with a randomly generated path component to ensure that
	// it is always different from the current page. Specifying None will omit
	// the field from the request.
	Referrer string `json:"referrer" default:"random"`

	// Whether to block until the page has finished loading.
	WaitForLoad bool `json:"wait_for_load" default:"true"`

	// The amount of time, in milliseconds, to wait for the the to load.
	Timeout time.Duration `json:"timeout" default:"30s"`

	// Whether the resources stack that is queried in page::resources and
	// page::resource is cleared before navigating. Set this to false to
	// preserve the ability to retrieve data that was loaded on previous pages.
	ClearRequests bool `json:"clear_requests" default:"true"`

	// Whether to continue execution if an error is encountered during page
	// load (e.g.: HTTP 4xx/5xx, SSL, TCP connection errors).
	ContinueOnError bool `json:"continue_on_error"`

	// Whether to continue execution if load_event_name is not seen before
	// timeout elapses.
	ContinueOnTimeout bool `json:"continue_on_timeout" default:"false"`

	// The RPC event to wait for before proceeding to the next command.
	LoadEventName string `json:"load_event_name" default:"Page.loadEventFired"`
}

type GoResponse struct {
	URL           string             `json:"url"`
	Status        int                `json:"status"`
	TimingDetails map[string]float64 `json:"timing"`
	Headers       map[string]string  `json:"headers"`
	MimeType      string             `json:"mimetype"`
	RemoteAddress string             `json:"remoteAddress"`
	Protocol      string             `json:"protocol"`
}

// Nagivate to a URL.
func (self *Commands) Go(uri string, args *GoArgs) (*GoResponse, error) {
	if args == nil {
		args = &GoArgs{}
	}

	defaults.SetDefaults(args)

	if args.Timeout == 0 {
		args.Timeout = time.Duration(30) * time.Second
	}

	// if specified as random, generate a referrer with a UUID in the url
	if args.Referrer == `random` {
		args.Referrer = fmt.Sprintf(
			"%s/%v",
			strings.TrimSuffix(RandomReferrerPrefix, `/`),
			stringutil.UUID(),
		)
	}

	// clear our network requests accumulated so far
	if args.ClearRequests {
		self.browser.Tab().ResetNetworkRequests()
	}

	if u, err := url.Parse(uri); err == nil {
		// register the waiter BEFORE making the Page.navigate call because some pages will load
		// so fast that we get a race condition otherwise
		if waiter, err := self.browser.Tab().CreateEventWaiter(args.LoadEventName); err == nil {
			defer waiter.Remove()
			var commandIssued = time.Now()
			var totalTime time.Duration

			if rv, err := self.browser.Tab().RPC(`Page`, `navigate`, map[string]interface{}{
				`URL`: u.String(),
			}); err == nil {
				if args.WaitForLoad {
					// wait for the first event matching the given pattern
					if event, err := waiter.Wait(args.Timeout); err != nil {
						if utils.IsTimeoutErr(err) {
							if !args.ContinueOnTimeout {
								return nil, fmt.Errorf("timed out waiting for event %s", args.LoadEventName)
							}
						} else {
							return nil, err
						}
					} else {
						log.Debugf("core::go proceeding: got event %v", event.Name)
					}
				}

				totalTime = time.Since(commandIssued)
				rvM := maputil.M(rv)

				// locate the network request, response/error that resulted from the page navigation call
				if req, res, rerr := self.browser.Tab().GetLoaderRequest(
					rvM.String(`loaderId`, rvM.String(`frameId`)),
				); req != nil {
					cmdresp := &GoResponse{}

					if res != nil {
						if v := res.Params.Int(`response.status`); v >= 0 {
							if v >= 400 && !args.ContinueOnError {
								return nil, fmt.Errorf("HTTP %v", v)
							}

							cmdresp.Status = int(v)
							cmdresp.URL = res.Params.String(`response.url`)
							cmdresp.MimeType = res.Params.String(`response.mimeType`)
							cmdresp.Protocol = res.Params.String(`response.protocol`)
							cmdresp.RemoteAddress = fmt.Sprintf(
								"%v:%v",
								res.Params.String(`response.remoteIPAddress`),
								res.Params.Int(`response.remotePort`, 80),
							)
							cmdresp.TimingDetails = make(map[string]float64)
							cmdresp.Headers = make(map[string]string)

							// build timing
							for key, value := range res.Params.Map(`response.timing`) {
								cmdresp.TimingDetails[key.String()] = value.Float()
							}

							cmdresp.TimingDetails[`overallTimeMs`] = float64(totalTime.Nanoseconds()) / float64(1e6)

							// build headers
							for key, value := range res.Params.Map(`response.headers`) {
								cmdresp.Headers[key.String()] = value.String()
							}
						}
					} else if rerr != nil && !args.ContinueOnError {
						return nil, fmt.Errorf("Request error: %v", rerr.Params.String(`errorText`, `Unknown Error`))
					}

					log.Debugf("Page loaded in %v: HTTP %d: %v", totalTime, cmdresp.Status, cmdresp.URL)

					return cmdresp, nil
				} else {
					return nil, fmt.Errorf("Unable to locate originating network request")
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}
}
