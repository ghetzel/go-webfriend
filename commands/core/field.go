package core

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/browser"
)

type FieldArgs struct {
	// Whether to clear the existing contents of the field before entering new data.
	Autoclear bool `json:autoclear"` // true
}

// Locate and enter data into a form input field.
func (self *Commands) Field(selector browser.Selector, value interface{}, args *FieldArgs) (string, error) {
	return ``, fmt.Errorf(`NI`)
}
