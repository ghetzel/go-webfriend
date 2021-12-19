package page

import (
	"fmt"
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/log"
)

type CastArgs struct {
	// The amount of time to wait for Castable devices to appear.
	Timeout time.Duration `json:"timeout" default:"30s"`
}

// Cast the current tab to a Google Cast receiver.
func (self *Commands) Cast(device string, args *CastArgs) error {
	if args == nil {
		args = &CastArgs{}
	}

	defaults.SetDefaults(args)

	if err := self.browser.Tab().AsyncRPC(`Cast`, `enable`, map[string]interface{}{
		`presentationUrl`: self.browser.Tab().Info().URL,
	}); err != nil {
		return fmt.Errorf("failed to enable Cast: %v", err)
	}

	var start = time.Now()

	for time.Since(start) < args.Timeout {
		log.Debugf("poll (%v)", time.Since(start))
		if sinks := self.browser.Tab().CastSinks(); len(sinks) > 0 {
			log.DumpJSON(sinks)
			return nil
		}

		time.Sleep(125 * time.Millisecond)
	}

	return nil
}
