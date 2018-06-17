// Suite of testing-oriented commands that will trigger errors or failures if they aren't satistifed.
//
// The Assert module is aimed at making it easier to write scripts that perform test validation,
// as is frequently done during automated acceptance testing of web applications.
//
package assert

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
