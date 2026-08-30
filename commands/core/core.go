// Commonly used commands for basic browser interaction.
package core

import (
	"fmt"

	"go.gary.cool/friendscript/commands/core"
	"go.gary.cool/go-webfriend/browser"
)

type Commands struct {
	*core.Commands
	browser  *browser.Browser
	exported []string
}

func New(browser *browser.Browser) *Commands {
	var cmd = &Commands{
		Commands: core.New(browser),
		browser:  browser,
		exported: make([]string, 0),
	}

	cmd.SetInstance(cmd)

	return cmd
}

func normalizeKeyCodes(
	keycodes string,
	meta bool,
	ctrl bool,
	alt bool,
	shift bool,
) string {
	if shift {
		keycodes = fmt.Sprintf("Shift+%v", keycodes)
	}
	if alt {
		keycodes = fmt.Sprintf("Alt+%v", keycodes)
	}
	if ctrl && meta {
		keycodes = fmt.Sprintf("ControlOrMeta+%v", keycodes)
	} else {
		if ctrl {
			keycodes = fmt.Sprintf("Control+%v", keycodes)
		}
		if meta {
			keycodes = fmt.Sprintf("Meta+%v", keycodes)
		}
	}

	return keycodes
}
