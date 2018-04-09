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

func (self *Commands) Pdf(filenameOrWriter interface{}, args *PdfArgs) error {
	var dest io.Writer

	switch filenameOrWriter.(type) {
	case string:
		if d, err := os.Create(filenameOrWriter.(string)); err == nil {
			dest = d
		} else {
			return err
		}
	case io.Writer:
		dest = filenameOrWriter.(io.Writer)
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
