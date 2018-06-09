package browser

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type Document struct {
	tab      *Tab
	parent   *Document
	root     *Element
	elements sync.Map
}

func NewDocument(tab *Tab, parent *Document) *Document {
	doc := &Document{
		tab:    tab,
		parent: parent,
	}

	// call this to populate the root element
	doc.Root()

	return doc
}

func (self *Document) Reset() {
	self.elements = sync.Map{}
	self.root = nil
	self.parent = nil
	self.Root()
}

// Create an element from a maputil.Map of element properties and add it to the
// document's element index.
func (self *Document) addElementFromResult(node *maputil.Map) *Element {
	if nodeId := int(node.Int(`nodeId`)); nodeId > 0 {
		var element *Element
		var children = node.Slice(`children`)

		// load the various properties from the given node map into a new elements
		if el, ok := self.elements.Load(nodeId); ok {
			element = el.(*Element)
		} else {
			// build the element
			element = &Element{
				document:   self,
				name:       sliceutil.OrString(node.String(`localName`), node.String(`nodeName`)),
				attributes: make(map[string]interface{}),
				value:      node.String(`nodeValue`),
				backendId:  int(node.Int(`backendNodeId`)),
				id:         nodeId,
			}
		}

		element.setAttributesFromInterleavedArray(node.Slice(`attributes`))
		collapsed := false

		if len(children) == 1 {
			child := maputil.M(children[0])

			if child.Int(`nodeType`) == 3 {
				element.value = strings.TrimSpace(child.String(`nodeValue`))
				collapsed = true
			}
		}

		log.Debugf("Store element %d", element.ID())
		self.elements.Store(nodeId, element)

		if !collapsed {
			for _, child := range children {
				self.addElementFromResult(maputil.M(child))
			}
		}

		// // if this element's parent already exists, and we're not in it...TODO, this is a TODO
		// if parent, ok := self.elements.Load(node.Int(`parentId`)); ok {
		//   TODO: big ol' todo
		// }

		return element
	} else {
		log.Panicf("Received invalid node")
		return nil
	}
}

// Retrieve a known element by its Node ID
func (self *Document) Element(id int) (*Element, bool) {
	if v, ok := self.elements.Load(id); ok {
		if el, ok := v.(*Element); ok {
			return el, true
		}
	}

	return nil, false
}

// Return the root element of the current document.
func (self *Document) Root() (*Element, error) {
	if self.root == nil {
		if rv, err := self.tab.RPC(`DOM`, `getDocument`, map[string]interface{}{
			`pierce`: true,
			`depth`:  1,
		}); err == nil {
			docElem := maputil.M(rv.Result)

			for _, child := range docElem.Slice(`root.children`) {
				node := maputil.M(child)

				switch node.Int(`nodeType`) {
				case 1, 3: // handle element and text nodes
					self.root = self.addElementFromResult(node)
					break
				}
			}

		} else {
			return nil, fmt.Errorf("Failed to get root element: %v", err)
		}
	}

	if self.root == nil {
		return nil, fmt.Errorf("Failed to locate root element")
	}

	return self.root, nil
}

// Retrieve the current document's dimensions (scroll width and height).
func (self *Document) PageSize() (float64, float64, error) {
	if root, err := self.Root(); err == nil {
		if result, err := root.Evaluate(`return [document.documentElement.scrollWidth, document.documentElement.scrollHeight]`); err == nil {
			if sz := typeutil.V(result).Slice(); len(sz) == 2 {
				return sz[0].Float(), sz[1].Float(), nil
			} else {
				return 0, 0, fmt.Errorf("Invalid response while retrieving page dimensions")
			}
		} else {
			return 0, 0, err
		}
	} else {
		return 0, 0, err
	}
}

// Select one or more elements from the current DOM.
func (self *Document) Query(selector Selector, queryRoot *Element) ([]*Element, error) {
	if queryRoot == nil {
		if el, err := self.Root(); err == nil {
			queryRoot = el
		} else {
			return nil, err
		}
	}

	if rv, err := self.tab.RPC(`DOM`, `querySelectorAll`, map[string]interface{}{
		`nodeId`:   queryRoot.ID(),
		`selector`: selector,
	}); err == nil {
		results := make([]*Element, 0)

		for _, nid := range maputil.M(rv.Result).Slice(`nodeIds`) {
			if element, ok := self.Element(int(nid.Int())); ok {
				results = append(results, element)
			}
		}

		return results, nil
	} else {
		return nil, err
	}
}

// Highlight all nodes matching the given selector.
func (self *Document) HighlightAll(selector Selector) error {
	if root, err := self.Root(); err == nil {
		return self.tab.AsyncRPC(`Overlay`, `highlightNode`, map[string]interface{}{
			`highlightConfig`: map[string]interface{}{
				`contentColor`: map[string]interface{}{
					`r`: 0,
					`g`: 128,
					`b`: 128,
					`a`: 0.5,
				},
				`selectorList`: selector,
			},
			`nodeId`: root.ID(),
		})
	} else {
		return err
	}
}

// Recursively print the entire document tree, retrieving elements as neccessary.
func (self *Document) PrintTree() {
	if root, err := self.Root(); err == nil {
		start := time.Now()

		for _, line := range strings.Split(root.TreeString(0), "\n") {
			if !typeutil.IsEmpty(line) {
				log.Info(line)
			}
		}

		log.Debugf("Finished, took %v", time.Since(start))
	} else {
		log.Warning(err)
	}
}

func (self *Document) String() string {
	return fmt.Sprintf("%v", self.root)
}
