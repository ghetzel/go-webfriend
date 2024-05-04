package page

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/browser"
)

type InterceptArgs struct {
	// Rewrite the response body with the provided string or with the contents of
	// an io.Reader.
	Body interface{} `json:"body"`

	// Read the response body contents from the named file.
	File string `json:"file"`

	// Specify that the interception should wait for response headers to be sent.  Otherwise the
	// request is intercepted prior to making the request.
	WaitForHeaders bool `json:"wait_for_headers"`

	// Should the request be aborted/rejected.
	Reject bool `json:"reject"`

	// Rewrite the request method to this.
	Method string `json:"method"`

	// Rewrite the request URL to this value.
	URL string `json:"url"`

	// Set the request headers.  Not valid is WaitForHeaders is set.
	Headers map[string]interface{} `json:"headers"`

	// Update the POST data to these values.
	PostData map[string]interface{} `json:"post_data"`

	// Only apply to response HTTP status codes in this list.
	Statuses []int `json:"statuses"`

	// Send credentials in response to this realm.  If empty, the provided credentials
	// will be sent to any authentication challenge.
	Realm string `json:"realm"`

	// Username to authenticate with.
	Username string `json:"username"`

	// Password to authenticate with.
	Password string `json:"password"`

	// Specify whether the intercept should persist after the first match.
	Persistent bool `json:"persistent"`
}

// Intercept all requests where the requested URL matches *match*, and modify the request
// according to the provided arguments.
func (self *Commands) Intercept(match string, args *InterceptArgs) error {
	if args == nil {
		args = &InterceptArgs{}
	}

	defaults.SetDefaults(args)

	if filename := args.File; filename != `` {
		if file, err := self.browser.GetReaderForPath(filename); err == nil {
			defer file.Close()

			buf := bytes.NewBuffer(nil)

			if _, err := io.Copy(buf, file); err == nil {
				args.Body = buf
			} else {
				return err
			}
		} else {
			return err
		}
	} else if contents, ok := args.Body.(string); ok {
		args.Body = bytes.NewBufferString(contents)
	} else if reader, ok := args.Body.(io.Reader); ok {
		args.Body = reader
	} else if contents, ok := args.Body.([]byte); ok {
		args.Body = bytes.NewBuffer(contents)
	} else if contents, ok := args.Body.([]uint8); ok {
		args.Body = bytes.NewBuffer([]byte(contents))
	} else {
		return fmt.Errorf("Must specify a filename or reader")
	}

	return self.browser.Tab().AddNetworkIntercept(match, args.WaitForHeaders, func(tab *browser.Tab, pattern *browser.NetworkRequestPattern, event *browser.Event) *browser.NetworkInterceptResponse {
		response := &browser.NetworkInterceptResponse{
			Autoremove: !args.Persistent,
		}

		if reader, ok := args.Body.(io.Reader); ok {
			log.Debugf("Setting request body override")
			response.Body = reader
		}

		if status := event.P().Int(`responseStatusCode`); len(args.Statuses) == 0 || sliceutil.Contains(args.Statuses, status) {
			if args.Reject {
				response.Error = errors.New(`Aborted`)
			}

			if method := args.Method; method != `` {
				response.Method = method
			}

			if url := args.URL; url != `` {
				response.URL = url
			}

			if hdr := args.Headers; len(hdr) > 0 {
				response.Header = make(http.Header)

				for k, v := range hdr {
					response.Header.Set(k, stringutil.MustString(v))
				}
			}

			if data := args.PostData; len(data) > 0 {
				response.PostData = data
			}
		}

		return response
	})
}

// Clear all request intercepts.
func (self *Commands) ClearIntercepts() error {
	return self.browser.Tab().ClearNetworkIntercepts()
}
