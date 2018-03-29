package browser

import (
	"fmt"

	"github.com/mafredri/cdp/protocol/dom"
)

type Selector string

type Element struct {
	document *Document
	node     dom.Node
	parent   *Element
}

func newElementFromNode(document *Document, node dom.Node, parent *Element) *Element {
	document.tab.rpc.DOM.DescribeNode(
		document.tab.browser.ctx(),
		dom.NewDescribeNodeArgs().SetNodeID(node.NodeID),
	)

	return &Element{
		document: document,
		node:     node,
		parent:   parent,
	}
}

func (self *Element) String() string {
	return fmt.Sprintf("[NODE %v] %v", self.node.NodeID, self.node.NodeName)
}

func (self *Element) ID() string {
	return string(self.node.NodeID)
}
