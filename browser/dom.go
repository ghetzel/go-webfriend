package browser

import (
	"fmt"
)

type Document struct {
	tab    *Tab
	parent *Document
	root   *Element
}

func NewDocument(tab *Tab, parent *Document) *Document {
	doc := &Document{
		tab:    tab,
		parent: parent,
	}

	doc.populate()

	return doc
}

func (self *Document) RootID() string {
	return self.root.ID()
}

func (self *Document) String() string {
	return fmt.Sprintf("%v", self.root)
}

func (self *Document) Reset() error {
	self.root = nil
	return self.populate()
}

func (self *Document) populate() error {
	if reply, err := self.tab.RPC(`DOM`, `GetDocument`, map[string]interface{}{
		`Pierce`: false,
	}); err == nil {
		log.Debugf("%#v", reply)
		// self.root = newElementFromNode(self, , nil)
		return nil
	} else {
		return err
	}

	return nil
}
