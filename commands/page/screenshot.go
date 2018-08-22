package page

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/ghetzel/go-webfriend/browser"
	defaults "github.com/mcuadros/go-defaults"
)

type ScreenshotArgs struct {
	// If specified, the screenshot will attempt to capture just the matching element.
	Selector browser.Selector `json:"selector,omitempty"`

	// Determines how to handle multiple elements that are matched by Selector.
	// May be "tallest" or "first".
	Use string `json:"use,omitempty" default:"tallest"`

	Width  int `json:"width"`
	Height int `json:"height"`
	X      int `json:"x" default:"-1"`
	Y      int `json:"y" default:"-1"`

	// The output image format of the screenshot.  May be "png" or "jpeg".
	Format string `json:"format" default:"png"`

	// The quality of the image used during encoding.  Only applies to "jpeg" format.
	Quality int `json:"quality"`

	// Whether the given destination should be automatically closed for writing after the
	// screenshot is written.
	Autoclose bool `json:"autoclose" default:"true"`

	// Automatically resize the screen to the width and height.
	Autoresize bool `json:"autoresize" default:"true"`
}

type ScreenshotResponse struct {
	// Details about the element that matched the given selector (if any).
	Element *browser.Element `json:"element,omitempty"`

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
func (self *Commands) Screenshot(destination interface{}, args *ScreenshotArgs) (*ScreenshotResponse, error) {
	if args == nil {
		args = &ScreenshotArgs{}
	}

	defaults.SetDefaults(args)
	docroot := self.browser.Tab().DOM()
	response := &ScreenshotResponse{}

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
		return response, fmt.Errorf("A destination for the screenshot must be specified")
	}

	if args.Autoclose {
		if closer, ok := writer.(io.Closer); ok {
			defer closer.Close()
		}
	}

	// if one or both of the dimensions are not explicitly given, fill them in from the current
	// page dimensions (scroll width/height)
	if args.Width == 0 || args.Height == 0 {
		if pageWidth, pageHeight, err := docroot.PageSize(); err == nil {
			if args.Width == 0 {
				args.Width = int(pageWidth)
			}

			if args.Height == 0 {
				args.Height = int(pageHeight)
			}
		} else {
			return response, fmt.Errorf("Unable to determine page size, and either width or height were not explicitly given.")
		}
	}

	if err := self.screenshotPopulateArgsFromElement(docroot, args, response); err != nil {
		return response, err
	}

	// start building screenshot RPC request
	screenshot := map[string]interface{}{
		`format`:      args.Format,
		`fromSurface`: false,
	}

	// set quality for format="jpeg"
	if args.Format == `jpeg` && args.Quality > 0 {
		screenshot[`quality`] = args.Quality
	}

	// setup clipping region to the given X/Y if specified
	if args.X > 0 || args.Y > 0 {
		screenshot[`clip`] = map[string]interface{}{
			`width`:  args.Width,
			`height`: args.Height,
			`x`:      args.X,
			`y`:      args.Y,
			`scale`:  1.0,
		}
	} else if args.Autoresize && args.Width > 0 && args.Height > 0 {
		// resize viewport to given width and height
		if _, err := self.browser.Tab().RPC(`Emulation`, `setDeviceMetricsOverride`, map[string]interface{}{
			`width`:             args.Width,
			`height`:            args.Height,
			`deviceScaleFactor`: 1.0,
			`mobile`:            false,
		}); err != nil {
			return response, fmt.Errorf("Failed to resize screen: %v", err)
		}
	}

	// take the screenshot
	if reply, err := self.browser.Tab().RPC(`Page`, `captureScreenshot`, screenshot); err == nil {
		defer func() {
			self.browser.Tab().RPC(`Emulation`, `clearDeviceMetricsOverride`, nil)
			self.browser.Tab().RPC(`Emulation`, `resetPageScaleFactor`, nil)
		}()

		response.Width = args.Width
		response.Height = args.Height
		response.X = args.X
		response.Y = args.Y

		// decode the response from base64, write the data to destination
		if data := reply.R().String(`data`); data != `` {
			decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(data))

			if n, err := io.Copy(writer, decoder); err == nil {
				response.Size = n

				return response, nil
			} else {
				return response, fmt.Errorf("Error decoding screenshot response: %v", err)
			}
		} else {
			return response, fmt.Errorf("Empty response received while capturing screenshot")
		}
	} else {
		return response, err
	}
}

func (self *Commands) screenshotPopulateArgsFromElement(docroot *browser.Document, args *ScreenshotArgs, response *ScreenshotResponse) error {
	// if screenshotting an element, find that element now
	if args.Selector != `` {
		if elements, err := docroot.Query(args.Selector, nil); err == nil {
			var winner *browser.Dimensions

			for _, element := range elements {
				if winner == nil {
					if winnerD, err := element.Position(); err == nil {
						winner = &winnerD
						response.Element = element
					} else {
						return fmt.Errorf("Could not determine element dimensions: %v", err)
					}
				} else {
					switch args.Use {
					case `first`:
						break
					case `tallest`:
						if elementD, err := element.Position(); err == nil {
							if elementD.Height > winner.Height {
								winner = &elementD
								response.Element = element
								continue
							}
						} else {
							return fmt.Errorf("Could not determine element dimensions: %v", err)
						}
					default:
						return fmt.Errorf("Unsupported argument for 'use': %q", args.Use)
					}
				}
			}

			if winner != nil {
				if args.Width == 0 {
					args.Width = winner.Width
				}

				if args.Height == 0 {
					args.Height = winner.Height
				}

				if args.X < 0 {
					args.X = winner.Left
				}

				if args.Y < 0 {
					args.Y = winner.Top
				}
			} else {
				return fmt.Errorf("No element found")
			}
		} else {
			return fmt.Errorf("failed to take screenshot of element: %v", err)
		}
	}

	return nil
}
