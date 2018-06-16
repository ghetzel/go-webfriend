// Suite of string formatting utilities.
package fmt

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/utils"
	defaults "github.com/mcuadros/go-defaults"
)

type Commands struct {
	browser   *browser.Browser
	scopeable utils.Scopeable
}

func New(browser *browser.Browser, scopeable utils.Scopeable) *Commands {
	return &Commands{
		browser:   browser,
		scopeable: scopeable,
	}
}

func (self *Commands) ExecuteCommand(name string, arg interface{}, objargs map[string]interface{}) (interface{}, error) {
	return utils.CallCommandFunction(self, stringutil.Camelize(name), arg, objargs)
}

// Takes an input value and returns that value as the most appropriate data type based on its contents.
func (self *Commands) Autotype(input interface{}) (interface{}, error) {
	return typeutil.V(input).Auto(), nil
}

type JoinArgs struct {
	Joiner string `json:"joiner" default:","`
}

// Join an array of inputs into a single string, with each item separated by a given joiner string.
func (self *Commands) Join(inputs interface{}, args *JoinArgs) (string, error) {
	if args == nil {
		args = &JoinArgs{}
	}

	defaults.SetDefaults(args)

	return strings.Join(sliceutil.Stringify(inputs), args.Joiner), nil
}

// Return the given string converted to camelCase.
func (self *Commands) Camelize(input interface{}) (string, error) {
	out := stringutil.Camelize(typeutil.V(input).String())

	if len(out) > 0 {
		old := out
		out = strings.ToLower(string(old[0]))

		if len(old) > 1 {
			out += old[1:]
		}
	}

	return out, nil
}

// Return the given string converted to PascalCase.
func (self *Commands) Pascalize(input interface{}) (string, error) {
	out := stringutil.Camelize(typeutil.V(input).String())
	out = strings.Title(out)

	return out, nil
}

// Return the given string converted to lowercase.
func (self *Commands) Lower(input interface{}) (string, error) {
	return strings.ToLower(typeutil.V(input).String()), nil
}

type ReplaceArgs struct {
	Find    interface{} `json:"find"`
	Replace string      `json:"replace"`
	Count   int         `json:"count" default:"-1"`
}

// Replaces values in an input string (exact matches or regular expressions) with a replacement value.
// Exact matches will be replaced up to a certain number of times, or all occurrences of count is -1 (default).
func (self *Commands) Replace(input interface{}, args *ReplaceArgs) (string, error) {
	if args == nil {
		args = &ReplaceArgs{}
	}

	defaults.SetDefaults(args)

	in := typeutil.V(input).String()

	if typeutil.IsZero(args.Find) {
		return in, nil
	} else if rx, ok := args.Find.(*regexp.Regexp); ok {
		return string(rx.ReplaceAll([]byte(in), []byte(args.Replace))), nil
	} else {
		find := typeutil.V(args.Find).String()

		if stringutil.IsSurroundedBy(find, `/`, `/`) {
			find = stringutil.Unwrap(find, `/`, `/`)

			if rx, err := regexp.Compile(find); err == nil {
				return string(rx.ReplaceAll([]byte(in), []byte(args.Replace))), nil
			} else {
				return ``, fmt.Errorf("Invalid regular expression: %v", err)
			}
		} else {
			return strings.Replace(in, find, args.Replace, args.Count), nil
		}
	}
}

type SplitArgs struct {
	On string `json:"on" default:","`
}

// Split a given string by a given delimiter.
func (self *Commands) Split(input interface{}, args *SplitArgs) ([]string, error) {
	if args == nil {
		args = &SplitArgs{}
	}

	defaults.SetDefaults(args)

	return strings.Split(
		typeutil.V(input).String(),
		args.On,
	), nil
}

// Strip leading and trailing whitespace from the given string.
func (self *Commands) Strip(input interface{}) (string, error) {
	return strings.TrimSpace(typeutil.V(input).String()), nil
}

// Return the given string converted to Title Case.
func (self *Commands) Title(input interface{}) (string, error) {
	return strings.Title(typeutil.V(input).String()), nil
}

// Return the given string converted to underscore_case.
func (self *Commands) Underscore(input interface{}) (string, error) {
	return stringutil.Underscore(typeutil.V(input).String()), nil
}

// Return the given string converted to UPPERCASE.
func (self *Commands) Upper(input interface{}) (string, error) {
	return strings.ToUpper(typeutil.V(input).String()), nil
}

type TrimArgs struct {
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

// Remove a leading and/org trailing string value from the given string.
func (self *Commands) Trim(input interface{}, args *TrimArgs) (string, error) {
	if args == nil {
		args = &TrimArgs{}
	}

	defaults.SetDefaults(args)

	in := typeutil.V(input).String()

	if args.Prefix != `` {
		in = strings.TrimPrefix(in, args.Prefix)
	}

	if args.Suffix != `` {
		in = strings.TrimSuffix(in, args.Suffix)
	}

	return in, nil
}

// Returns the longest common prefix among an array of input strings.
func (self *Commands) Lcp(inputs interface{}) (string, error) {
	return stringutil.LongestCommonPrefix(sliceutil.Stringify(inputs)), nil
}

type FormatArgs struct {
	Data interface{} `json:"data"`
}

// Format the given string according to the given pattern and values.
func (self *Commands) Format(pattern string, args *FormatArgs) (string, error) {
	if args == nil {
		args = &FormatArgs{}
	}

	defaults.SetDefaults(args)

	if typeutil.IsZero(args.Data) {
		return ``, nil
	}

	return fmt.Sprintf(pattern, sliceutil.Sliceify(args.Data)...), nil
}
