// Commands for interacting with the browser's cookie storage backend.
package cookies

import (
	"github.com/ghetzel/friendscript"
	"github.com/ghetzel/friendscript/utils"
	"github.com/ghetzel/go-webfriend/browser"
)

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
