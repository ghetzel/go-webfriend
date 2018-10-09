package core

import (
	"fmt"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/browser"
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
func (self *Commands) Highlight(selector browser.Selector, args *HighlightArgs) error {
	if args == nil {
		args = &HighlightArgs{}
	}

	defaults.SetDefaults(args)

	if selector.IsNone() {
		return self.browser.Tab().AsyncRPC(`DOM`, `hideHighlight`, nil)
	} else {
		docroot := self.browser.Tab().DOM()

		if elements, err := docroot.Query(selector, nil); err == nil || browser.IsElementNotFoundErr(err) {
			for _, element := range elements {
				if err := element.Highlight(args.R, args.G, args.B, args.A); err != nil {
					return err
				}
			}

			return nil
		} else {
			return err
		}
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
		if element, ok := self.browser.Tab().DOM().Element(int(rv.R().Int(`nodeId`))); ok {
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
