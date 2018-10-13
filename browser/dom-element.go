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
	nodeId        int
	backendNodeId int
	objectId      string
	document      *Document
	lastDetails   *maputil.Map
	textOverride  string
}

// Return a map representation of the element. This is how values are exposed in
// the Friendscript runtime environment.
func (self *Element) refresh() {
	output := map[string]interface{}{}

	if node, err := self.document.tab.RPC(`DOM`, `describeNode`, map[string]interface{}{
		`nodeId`:        self.nodeId,
		`backendNodeId`: self.backendNodeId,
		`objectId`:      self.objectId,
	}); err == nil {
		details := maputil.M(node.R().Get(`node`))

		if nodeId := int(details.Int(`nodeId`)); nodeId > 0 {
			self.nodeId = nodeId
		}

		if backendNodeId := int(details.Int(`backendNodeId`)); backendNodeId > 0 {
			self.backendNodeId = backendNodeId
		}

		output[`name`] = details.String(`localName`)
		output[`attributes`] = self.getAttributesFromInterleavedArray(details.Slice(`attributes`))

		if n := details.Int(`childNodeCount`); n > 0 {
			output[`child_count`] = n
		}

		if v := details.String(`nodeValue`); v != `` {
			output[`text`] = v
		}
	}

	if position, err := self.Position(); err == nil {
		output[`position`] = position
	}

	output[`id`] = self.backendNodeId

	self.lastDetails = maputil.M(output)
}

func (self *Element) refreshIfMissing() {
	if self.lastDetails == nil {
		self.refresh()
	}
}

func (self *Element) ToMap() map[string]interface{} {
	self.refresh()
	return self.lastDetails.Value().(map[string]interface{})
}

func (self *Element) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.ToMap())
}

// Satisifies the fmt.Stringer interface.
func (self *Element) String() string {
	return fmt.Sprintf("[NODE %v] %v", self.backendNodeId, self.TreeString(0))
}

// Retrieve the name of the element.
func (self *Element) Name() string {
	self.refreshIfMissing()

	return self.lastDetails.String(`name`)
}

// Retrieve the text value of the element.
func (self *Element) Text() string {
	self.refreshIfMissing()

	if txt := self.textOverride; txt != `` {
		return txt
	} else {
		return self.lastDetails.String(`text`)
	}
}

// Retrieve the current attributes on the element.
func (self *Element) Attributes() map[string]interface{} {
	self.refreshIfMissing()

	return maputil.DeepCopy(self.lastDetails.Get(`attributes`).Value)
}

func (self *Element) BackendID() int {
	return self.backendNodeId
}

func (self *Element) NodeID() int {
	if self.nodeId == 0 {
		self.refresh()

		if self.nodeId == 0 {
			log.DebugStack()
			log.Panicf("%v: nodeId was explicitly requested before it is available", self)
		}
	}

	return self.nodeId
}

// Retrieve the current position and dimensions of the element.
func (self *Element) Position() (Dimensions, error) {
	if result, err := self.evaluate(`return Object.assign({}, this.getBoundingClientRect().toJSON())`, true); err == nil {
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

// Set the given named attribute to the stringified output of value.
func (self *Element) SetAttribute(attrName string, value interface{}) error {
	_, err := self.evaluate(fmt.Sprintf(
		"return this.setAttribute(%q, %q)",
		attrName,
		typeutil.V(value).String(),
	), true)

	return err
}

// Focus the current element.
func (self *Element) Focus() error {
	_, err := self.evaluate(`return this.focus()`, true)
	return err
}

// Click on the current element.
func (self *Element) Click() error {
	_, err := self.evaluate(`return this.click()`, true)
	return err
}

// Remove the element.
func (self *Element) Remove() error {
	_, err := self.evaluate(`this.remove()`, true)
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
		`nodeId`: self.NodeID(),
	})
}

// Evaluate the given JavaScript as an anonymous function on the current element.
func (self *Element) Evaluate(script string) (interface{}, error) {
	return self.evaluate(script, false)
}

func (self *Element) evaluate(script string, skipPrescript bool) (interface{}, error) {
	if rv, err := self.document.tab.RPC(`DOM`, `resolveNode`, map[string]interface{}{
		`backendNodeId`: self.BackendID(),
	}); err == nil {
		remoteObject := maputil.M(rv.Result)
		callGroupId := stringutil.UUID().String()

		if oid := remoteObject.String(`object.objectId`); oid != `` {
			prescript := `true`

			if !skipPrescript {
				prescript = self.document.prescript()
			}

			decl := fmt.Sprintf(
				"function(){ try{ %s; %s; } catch(e) { return e; } }",
				prescript,
				script,
			)

			if rv, err := self.document.tab.RPC(`Runtime`, `callFunctionOn`, map[string]interface{}{
				`objectId`:            oid,
				`functionDeclaration`: decl,
				`returnByValue`:       false,
				`awaitPromise`:        false,
				`objectGroup`:         callGroupId,
			}); err == nil {
				defer self.document.tab.releaseObjectGroup(callGroupId)
				out := maputil.M(rv.Result)

				// return runtime exceptions as errors
				if exc := out.Get(`exceptionDetails`); !exc.IsZero() {
					excM := maputil.M(exc)

					return nil, fmt.Errorf(
						"Evaluation error:\nline %d, col %d\n%v\n%v",
						excM.Int(`lineNumber`),
						excM.Int(`columnNumber`),
						excM.String(`text`),
						excM.String(`exception.description`),
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
			return nil, fmt.Errorf("Unable to determine RemoteObjectID for node=%d backend=%d", self.NodeID(), self.BackendID())
		}
	} else {
		return nil, err
	}
}

// Prints this element and all subelements.
func (self *Element) TreeString(depth int) string {
	output := ``

	switch name := self.Name(); name {
	case `#comment`:
		output += color.GreenString(`<!-- -->`)
	case `#text`:
		output += strings.Repeat(`  `, depth) + strings.TrimSpace(self.Text()) + "\n"

	default:
		attrs := []string{}
		astr := ``

		maputil.Walk(self.Attributes(), func(value interface{}, path []string, isLeaf bool) error {
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
		line += color.RedString(name)
		line += astr
		line += color.MagentaString(`>`)

		line += self.Text()

		line += color.MagentaString(`</`)
		line += color.RedString(name)
		line += color.MagentaString(`>`)

		output += line + "\n"
	}

	// for _, child := range self.Children() {
	// 	output += child.TreeString(depth + 1)
	// }

	return output
}

func (self *Element) getAttributesFromInterleavedArray(attrpairs []typeutil.Variant) map[string]interface{} {
	attributes := make(map[string]interface{})

	for i := 0; i < len(attrpairs); i += 2 {
		if (i + 1) < len(attrpairs) {
			attributes[attrpairs[i].String()] = attrpairs[i+1].Auto()
		}
	}

	return attributes
}
