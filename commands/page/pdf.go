package page

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"github.com/ghetzel/go-stockutil/maputil"
)

type PdfArgs struct {
}

// Render the current page as a PDF document, writing it to the given filename or writable
// destination object.
func (self *Commands) Pdf(destination interface{}, args *PdfArgs) error {
	var dest io.Writer

	switch destination.(type) {
	case string:
		if d, err := os.Create(destination.(string)); err == nil {
			dest = d
		} else {
			return err
		}
	case io.Writer:
		dest = destination.(io.Writer)
	default:
		return fmt.Errorf("Must specify either a filename or io.Writer destination")
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
