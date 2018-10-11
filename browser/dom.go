package browser

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

var ElementPollTimeout = time.Second

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

		log.Debugf("Store element %d: %v", element.ID(), element)
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

// Retrieve all elements matching the given name
func (self *Document) ElementsByName(name string) []*Element {
	elements := make([]*Element, 0)

	self.elements.Range(func(_ interface{}, v interface{}) bool {
		if el, ok := v.(*Element); ok {
			if el.Name() == name {
				elements = append(elements, el)
			}
		}

		return true
	})

	return elements
}

// Retrieve a known element by its Node ID
func (self *Document) Element(id int) (*Element, bool) {
	start := time.Now()

	for {
		if time.Since(start) > ElementPollTimeout {
			break
		}

		if v, ok := self.elements.Load(id); ok {
			if el, ok := v.(*Element); ok {
				return el, true
			}
		}
	}

	return nil, false
}

// Retrieve a known element by its Node ID
func (self *Document) ElementByBackendId(id int) (*Element, bool) {
	var element *Element

	self.elements.Range(func(key interface{}, el interface{}) bool {
		if elem := el.(*Element); elem.backendId == id {
			element = elem
			return false
		}

		return true
	})

	if element == nil {
		return nil, false
	} else {
		return element, true
	}
}

// Return the root element of the current document.
func (self *Document) Root() (*Element, error) {
	if self.root == nil {
		if rv, err := self.tab.RPC(`DOM`, `getDocument`, map[string]interface{}{
			`pierce`: true,
			`depth`:  -1,
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
	if result, err := self.Evaluate(`return [document.documentElement.scrollWidth, document.documentElement.scrollHeight]`); err == nil {
		if sz := typeutil.V(result).Slice(); len(sz) == 2 {
			return sz[0].Float(), sz[1].Float(), nil
		} else {
			return 0, 0, fmt.Errorf("Invalid response while retrieving page dimensions")
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

	if rv, err := self.Evaluate(fmt.Sprintf(`return document.querySelectorAll(%q)`, selector)); err == nil {
		results := make([]*Element, 0)

		for _, el := range sliceutil.Sliceify(rv) {

			if element, ok := el.(*Element); ok {
				results = append(results, element)
			}
		}

		return results, nil
	} else {
		return nil, err
	}
}

// Remove the given element from the document.
func (self *Document) RemoveElement(element *Element) error {
	if element != nil {
		defer self.elements.Delete(element.ID())

		_, err := self.tab.RPC(`DOM`, `removeNode`, map[string]interface{}{
			`nodeId`: element.ID(),
		})

		return err
	} else {
		return nil
	}
}

func (self *Document) Evaluate(stmt string) (interface{}, error) {
	callGroupId := stringutil.UUID().Base58()

	if rv, err := self.tab.RPC(`Runtime`, `evaluate`, map[string]interface{}{
		`expression`: fmt.Sprintf(
			"%s;\nvar fn_%s = function(){ %s }.bind(webfriend); fn_%s()",
			self.prescript(),
			callGroupId,
			stmt,
			callGroupId,
		),
		`returnByValue`: false,
		`awaitPromise`:  false,
		`objectGroup`:   callGroupId,
	}); err == nil {
		defer self.tab.releaseObjectGroup(callGroupId)
		out := maputil.M(rv.Result)

		// return runtime exceptions as errors
		if exc := out.Get(`exceptionDetails`); !exc.IsZero() {
			excM := maputil.M(exc)

			return nil, fmt.Errorf(
				"Evaluation error: %v",
				excM.String(`exception.description`, excM.String(`text`)),
			)
		} else if returnOid := out.String(`result.objectId`); returnOid != `` {
			// recursively populate the output result and return it as a native value
			return self.tab.getJavascriptResponse(maputil.M(out.Get(`result`)))
		} else if returnValue := out.Get(`result.value`).Value; returnValue != nil {
			return returnValue, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

func (self *Document) EvaluateOn(selector Selector, stmt string) (interface{}, error) {
	var element *Element

	if selector.IsNone() {
		if root, err := self.Root(); err == nil {
			element = root
		} else {
			return ``, err
		}
	} else {
		if elements, err := self.Query(selector, nil); err == nil {
			if len(elements) == 1 {
				element = elements[0]
			} else {
				return ``, fmt.Errorf("Ambiguous selector returned %d elements.", len(elements))
			}
		} else {
			return ``, err
		}
	}

	if element == nil {
		return ``, fmt.Errorf("Could not find element.")
	}

	return element.Evaluate(stmt)
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

func (self *Document) prescript() string {
	if scopeable := self.tab.browser.scopeable; scopeable != nil {
		if scope := scopeable.Scope(); scope != nil {
			out := `var webfriend = `

			if data, err := json.Marshal(scope.Data()); err == nil {
				out += string(data)

				return out
			}
		}
	}

	return ``
}
