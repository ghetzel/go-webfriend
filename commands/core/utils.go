package core

import (
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/pathutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/jdxcode/netrc"
)

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
		args = new(NetrcArgs)
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
