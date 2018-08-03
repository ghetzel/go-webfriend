package webfriend

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/ghetzel/go-webfriend/browser"
	"github.com/stretchr/testify/require"
)

func TestPathHandlers(t *testing.T) {
	assert := require.New(t)

	chrome := browser.NewBrowser()

	err := chrome.Launch()
	assert.NoError(err)
	defer chrome.Stop()

	env := NewEnvironment(chrome)
	sawPath := ``

	chrome.RegisterPathHandler(func(path string) (string, io.Writer, bool) {
		if sawPath == `` {
			sawPath = path
			return ``, ioutil.Discard, true
		} else {
			return ``, nil, false
		}
	})

	_, err = env.EvaluateString(`page::screenshot 'dinglebop'`)
	assert.NoError(err)
	assert.Equal(`dinglebop`, sawPath)

	_, err = env.EvaluateString(`page::screenshot '/dev/null'`)
	assert.NoError(err)
	assert.Equal(`dinglebop`, sawPath)
}
