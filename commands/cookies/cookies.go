// Commands for interacting with the browser's cookie storage backend.
package cookies

import (
	"fmt"
	"time"

	"github.com/ghetzel/friendscript"
	"github.com/ghetzel/friendscript/utils"
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-webfriend/browser"
)

type Cookie struct {
	// The name of the cookie.
	Name string `json:"name"`

	// The cookie's value.
	Value string `json:"value"`

	// The applicable domain for the cookie.
	Domain string `json:"domain"`

	// The cookie's path.
	Path string `json:"path"`

	// The size of the cookie.
	Size int `json:"size"`

	// The cookie's expiration date.
	Expires *time.Time `json:"expires"`

	// The cookie is flagged as being inaccessible to client-side scripts.
	HttpOnly bool `json:"http_only"`

	// The cookie is flagged as "secure"
	Secure bool `json:"secure"`

	// This is a session cookie.
	Session bool `json:"session"`

	// The same site value of the cookie ("Strict" or "Lax")
	SameSite string `json:"same_site"`

	// The URL the cookie should refer to (when setting)
	URL string `json:"url,omitempty"`
}

func (self *Cookie) native() map[string]interface{} {
	params := make(map[string]interface{})

	params[`name`] = self.Name
	params[`value`] = self.Value

	if v := self.URL; v != `` {
		params[`url`] = v
	}

	if v := self.Domain; v != `` {
		params[`domain`] = v
	}

	if v := self.Path; v != `` {
		params[`path`] = v
	}

	if self.Secure {
		params[`secure`] = self.Secure
	}

	if self.HttpOnly {
		params[`httpOnly`] = self.HttpOnly
	}

	if v := self.SameSite; v != `` {
		params[`sameSite`] = v
	}

	if tm := self.Expires; tm != nil {
		params[`expires`] = int(tm.Unix())
	}

	return params
}

type Commands struct {
	friendscript.Module
	browser *browser.Browser
}

func New(browser *browser.Browser, scopeable utils.Scopeable) *Commands {
	cmd := &Commands{}

	cmd.browser = browser
	cmd.Module = friendscript.CreateModule(cmd)

	return cmd
}

type ListArgs struct {
	// an array of strings representing the URLs to retrieve cookies for.  If omitted, the
	// URL of the current browser tab will be used
	URLs []string `json:"urls"`
}

// List all cookies, either for the given set of URLs or for the current tab (if omitted).
func (self *Commands) List(args *ListArgs) ([]*Cookie, error) {
	if args == nil {
		args = &ListArgs{}
	}

	defaults.SetDefaults(args)

	params := make(map[string]interface{})

	if len(args.URLs) > 0 {
		params[`urls`] = args.URLs
	}

	if response, err := self.browser.Tab().RPC(`Network`, `getCookies`, params); err == nil {
		cookies := make([]*Cookie, 0)

		for _, res := range response.R().Slice(`cookies`) {
			cookie := maputil.M(res)

			c := &Cookie{
				Name:     cookie.String(`name`),
				Value:    cookie.String(`value`),
				Domain:   cookie.String(`domain`),
				Path:     cookie.String(`path`),
				Size:     int(cookie.Int(`size`)),
				HttpOnly: cookie.Bool(`httpOnly`),
				SameSite: cookie.String(`sameSite`),
				Session:  cookie.Bool(`session`),
			}

			if expiresAt := cookie.Int(`expires`); expiresAt > 0 {
				expiry := time.Unix(expiresAt, 0)
				c.Expires = &expiry
			}

			cookies = append(cookies, c)
		}

		return cookies, nil
	} else {
		return nil, err
	}
}

// Set a cookie.
func (self *Commands) Set(cookie *Cookie) error {
	if response, err := self.browser.Tab().RPC(`Network`, `setCookie`, cookie.native()); err == nil {
		if response.R().Bool(`success`) {
			return nil
		} else {
			return fmt.Errorf("Failed to set cookie")
		}
	} else {
		return err
	}
}

type DeleteArgs struct {
	// Deletes all cookies with the given name where domain and path match the given URL.
	URL string `json:"url"`

	// If specified, deletes only cookies with this exact domain.
	Domain string `json:"domain"`

	// If specified, deletes only cookies with this exact path.
	Path string `json:"path"`
}

// Deletes a cookie by name, and optionally matching additional criteria.
func (self *Commands) Delete(name string, args *DeleteArgs) error {
	if args == nil {
		args = &DeleteArgs{}
	}

	defaults.SetDefaults(args)

	params := make(map[string]interface{})

	params[`name`] = name

	if v := args.URL; v != `` {
		params[`url`] = v
	}

	if v := args.URL; v != `` {
		params[`url`] = v
	}

	if v := args.Domain; v != `` {
		params[`domain`] = v
	}

	if v := args.Path; v != `` {
		params[`path`] = v
	}

	return self.browser.Tab().AsyncRPC(`Network`, `deleteCookies`, params)
}

// Clears all browser cookies.
func (self *Commands) Clear() error {
	return self.browser.Tab().AsyncRPC(`Network`, `clearBrowserCookies`, nil)
}
