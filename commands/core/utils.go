package core

import (
	"fmt"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/pathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/jdxcode/netrc"
)

// [SKIP]
// Directly call an RPC method with the given parameters.
func (self *Commands) Rpc(method string, args map[string]interface{}) (interface{}, error) {
	mod, meth := stringutil.SplitPair(method, `::`)

	return self.browser.Tab().RPC(mod, meth, args)
}

// [SKIP]
// Change the current selector scope to be rooted at the given element. If
// selector is empty, the scope is set to the document element (i.e.: global).
func (self *Commands) SwitchRoot(selector browser.Selector) error {
	return fmt.Errorf(`Not Implemented Yet`)
}

type HighlightArgs struct {
	// The red component of the highlight color (0 <= r < 256)
	R int `json:"r" default:"0"`

	// The green component of the highlight color (0 <= g < 256)
	G int `json:"g" default:"128"`

	// The blue component of the highlight color (0 <= b < 256)
	B int `json:"b" default:"128"`

	// The alpha component of the highlight color (0.0 <= a <= 1.0)
	A float64 `json:"a" default:"0.5"`
}

// Highlight the node matching the given selector, or clear all highlights if
// the selector is "none"
func (self *Commands) Highlight(selector interface{}, args *HighlightArgs) error {
	if args == nil {
		args = &HighlightArgs{}
	}

	defaults.SetDefaults(args)

	if elements, isNone, err := self.browser.ElementsFromSelector(selector); err == nil {
		if isNone {
			return self.browser.Tab().AsyncRPC(`DOM`, `hideHighlight`, nil)
		} else {
			for _, element := range elements {
				if err := element.Highlight(args.R, args.G, args.B, args.A); err != nil {
					return err
				}
			}

			return nil
		}
	} else {
		return err
	}
}

type InspectArgs struct {
	// The X-coordinate to inspect.
	X float64 `json:"x"`

	// The Y-coordinate to inspect.
	Y float64 `json:"y"`

	// Whether to highlight the inspected DOM element or not.
	Highlight bool `json:"highlight" default:"true"`

	// The red component of the highlight color (0 <= r < 256)
	R int `json:"r" default:"0"`

	// The green component of the highlight color (0 <= g < 256)
	G int `json:"g" default:"128"`

	// The blue component of the highlight color (0 <= b < 256)
	B int `json:"b" default:"128"`

	// The alpha component of the highlight color (0.0 <= a <= 1.0)
	A float64 `json:"a" default:"0.5"`
}

// Retrieve the element at the given coordinates, optionally highlighting it.
func (self *Commands) Inspect(args *InspectArgs) (*browser.Element, error) {
	if args == nil {
		args = &InspectArgs{}
	}

	defaults.SetDefaults(args)

	if rv, err := self.browser.Tab().RPC(`DOM`, `getNodeForLocation`, map[string]interface{}{
		`x`: int(args.X),
		`y`: int(args.Y),
	}); err == nil {
		if element, ok := self.browser.Tab().DOM().Element(int(rv.R().Int(`backendNodeId`))); ok {
			if args.Highlight {
				if err := element.Highlight(args.R, args.G, args.B, args.A); err != nil {
					return nil, err
				}
			}

			return element, nil
		} else {
			return nil, fmt.Errorf("No element was found at the given coordinates.")
		}
	} else {
		return nil, err
	}
}

// Immediately close the browser without error or delay.
func (self *Commands) Exit() error {
	return browser.ExitRequested
}

type NetrcArgs struct {
	// The path to the .netrc file to load values from.
	Filename string `json:"filename" default:"~/.netrc"`

	// A list of additional, non-standard fields to retrieve from the .netrc entry
	ExtraFields []string `json:"extra_fields"`
}

type NetrcResponse struct {
	// Whether there was a match or not.
	OK bool `json:"ok"`

	// The machine name that matched.
	Machine string `json:"machine"`

	// The login name.
	Login string `json:"login"`

	// The password.
	Password string `json:"password"`

	// Any additional values retrieved from the entry
	Fields map[string]string
}

// Retrieve a username and password from a .netrc-formatted file.
func (self *Commands) Netrc(machine string, args *NetrcArgs) (*NetrcResponse, error) {
	if args == nil {
		args = &NetrcArgs{}
	}

	defaults.SetDefaults(args)

	if expanded, err := pathutil.ExpandUser(args.Filename); err == nil {
		if nrc, err := netrc.Parse(expanded); err == nil {
			response := &NetrcResponse{
				Machine: machine,
			}

			// retrieve and populate the machine-specific entry (if present)
			if entry := nrc.Machine(machine); entry != nil {
				response.OK = true
				response.Login = entry.Get(`login`)
				response.Password = entry.Get(`password`)

				if len(args.ExtraFields) > 0 {
					response.Fields = make(map[string]string)

					for _, field := range args.ExtraFields {
						if field == `login` || field == `password` {
							continue
						} else {
							response.Fields[field] = entry.Get(field)
						}
					}
				}
			}

			return response, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
