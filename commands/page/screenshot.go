package page

import (
	"io"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/pkg/errors"
	"github.com/playwright-community/playwright-go"
)

type ScreenshotArgs struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x" default:"-1"`
	Y      int `json:"y" default:"-1"`

	// The output image format of the screenshot.  Valid options are "png" or "jpeg".
	Format string `json:"format" default:"png"`

	// The quality of the image used during encoding.  Only applies to "jpeg" format.
	Quality int `json:"quality" default:"100"`

	// Whether the given destination should be automatically closed for writing after the
	// screenshot is written.
	Autoclose bool `json:"autoclose" default:"true"`

	// Automatically resize the screen to the width and height.
	Autoresize bool `json:"autoresize" default:"true"`

	// A stylesheet to apply while taking the screenshot.
	Style string `json:"style"`

	// Hide the default page background to produce transparent screenshots.
	OmitBackground bool `json:"omitbackground"`

	// Disable CSS animations and transitions when taking the screenshot.
	AllowAnimations bool `json:"animations"`
}

type ScreenshotResponse struct {
	// The width of the screenshot (in pixels).
	Width int `json:"width"`

	// The height of the screenshot (in pixels).
	Height int `json:"height"`

	// The X position (relative to the viewport) the screenshot was taken at.
	X int `json:"x"`

	// The Y position (relative to the viewport) the screenshot was taken at.
	Y int `json:"y"`

	// The filesystem path that the screenshot was written to.
	Path string `json:"path,omitempty"`

	// The size of the screenshot (in bytes).
	Size int64 `json:"size,omitempty"`
}

// Render the current page as a PNG or JPEG image, writing it to the given filename or writable
// destination object.
//
// If the filename is the string `"temporary"`, a file will be created in the system's
// temporary area (e.g.: `/tmp`) and the screenshot will be written there.  It is the caller's
// responsibility to remove the temporary file if desired.  The temporary file path is available in
// the return object's `path` parameter.
func (self *Commands) Screenshot(destination any, args *ScreenshotArgs) (*ScreenshotResponse, error) {
	if args == nil {
		args = new(ScreenshotArgs)
	}

	defaults.SetDefaults(args)
	var response = new(ScreenshotResponse)
	var writer io.Writer

	if w, f, err := getWritableDestination(self, destination); err == nil {
		writer = w
		response.Path = f
	} else {
		return response, err
	}

	if pg := self.browser.Page(); pg != nil {
		var opts playwright.PageScreenshotOptions

		opts.FullPage = playwright.Bool(true)

		switch args.Format {
		case `jpeg`:
			opts.Type = playwright.ScreenshotTypeJpeg
			opts.Quality = playwright.Int(args.Quality)
		default:
			opts.Type = playwright.ScreenshotTypePng
		}

		if args.Width > 0 && args.Height > 0 {
			if err := pg.SetViewportSize(args.Width, args.Height); err == nil {
				response.Width = args.Width
				response.Height = args.Height
				response.X = args.X
				response.Y = args.Y
			} else {
				return response, err
			}

			if args.X > 0 || args.Y > 0 {
				opts.Clip = &playwright.Rect{
					Width:  float64(args.Width),
					Height: float64(args.Height),
					X:      float64(args.X),
					Y:      float64(args.Y),
				}
			}
		}

		if args.Style != `` {
			opts.Style = playwright.String(args.Style)
		}

		if args.AllowAnimations {
			opts.Animations = playwright.ScreenshotAnimationsAllow
		} else {
			opts.Animations = playwright.ScreenshotAnimationsDisabled
		}

		if data, err := pg.Screenshot(opts); err == nil {
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
		return nil, browser.NoActivePage
	}
}
