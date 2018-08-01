package browser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Top    int `json:"top"`
	Left   int `json:"left"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
}

type Selector string

func (self Selector) IsNone() bool {
	return (self == `none` || self == ``)
}

func (self Selector) IsAnnotated() bool {
	return stringutil.IsSurroundedBy(self, `@`, `]`)
}

func (self Selector) GetAnnotation() (string, string, error) {
	var atype string
	var inner string

	if self.IsAnnotated() {
		expr := strings.TrimPrefix(string(self), `@`)
		expr = strings.TrimSuffix(expr, `]`)
		atype, inner = stringutil.SplitPair(expr, `[`)
	} else {
		atype = `css`
		inner = string(self)
	}

	switch atype {
	case ``:
		atype = `text`
	case `xpath`, `css`:
		break
	default:
		return ``, ``, fmt.Errorf("Unsupported annotation type %q", atype)
	}

	return atype, inner, nil
}

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

// Return the parent element of this element, or nil if there isn't one.
func (self *Element) Parent() *Element {
	if parent, ok := self.document.Element(self.parent); ok {
		return parent
	}

	return nil
}

// Return a map representation of the element. This is how values are exposed in
// the Friendscript runtime environment.
func (self *Element) ToMap() map[string]interface{} {
	output := map[string]interface{}{
		`id`:         self.id,
		`name`:       self.name,
		`attributes`: self.attributes,
	}

	if l := len(self.Children()); l > 0 {
		output[`child_count`] = l
	}

	if self.value != `` {
		output[`text`] = self.value
	}

	if position, err := self.Position(); err == nil {
		output[`position`] = position
	} else {
		log.Warningf("Error retrieving element position: %v", err)
	}

	return output
}

func (self *Element) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.ToMap())
}

// Satisifies the fmt.Stringer interface.
func (self *Element) String() string {
	return fmt.Sprintf("[NODE %v] %v", self.id, self.name)
}

// Retrieve the text value of the element.
func (self *Element) Text() string {
	return self.value
}

// Retrieve the current attributes on the element.
func (self *Element) Attributes() map[string]interface{} {
	return maputil.DeepCopy(self.attributes)
}

func (self *Element) ID() int {
	return self.id
}

// Retrieve the current position and dimensions of the element.
func (self *Element) Position() (Dimensions, error) {
	if result, err := self.Evaluate(`return Object.assign({}, this.getBoundingClientRect().toJSON())`); err == nil {
		dimensions := maputil.M(result)

		return Dimensions{
			Width:  int(dimensions.Int(`width`)),
			Height: int(dimensions.Int(`height`)),
			Top:    int(dimensions.Int(`top`)),
			Left:   int(dimensions.Int(`left`)),
			Bottom: int(dimensions.Int(`bottom`)),
			Right:  int(dimensions.Int(`right`)),
		}, nil
	} else {
		return Dimensions{}, err
	}
}

// Loads all child elements under this element.
func (self *Element) Children() []*Element {
	if !self.loadedChildren {
		// setup an accumulator that will capture all setChildNodes events received between
		// now and the end of the RequestChildNodes call
		accumulator, _ := self.document.tab.CreateAccumulator(`DOM.setChildNodes`)
		defer accumulator.Destroy()

		if _, err := self.document.tab.RPC(`DOM`, `requestChildNodes`, map[string]interface{}{
			`nodeId`: self.id,
			`pierce`: true,
			`depth`:  1,
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

// Retrieve the current attributes on this node and update our local copy.
func (self *Element) RefreshAttributes() error {
	if rv, err := self.document.tab.RPC(`DOM`, `getAttributes`, map[string]interface{}{
		`nodeId`: self.ID(),
	}); err == nil {
		self.setAttributesFromInterleavedArray(maputil.M(rv.Result).Slice(`attributes`))
		return nil
	} else {
		return err
	}
}

// Set the given named attribute to the stringified output of value.
func (self *Element) SetAttribute(attrName string, value interface{}) error {
	_, err := self.document.tab.RPC(`DOM`, `setAttributeValue`, map[string]interface{}{
		`nodeId`: self.ID(),
		`name`:   attrName,
		`value`:  typeutil.V(value).String(),
	})

	return err
}

// Focus the current element.
func (self *Element) Focus() error {
	_, err := self.document.tab.RPC(`DOM`, `focus`, map[string]interface{}{
		`nodeId`: self.ID(),
	})

	return err
}

// Click on the current element.
func (self *Element) Click() error {
	_, err := self.Evaluate(`return this.click()`)
	return err
}

// Remove the element.
func (self *Element) Remove() error {
	_, err := self.Evaluate(`this.remove()`)
	return err
}

func (self *Element) Highlight(r int, g int, b int, a float64) error {
	r = int(mathutil.Clamp(float64(r), 0, 255))
	g = int(mathutil.Clamp(float64(g), 0, 255))
	b = int(mathutil.Clamp(float64(b), 0, 255))
	a = mathutil.Clamp(a, 0, 1)

	return self.document.tab.AsyncRPC(`Overlay`, `highlightNode`, map[string]interface{}{
		`highlightConfig`: map[string]interface{}{
			`contentColor`: map[string]interface{}{
				`r`: r,
				`g`: g,
				`b`: b,
				`a`: a,
			},
		},
		`nodeId`: self.id,
	})
}

// Evaluate the given JavaScript as an anonymous function on the current element.
func (self *Element) Evaluate(script string) (interface{}, error) {
	if rv, err := self.document.tab.RPC(`DOM`, `resolveNode`, map[string]interface{}{
		`nodeId`: self.ID(),
	}); err == nil {
		remoteObject := maputil.M(rv.Result)
		callGroupId := stringutil.UUID().String()

		if oid := remoteObject.String(`object.objectId`); oid != `` {
			if rv, err := self.document.tab.RPC(`Runtime`, `callFunctionOn`, map[string]interface{}{
				`objectId`: oid,
				`functionDeclaration`: fmt.Sprintf(
					"function(){ try{ %s; %s } catch(e) { return e; }",
					self.document.prescript(),
					script,
				),
				`returnByValue`: false,
				`awaitPromise`:  false,
				`objectGroup`:   callGroupId,
			}); err == nil {
				defer self.document.tab.releaseObjectGroup(callGroupId)
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
					return self.document.tab.getJavascriptResponse(maputil.M(out.Get(`result`)))
				} else if returnValue := out.Get(`result.value`).Value; returnValue != nil {
					return returnValue, nil
				} else {
					return nil, nil
				}
			} else {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("Unable to determine RemoteObjectID for node %d", self.ID())
		}
	} else {
		return nil, err
	}
}

// Prints this element and all subelements.
func (self *Element) TreeString(depth int) string {
	output := ``

	switch self.name {
	case `#text`:
		output += strings.Repeat(`  `, depth) + strings.TrimSpace(self.value) + "\n"

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

		output += line + "\n"
	}

	for _, child := range self.Children() {
		output += child.TreeString(depth + 1)
	}

	return output
}

func (self *Element) setAttributesFromInterleavedArray(attrpairs []typeutil.Variant) {
	attributes := make(map[string]interface{})

	for i := 0; i < len(attrpairs); i += 2 {
		if (i + 1) < len(attrpairs) {
			attributes[attrpairs[i].String()] = attrpairs[i+1].Auto()
		}
	}

	self.attributes = attributes
}
