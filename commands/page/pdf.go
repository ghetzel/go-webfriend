package page

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/ghetzel/go-stockutil/maputil"
	defaults "github.com/mcuadros/go-defaults"
)

type PdfArgs struct {
	// Whether the given destination should be automatically closed for writing after the
	// PDF is written.
	Autoclose bool `json:"autoclose" default:"true"`
}

// Render the current page as a PDF document, writing it to the given filename or writable
// destination object.
func (self *Commands) Pdf(destination interface{}, args *PdfArgs) error {
	var dest io.Writer

	if args == nil {
		args = &PdfArgs{}
	}

	defaults.SetDefaults(args)

	switch destination.(type) {
	case string:
		filename := destination.(string)

		if _, w, ok := self.browser.GetWriterForPath(filename); ok {
			dest = w
		} else if d, err := os.Create(filename); err == nil {
			dest = d
		} else {
			return err
		}
	case io.Writer:
		dest = destination.(io.Writer)
	default:
		return fmt.Errorf("Must specify either a filename or io.Writer destination")
	}

	if dest == nil {
		return fmt.Errorf("A destination for the PDF must be specified")
	}

	if args.Autoclose {
		if closer, ok := dest.(io.Closer); ok {
			defer closer.Close()
		}
	}

	if rv, err := self.browser.Tab().RPC(`Page`, `printToPDF`, map[string]interface{}{
		`scale`: 1,
	}); err == nil {
		if dataS := maputil.M(rv.Result).String(`data`); len(dataS) > 0 {
			if data, err := base64.StdEncoding.DecodeString(dataS); err == nil {
				_, err := dest.Write(data)
				return err
			} else {
				return fmt.Errorf("decode error: %v", err)
			}
		} else {
			return fmt.Errorf("Empty response")
		}
	} else {
		return err
	}
}
