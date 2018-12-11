package core

import (
	"fmt"
	"net/url"
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/utils"
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
	Referrer string `json:"referrer"`

	// Whether to block until the page has finished loading.
	WaitForLoad bool `json:"wait_for_load" default:"true"`

	// The amount of time to wait for the page to load.
	Timeout time.Duration `json:"timeout" default:"30s"`

	// The amount of time to poll for the originating network request.
	RequestPollTimeout time.Duration `json:"request_poll_timeout" default:"5s"`

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

	// Whether to continue execution if load_event_name is not seen before
	// timeout elapses.
	ContinueOnTimeout bool `json:"continue_on_timeout" default:"false"`

	// The RPC event to wait for before proceeding to the next command.
	LoadEventName string `json:"load_event_name" default:"Page.loadEventFired"`

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
// go "https://www.exmaple.com" {
//   timeout:             '10s',
//   continue_on_timeout: true,
//   load_event_name:     'Network.responseReceived',
// }
// ```
//
func (self *Commands) Go(uri string, args *GoArgs) (*GoResponse, error) {
	if args == nil {
		args = &GoArgs{}
	}

	defaults.SetDefaults(args)

	args.Timeout = utils.FudgeDuration(args.Timeout)
	args.RequestPollTimeout = utils.FudgeDuration(args.RequestPollTimeout)

	// if specified as random, generate a referrer with a UUID in the url
	switch args.Referrer {
	case `random`:
		args.Referrer = stringutil.UUID().String()
	case ``:
		args.Referrer = RandomReferrerPrefix
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

			// if a scheme wasn't given, prepend HTTPS
			if u.Scheme == `` {
				u.Scheme = `https`
			}

			// if basic auth credentials are specified, setup the request intercept to provide them
			username := args.Username
			password := args.Password

			if username != `` || password != `` {
				if err := self.browser.Tab().AddNetworkIntercept(``, true, func(tab *browser.Tab, pattern *browser.NetworkRequestPattern, event *browser.Event) *browser.NetworkInterceptResponse {
					response := &browser.NetworkInterceptResponse{}

					if event.P().Bool(`isNavigationRequest`) {
						if origin := event.P().String(`authChallenge.origin`); origin != `` {
							if args.Realm == `` || args.Realm == event.P().String(`authChallenge.realm`) {
								u := args.Username
								p := args.Password

								if u == `` && p == `` {
									response.AuthResponse = `Cancel`
								} else {
									response.AuthResponse = `ProvideCredentials`
									response.Username = username
									response.Password = password
								}
							}
						}
					}

					return response
				}); err != nil {
					log.Warning(err)
					return nil, fmt.Errorf("Failed to setup authentication intercept")
				}
			}

			if rv, err := self.browser.Tab().Navigate(u.String()); err == nil {
				if args.WaitForLoad && args.Timeout > 0 {
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
				} else {
					log.Debugf("core::go not waiting for navigation: WaitForLoad=%v Timeout=%v", args.WaitForLoad, args.Timeout)
				}

				totalTime = time.Since(commandIssued)
				rvM := maputil.M(rv.Result)
				netPollStart := time.Now()

				var netreq *browser.NetworkRequest

				// poll aggressively waiting to receive the network request that
				// loaded the page
				for time.Since(netPollStart) < args.RequestPollTimeout {
					// locate the network request, response/error that resulted
					// from the page navigation call
					if req := self.browser.Tab().GetLoaderRequest(
						rvM.String(`loaderId`, rvM.String(`frameId`)),
					); req != nil && req.IsCompleted() {
						netreq = req
						break
					}

					time.Sleep(33 * time.Millisecond)
				}

				if netreq != nil {
					if err := netreq.Error(); err == nil {
						cmdresp := &GoResponse{}

						if v := netreq.R().Int(`response.status`); v >= 0 {
							if v >= 400 && !args.ContinueOnError {
								// ContinueStatuses (if set) gives us one last chance to accept this response before erroring out
								if len(args.ContinueStatuses) == 0 || !sliceutil.Contains(args.ContinueStatuses, v) {
									return nil, fmt.Errorf("HTTP %v", v)
								}
							}

							cmdresp.Status = int(v)
							cmdresp.URL = netreq.R().String(`response.url`)
							cmdresp.MimeType = netreq.R().String(`response.mimeType`)
							cmdresp.Protocol = netreq.R().String(`response.protocol`)
							cmdresp.RemoteAddress = fmt.Sprintf(
								"%v:%v",
								netreq.R().String(`response.remoteIPAddress`),
								netreq.R().Int(`response.remotePort`, 80),
							)
							cmdresp.TimingDetails = make(map[string]float64)
							cmdresp.Headers = make(map[string]string)

							// build timing
							for key, value := range netreq.R().Map(`response.timing`) {
								cmdresp.TimingDetails[key.String()] = value.Float()
							}

							cmdresp.TimingDetails[`overallTimeMs`] = float64(totalTime.Nanoseconds()) / float64(1e6)

							// build headers
							for key, value := range netreq.R().Map(`response.headers`) {
								cmdresp.Headers[key.String()] = value.String()
							}

							log.Debugf("Page loaded in %v: HTTP %d: %v", totalTime, cmdresp.Status, cmdresp.URL)

							return cmdresp, nil
						} else {
							return nil, fmt.Errorf("Got invalid HTTP status")
						}
					} else {
						return &GoResponse{}, fmt.Errorf("Request failed: %v", err)
					}
				} else if !args.RequireOriginatingRequest {
					return &GoResponse{}, nil
				} else {
					return nil, fmt.Errorf("Failed to locate originating network request")
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
