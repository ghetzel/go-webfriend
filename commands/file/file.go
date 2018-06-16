// Commands for reading and writing files.
package file

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
