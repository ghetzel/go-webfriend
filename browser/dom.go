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

func IsElementNotFoundErr(err error) bool {
	if err != nil {
		if strings.Contains(err.Error(), `Could not find node with given id`) {
			return true
		}
	}

	return false
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
		log.Fatalf("Received invalid node")
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
func (self *Document) Root() *Element {
	if self.root == nil {
		if rv, err := self.tab.RPC(`DOM`, `getDocument`, map[string]interface{}{
			`Pierce`: true,
			`Depth`:  1,
		}); err == nil {
			docElem := maputil.M(rv)

			for _, child := range docElem.Slice(`root.children`) {
				node := maputil.M(child)

				switch node.Int(`nodeType`) {
				case 1, 3: // handle element and text nodes
					self.root = self.addElementFromResult(node)
					break
				}
			}

		} else {
			log.Fatalf("Failed to get root element: %v", err)
		}
	}

	if self.root == nil {
		log.Fatalf("Failed to locate root element")
	}

	return self.root
}

// Select one or more elements from the current DOM.
func (self *Document) Query(selector Selector, queryRoot *Element) ([]*Element, error) {
	if queryRoot == nil {
		queryRoot = self.Root()
	}

	if rv, err := self.tab.RPC(`DOM`, `querySelectorAll`, map[string]interface{}{
		`NodeID`:   queryRoot.ID(),
		`Selector`: selector,
	}); err == nil {
		results := make([]*Element, 0)

		for _, nid := range maputil.M(rv).Slice(`nodeIds`) {
			if element, ok := self.Element(int(nid.Int())); ok {
				results = append(results, element)
			} else {
				log.Warningf(
					"Element %d was returned in a query, but not found in the local DOM cache",
					nid.Int(),
				)
			}
		}

		return results, nil
	} else {
		return nil, err
	}
}

// Recursively print the entire document tree, retrieving elements as neccessary.
func (self *Document) PrintTree() {
	start := time.Now()

	for _, line := range strings.Split(self.Root().TreeString(0), "\n") {
		if !typeutil.IsEmpty(line) {
			log.Info(line)
		}
	}

	log.Debugf("Finished, took %v", time.Since(start))
}

func (self *Document) String() string {
	return fmt.Sprintf("%v", self.root)
}
