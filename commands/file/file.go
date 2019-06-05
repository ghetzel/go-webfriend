// File IO commands
package file

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghetzel/friendscript/commands/file"
	"github.com/ghetzel/friendscript/utils"
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
)

type Commands struct {
	*file.Commands
	browser *browser.Browser
}

func New(browser *browser.Browser, env utils.Scopeable) *Commands {
	cmd := &Commands{
		Commands: file.New(env),
		browser:  browser,
	}

	cmd.SetInstance(cmd)

	return cmd
}

// func (self *Commands) Open(filename string) (*os.File, error) {
// 	return os.Open(filename)
// }

// func (self *Commands) Create(filename string) (*os.File, error) {
// 	return os.Create(filename)
// }

type WriteArgs struct {
	// The data to write as a stream.
	Data io.Reader `json:"data"`

	// The data to write as a discrete value.
	Value interface{} `json:"value"`

	// Whether to attempt to close the destination (if possible) after reading/writing.
	Autoclose bool `json:"autoclose" default:"true"`
}

type WriteResponse struct {
	// The filesystem path that the data was written to.
	Path string `json:"path,omitempty"`

	// The size of the data (in bytes).
	Size int64 `json:"size,omitempty"`
}

// Write a value or a stream of data to a file at the given path.  The destination path can be a local
// filesystem path, a URI that uses a custom scheme registered outside of the application, or the string
// "temporary", which will write to a temporary file whose path will be returned in the response.
func (self *Commands) Write(destination interface{}, args *WriteArgs) (*WriteResponse, error) {
	if args == nil {
		args = &WriteArgs{}
	}

	defaults.SetDefaults(args)

	response := &WriteResponse{}

	var writer io.Writer

	if destination != nil {
		if filename, ok := destination.(string); ok {
			if newPath, w, ok := self.browser.GetWriterForPath(filename); ok {
				writer = w
				response.Path = newPath
			} else if filename == `temporary` {
				if temp, err := ioutil.TempFile(``, ``); err == nil {
					writer = temp
					response.Path = temp.Name()
				} else {
					return nil, err
				}
			} else if file, err := os.Create(filename); err == nil {
				writer = file
				response.Path = filename
			} else {
				return nil, err
			}
		} else if w, ok := destination.(io.Writer); ok {
			writer = w
		} else {
			return nil, fmt.Errorf("Unsupported destination %T; expected string or io.Writer", destination)
		}
	}

	if writer == nil {
		return response, fmt.Errorf("A destination must be specified")
	}

	if writer != nil {
		var err error

		if args.Data != nil {
			response.Size, err = io.Copy(writer, args.Data)
		} else if args.Value != nil {
			source := bytes.NewBufferString(fmt.Sprintf("%v", args.Value))
			response.Size, err = io.Copy(writer, source)
		} else {
			err = fmt.Errorf("Must specify source data or a discrete value to write")
		}

		// if whatever write (or write attempt) we just did succeeded...
		if err == nil {
			// if we're supposed to autoclose the destination, give that a shot now
			if args.Autoclose {
				if closer, ok := writer.(io.Closer); ok {
					return response, closer.Close()
				}
			}
		} else {
			return response, err
		}
	} else {
		return response, fmt.Errorf("Unable to write to destination")
	}

	return response, nil
}
