package page

import (
	"io"

	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
	defaults "go.gary.cool/go-defaults"
	"go.gary.cool/go-stockutil/mathutil"
	"go.gary.cool/go-webfriend/browser"
)

type PdfArgs struct {
	// Whether the given destination should be automatically closed for writing after the PDF is written.
	Autoclose bool `json:"autoclose" default:"true"`

	// Paper format of the PDF pages
	Format string `json:"format"`

	// Orient the PDF in landscape.
	Landscape bool `json:"landscape"`

	// Scale of the rendered page. Between [0.1, 2.0], default is 1.
	Scale float64 `json:"scale"`
}

type PdfResponse struct {
	// The filesystem path that the screenshot was written to.
	Path string `json:"path,omitempty"`

	// The size of the screenshot (in bytes).
	Size int64 `json:"size,omitempty"`
}

// Render the current page as a PDF document, writing it to the given filename or writable
// destination object.
func (self *Commands) Pdf(destination any, args *PdfArgs) (*PdfResponse, error) {
	var response = new(PdfResponse)
	var writer io.Writer

	if args == nil {
		args = new(PdfArgs)
	}

	defaults.SetDefaults(args)

	if w, f, err := getWritableDestination(self, destination); err == nil {
		writer = w
		response.Path = f
	} else {
		return response, err
	}

	if pg := self.browser.Page(); pg != nil {
		var opts playwright.PagePdfOptions

		opts.Landscape = playwright.Bool(args.Landscape)

		if f := args.Format; f != `` {
			opts.Format = playwright.String(f)
		}

		if s := args.Scale; s >= 0 {
			s = mathutil.Clamp(s, 0.1, 2.0)
			opts.Scale = playwright.Float(s)
		}

		if data, err := pg.PDF(opts); err == nil {
			response.Size = int64(len(data))

			if _, err := writer.Write(data); err == nil {
				if args.Autoclose {
					if closer, ok := writer.(io.Closer); ok {
						return response, closer.Close()
					}
				}

				return response, nil
			} else {
				return response, errors.Wrap(err, "bad write")
			}
		} else {
			return response, err
		}
	} else {
		return response, browser.NoActivePage
	}
}
