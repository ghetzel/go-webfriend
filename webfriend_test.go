package webfriend

import (
	"regexp"
	"testing"

	"github.com/ghetzel/go-stockutil/rxutil"
	"github.com/ghetzel/testify/assert"
)

// func TestPathHandlers(t *testing.T) {
// 	assert := require.New(t)

// 	chrome := browser.NewBrowser()

// 	err := chrome.Launch()
// 	assert.NoError(err)
// 	defer chrome.Stop()

// 	env := NewEnvironment(chrome)
// 	sawPath := ``

// 	chrome.RegisterPathWriter(func(path string) (string, io.Writer, error) {
// 		if sawPath == `` {
// 			sawPath = path
// 			return ``, ioutil.Discard, nil
// 		} else {
// 			return ``, nil, nil
// 		}
// 	})

// 	_, err = env.EvaluateString(`page::screenshot 'dinglebop'`)
// 	assert.NoError(err)
// 	assert.Equal(`dinglebop`, sawPath)

// 	_, err = env.EvaluateString(`page::screenshot '/dev/null'`)
// 	assert.NoError(err)
// 	assert.Equal(`dinglebop`, sawPath)
// }

func TestKeyNonsense(t *testing.T) {
	var kCode = `[ArrowUp][ArrowUp][ArrowDown][ArrowDown][ArrowLeft][ArrowRight][ArrowLeft][ArrowRight]BA`
	var rxKeyCodes = regexp.MustCompile(`(\[[^\]]*?\]|.)`)
	var symbols = rxutil.Match(rxKeyCodes, kCode).AllCaptures()

	assert.Equal(t, []string{
		`[ArrowUp]`,
		`[ArrowUp]`,
		`[ArrowDown]`,
		`[ArrowDown]`,
		`[ArrowLeft]`,
		`[ArrowRight]`,
		`[ArrowLeft]`,
		`[ArrowRight]`,
		`B`,
		`A`,
	}, symbols)
}
