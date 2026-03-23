// Commands for interacting with the browser's cookie storage backend.
package cookies

import (
	"fmt"
	"time"

	"github.com/ghetzel/friendscript"
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/playwright-community/playwright-go"
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

// Returns the cookie serialized the way CDP expects it.
func (self *Cookie) native() map[string]any {
	params := make(map[string]any)

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

func New(browser *browser.Browser) *Commands {
	cmd := &Commands{}

	cmd.browser = browser
	cmd.Module = friendscript.CreateModule(cmd)

	return cmd
}

type ListArgs struct {
	// an array of strings representing the URLs to retrieve cookies for.  If omitted, the
	// URL of the current browser tab will be used
	Urls []string `json:"urls"`

	// A list of cookie names to include in the output.  If non-empty, only these cookies will appear (if present).
	Names []string `json:"names"`
}

// List all cookies, either for the given set of URLs or for the current tab (if omitted).
func (self *Commands) List(args *ListArgs) ([]*Cookie, error) {
	if args == nil {
		args = new(ListArgs)
	}

	defaults.SetDefaults(args)

	var cookies = make([]*Cookie, 0)

	if pg := self.browser.Page(); pg != nil {
		if ckk, err := pg.Context().Cookies(args.Urls...); err == nil {
			for _, ck := range ckk {
				var cookie = &Cookie{
					Name:     ck.Name,
					Value:    ck.Value,
					Domain:   ck.Domain,
					Path:     ck.Path,
					Size:     len(ck.Value),
					HttpOnly: ck.HttpOnly,
					SameSite: string(*ck.SameSite),
				}

				// if we're filtering on cookie name, skip cookies that aren't in the list
				if len(args.Names) > 0 {
					if !sliceutil.ContainsString(args.Names, ck.Name) {
						continue
					}
				}

				if expiresAt := ck.Expires; expiresAt > 0 {
					var expiry = time.Unix(int64(expiresAt), 0)
					cookie.Expires = &expiry
				}

				cookies = append(cookies, cookie)
			}

			return cookies, nil
		} else {
			return cookies, err
		}
	} else {
		return cookies, browser.NoActivePage
	}
}

// A variant of cookies::list that returns matching cookies as a map of name=value pairs.
func (self *Commands) Map(args *ListArgs) (map[string]string, error) {
	var data = make(map[string]string)

	if cookies, err := self.List(args); err == nil {
		for _, cookie := range cookies {
			data[cookie.Name] = cookie.Value
		}
	} else {
		return nil, err
	}

	return data, nil
}

// Get a cookie by its name.
func (self *Commands) Get(name string) (*Cookie, error) {
	if cookies, err := self.List(nil); err == nil {
		for _, cookie := range cookies {
			if cookie.Name == name {
				return cookie, nil
			}
		}

		return nil, fmt.Errorf("no such cookie %q", name)
	} else {
		return nil, err
	}
}

// Set a cookie.
func (self *Commands) Set(cookie *Cookie) error {
	return browser.NotImplemented
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
		args = new(DeleteArgs)
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		return pg.Context().ClearCookies(playwright.BrowserContextClearCookiesOptions{
			Name:   name,
			Domain: args.Domain,
			Path:   args.Path,
		})
	} else {
		return browser.NoActivePage
	}
}

// Clears all browser cookies.
func (self *Commands) Clear() error {
	if pg := self.browser.Page(); pg != nil {
		return pg.Context().ClearCookies()
	} else {
		return browser.NoActivePage
	}
}
