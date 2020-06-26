package webfriend

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/testify/require"
)

func TestPathHandlers(t *testing.T) {
	assert := require.New(t)

	chrome := browser.NewBrowser()

	err := chrome.Launch()
	assert.NoError(err)
	defer chrome.Stop()

	env := NewEnvironment(chrome)
	sawPath := ``

	chrome.RegisterPathWriter(func(path string) (string, io.Writer, error) {
		if sawPath == `` {
			sawPath = path
			return ``, ioutil.Discard, nil
		} else {
			return ``, nil, nil
		}
	})

	_, err = env.EvaluateString(`page::screenshot 'dinglebop'`)
	assert.NoError(err)
	assert.Equal(`dinglebop`, sawPath)

	_, err = env.EvaluateString(`page::screenshot '/dev/null'`)
	assert.NoError(err)
	assert.Equal(`dinglebop`, sawPath)
}
