// Commands for inspecting and manipulating the current DOM document and browser viewport.
package page

import (
	"fmt"
	"io"
	"os"

	"go.gary.cool/friendscript"
	"go.gary.cool/go-stockutil/typeutil"
	"go.gary.cool/go-webfriend/browser"
)

type Commands struct {
	friendscript.Module
	browser *browser.Browser
}

func New(browser *browser.Browser) *Commands {
	var cmd = new(Commands)

	cmd.browser = browser
	cmd.Module = friendscript.CreateModule(cmd)

	return cmd
}

func getWritableDestination(cmd *Commands, destination any) (writer io.Writer, filename string, merr error) {
	if destination != nil {
		if w, ok := destination.(io.Writer); ok {
			writer = w
		} else if target := typeutil.String(destination); target != `` {

			if newPath, w, err := cmd.browser.GetWriterForPath(target); err == nil && w != nil {
				writer = w
				filename = newPath
			} else if target == `temporary` {
				if temp, err := os.CreateTemp(``, ``); err == nil {
					writer = temp
					filename = temp.Name()
				} else {
					merr = err
					return
				}
			} else if file, err := os.Create(target); err == nil {
				writer = file
				filename = target
			} else {
				merr = err
				return
			}
		} else {
			merr = fmt.Errorf("Unsupported destination %T; expected string or io.Writer", destination)
			return
		}
	}

	if writer == nil {
		merr = fmt.Errorf("unspecified destination")
	}

	return
}
