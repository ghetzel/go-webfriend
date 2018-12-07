// Commonly used commands for basic browser interaction.
package core

import (
	"github.com/ghetzel/friendscript/commands/core"
	"github.com/ghetzel/friendscript/utils"
	"github.com/ghetzel/go-webfriend/browser"
)

type Commands struct {
	*core.Commands
	browser  *browser.Browser
	exported []string
}

func New(browser *browser.Browser, env utils.Scopeable) *Commands {
	cmd := &Commands{
		Commands: core.New(env),
		browser:  browser,
		exported: make([]string, 0),
	}

	cmd.SetInstance(cmd)

	return cmd
}
