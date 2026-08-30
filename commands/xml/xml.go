// Commands for interacting with XML documents
package xml

import (
	"bytes"
	"io"

	"github.com/antchfx/xmlquery"
	"github.com/pkg/errors"
	"go.gary.cool/friendscript"
	defaults "go.gary.cool/go-defaults"
	"go.gary.cool/go-stockutil/typeutil"
	"go.gary.cool/go-webfriend/browser"
	"go.gary.cool/go-webfriend/dom"
)

type Commands struct {
	friendscript.Module
	browser *browser.Browser
}

func New(browser *browser.Browser) *Commands {
	var cmd = new(Commands)

	cmd.browser = browser
	cmd.Module = friendscript.CreateModule(cmd)

	return cmd
}

type QueryArgs struct {
	// Return a list of nodes matching the given XPath expression.
	Xpath string `json:"xpath"`
}

// List all cookies, either for the given set of URLs or for the current tab (if omitted).
func (self *Commands) Query(xmldoc any, args *QueryArgs) ([]*dom.Element, error) {
	if args == nil {
		args = new(QueryArgs)
	}

	defaults.SetDefaults(args)

	if root, err := xmlquery.Parse(readify(xmldoc)); err == nil {
		var matches = make([]*dom.Element, 0)

		if args.Xpath != `` {
			if nodes, err := xmlquery.QueryAll(root, args.Xpath); err == nil {
				for _, node := range nodes {
					if node.Type != xmlquery.ElementNode {
						continue
					}

					var match = &dom.Element{
						Name:       node.Data,
						Text:       node.InnerText(),
						Attributes: make(map[string]any),
					}

					for _, attr := range node.Attr {
						match.Attributes[attr.Name.Local] = typeutil.Auto(attr.Value)
					}

					matches = append(matches, match)
				}
			} else {
				return nil, errors.Wrap(err, "xpath query failed")
			}
		}

		return matches, nil
	} else {
		return nil, err
	}
}

func readify(xmldoc any) (reader io.Reader) {
	if s, ok := xmldoc.(string); ok {
		reader = bytes.NewBufferString(s)
	} else if r, ok := xmldoc.(io.Reader); ok {
		reader = r
	} else {
		reader = bytes.NewBuffer(nil)
	}

	return
}
