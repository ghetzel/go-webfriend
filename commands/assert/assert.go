// Suite of testing-oriented commands that will trigger errors or failures if they aren't satistifed.
//
// The Assert module is aimed at making it easier to write scripts that perform test validation,
// as is frequently done during automated acceptance testing of web applications.
//
package assert

import (
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/utils"
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
