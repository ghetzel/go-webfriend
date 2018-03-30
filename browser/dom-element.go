package browser

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/ghetzel/go-stockutil/maputil"
)

type Selector string

type Element struct {
	document       *Document
	parent         int
	name           string
	attributes     map[string]interface{}
	value          string
	backendId      int
	id             int
	children       []*Element
	loadedChildren bool
}

func (self *Element) String() string {
	return fmt.Sprintf("[NODE %v] %v", self.id, self.name)
}

func (self *Element) ID() string {
	return fmt.Sprintf("%d", self.id)
}

// Loads all child elements under this element.
func (self *Element) Children() []*Element {
	if !self.loadedChildren {
		// setup an accumulator that will capture all setChildNodes events received between
		// now and the end of the RequestChildNodes call
		accumulator, _ := self.document.tab.CreateAccumulator(`DOM.setChildNodes`)
		defer accumulator.Destroy()

		if _, err := self.document.tab.RPC(`DOM`, `RequestChildNodes`, map[string]interface{}{
			`NodeID`: self.id,
			`Pierce`: true,
			`Depth`:  1,
		}); err == nil {
			// stop receiving events now
			accumulator.Stop()

			for _, event := range accumulator.Events {
				for _, node := range maputil.M(event.Params).Slice(`nodes`) {
					self.children = append(self.children, self.document.addElementFromResult(
						maputil.M(node),
					))
				}
			}
		}

		self.loadedChildren = true
	}

	return self.children
}

// Prints this element and all subelements.
func (self *Element) PrintTree(depth int) {
	switch self.name {
	case `#text`:
		log.Info(strings.Repeat(`  `, depth) + strings.TrimSpace(self.value))

	default:
		attrs := []string{}
		astr := ``

		maputil.Walk(self.attributes, func(value interface{}, path []string, isLeaf bool) error {
			if isLeaf {
				attrs = append(attrs, fmt.Sprintf(
					"%v=\"%v\"",
					color.GreenString(strings.Join(path, `.`)),
					color.YellowString(fmt.Sprintf("%v", value)),
				))
			}

			return nil
		})

		if len(attrs) > 0 {
			astr = ` ` + strings.Join(attrs, ` `)
		}

		line := strings.Repeat(`  `, depth)
		line += color.MagentaString(`<`)
		line += color.RedString(self.name)
		line += astr
		line += color.MagentaString(`>`)

		line += self.value

		line += color.MagentaString(`</`)
		line += color.RedString(self.name)
		line += color.MagentaString(`>`)

		log.Info(line)
	}

	for _, child := range self.Children() {
		child.PrintTree(depth + 1)
	}
}
