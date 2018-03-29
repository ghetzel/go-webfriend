package page

import (
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

	if rv, err := self.browser.Tab().RPC(`Page`, `PrintToPDF`, map[string]interface{}{}); err == nil {
		dataI := maputil.Get(rv, `Data`)

		if data, ok := dataI.([]byte); ok {
			_, err := dest.Write(data)
			return err
		} else {
			return fmt.Errorf("Invalid response format: %T", dataI)
		}
	} else {
		return err
	}
}
