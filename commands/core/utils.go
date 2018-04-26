package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
	defaults "github.com/mcuadros/go-defaults"
)

type EnvArgs struct {
	// The value to return if the environment variable does not exist, or
	// (optionally) is empty.
	Fallback interface{} `json:"fallback"`

	// Whether empty values should be ignored or not.
	IgnoreEmpty bool `json:"ignore_empty" default:"true"`

	// Whether automatic type detection should be performed or not.
	DetectType bool `json:"detect_type" default:"true"`

	// If specified, this string will be used to split matching values into a
	// list of values. This is useful for environment variables that contain
	// multiple values joined by a separator (e.g: the PATH variable.)
	Joiner string `json:"joiner"`
}

// Retrieves a system environment variable and returns the value of it, or a
// fallback value if the variable does not exist or (optionally) is empty.
func (self *Commands) Env(name string, args *EnvArgs) (interface{}, error) {
	if args == nil {
		args = &EnvArgs{}
	}

	defaults.SetDefaults(args)

	if ev := os.Getenv(name); ev != `` {
		var rv interface{}

		if args.Joiner != `` {
			rv = strings.Split(ev, args.Joiner)
		}

		// perform type detection
		if args.DetectType {
			// for arrays, autotype each element
			if typeutil.IsArray(rv) {
				rv = sliceutil.Autotype(rv)
			} else {
				rv = stringutil.Autotype(ev)
			}
		} else {
			rv = ev
		}

		return rv, nil
	} else if !args.IgnoreEmpty {
		return nil, fmt.Errorf("Environment variable %q was not specified", name)
	} else {
		return nil, nil
	}
}

// Immediately exit the script in an error-like fashion with a specific message.
func (self *Commands) Fail(message string) error {
	if message == `` {
		message = `Unspecified error`
	}

	return errors.New(message)
}

// Directly call an RPC method with the given parameters.
func (self *Commands) Rpc(method string, args map[string]interface{}) (interface{}, error) {
	mod, meth := stringutil.SplitPair(method, `::`)

	return self.browser.Tab().RPC(mod, meth, args)
}

// Outputs a line to the log.
func (self *Commands) Log(message interface{}) error {
	if typeutil.IsScalar(reflect.ValueOf(message)) {
		fmt.Printf("%v\n", message)
	} else if data, err := json.MarshalIndent(message, ``, `  `); err == nil {
		fmt.Printf(string(data) + "\n")
	} else {
		log.Errorf("Failed to log message: %v", err)
		return err
	}

	return nil
}

// Store a value in the current scope. Strings will be automatically converted
// into the appropriate data types (float, int, bool) if possible.
func (self *Commands) Put(value interface{}) (interface{}, error) {
	return value, nil
}

type RunArgs struct {
	Data          interface{} `json:"data"`           // null
	Isolated      bool        `json:"isolated"`       // true
	PreserveState bool        `json:"preserve_state"` // true
	MergeScopes   bool        `json:"merge_scopes"`   // false
	ResultKey     string      `json:"result_key"`     // result
}

// Evaluates another Friendscript loaded from another file. The filename is the
// absolute path or basename of the file to search for in the WEBFRIEND_PATH
// environment variable to load and evaluate. The WEBFRIEND_PATH variable
// behaves like the the traditional *nix PATH variable, wherein multiple paths
// can be specified as a colon-separated (:) list. The current working directory
// will always be checked first.
//
// Returns: The value of the variable named by result_key at the end of the
// evaluated script's execution.
//
func (self *Commands) Run(filename string, args *RunArgs) (interface{}, error) {
	return nil, fmt.Errorf(`Not Implemented Yet`)
}

// Change the current selector scope to be rooted at the given element. If
// selector is empty, the scope is set to the document element (i.e.: global).
func (self *Commands) SwitchRoot(selector browser.Selector) error {
	return fmt.Errorf(`Not Implemented Yet`)
}

// Highlight the node matching the given selector, or clear all highlights if
// the selector is "none"
func (self *Commands) Highlight(selector browser.Selector) error {
	if selector.IsNone() {
		return self.browser.Tab().AsyncRPC(`DOM`, `hideHighlight`, nil)
	} else {
		docroot := self.browser.Tab().DOM()

		if elements, err := docroot.Query(selector, nil); err == nil || browser.IsElementNotFoundErr(err) {
			for _, element := range elements {
				if err := element.Highlight(); err != nil {
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
				if err := element.Highlight(); err != nil {
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
