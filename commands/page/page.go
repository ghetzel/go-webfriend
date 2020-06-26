// Commands for inspecting and manipulating the current DOM document and browser viewport.
package page

import (
	"github.com/ghetzel/friendscript"
	"github.com/ghetzel/go-webfriend/browser"
)

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
