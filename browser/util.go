package browser

import (
	"os"
	"os/exec"
	"runtime"
)

func LocateChromeExecutable() string {
	paths := []string{
		os.Getenv(`WEBFRIEND_BROWSER`),
	}

	switch runtime.GOOS {
	case `linux`, `freebsd`:
		paths = append(paths, []string{
			`chromium-browser`,
			`google-chrome`,
		}...)

	case `darwin`:
		paths = append(paths, []string{
			`/Applications/Chromium.app/Contents/MacOS/Chromium`,
			`/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`,
		}...)
	}

	for _, binpath := range paths {
		if path, err := exec.LookPath(binpath); err == nil {
			return path
		}
	}

	return `false`
}
