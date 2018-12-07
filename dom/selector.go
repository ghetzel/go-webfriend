package dom

import (
	"fmt"
	"strings"

	"github.com/ghetzel/go-stockutil/stringutil"
)

type Selector string

func (self *Selector) String() string {
	return string(*self)
}

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
