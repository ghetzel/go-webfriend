package scripting

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/rxutil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

var rxInterpolate = regexp.MustCompile(`({[^}]+})`)
var placeholderVarName = `_`
var maxInterpolateSequences = 64

type tracer int

type Scope struct {
	parent   *Scope
	data     map[string]interface{}
	isolated bool
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent: parent,
		data:   make(map[string]interface{}),
	}
}

func NewIsolatedScope(parent *Scope) *Scope {
	scope := NewScope(parent)
	scope.isolated = true
	return scope
}

func (self *Scope) Level() int {
	if self.parent == nil {
		return 0
	} else {
		return self.parent.Level() + 1
	}
}

func (self *Scope) String() string {
	if data, err := json.MarshalIndent(self.Data(), ``, `  `); err == nil {
		return string(data)
	} else {
		return err.Error()
	}
}

func (self *Scope) Data() map[string]interface{} {
	output := make(map[string]interface{})

	maputil.Walk(self.data, func(value interface{}, path []string, isLeaf bool) error {
		if typeutil.IsArray(value) {
			maputil.DeepSet(output, path, value)
			return maputil.SkipDescendants

		} else if isLeaf {
			if _, ok := value.(emptyValue); ok {
				maputil.DeepSet(output, path, nil)
			} else {
				maputil.DeepSet(output, path, value)
			}
		}

		return nil
	})

	return output
}

func (self *Scope) Declare(key string) {
	if key == `` || key == placeholderVarName {
		return
	}

	var e emptyValue
	key = self.prepVariableName(key)

	// log.Infof("DECL scope(%d)[%v]", self.Level(), key)
	maputil.DeepSet(self.data, strings.Split(key, `.`), e)
}

func (self *Scope) Set(key string, value interface{}) {
	key = self.prepVariableName(key)
	scope := self.OwnerOf(key)
	scope.set(key, value)
}

func (self *Scope) Get(key string, fallback ...interface{}) interface{} {
	value, _ := self.get(key, fallback...)

	// the emptyValue type is used by the "declare" statement to put a non-nil placeholder
	// value in a scope for the purpose of occupying they key.  When used as a value outside
	// of this package, it should be nil.
	if _, ok := value.(emptyValue); ok {
		return nil
	}

	return value
}

// Returns the scope that "owns" the given key.  This works by first checking for an
// already-set key in the current scope.  If none exists, the parent scope
// is consulted for non-nil values (and so on, all the way up the scope chain).
//
// If none of the ancestor scopes have a non-nil value at the given key, the current
// scope becomes the owner of the key and will be returned.
//
func (self *Scope) OwnerOf(key string) *Scope {
	if self.isolated || self.IsLocal(key) {
		return self
	} else {
		_, scope := self.get(key)
		return scope
	}
}

func (self *Scope) IsLocal(key string) bool {
	if _, ok := maputil.DeepGet(self.data, strings.Split(key, `.`), tracer(0)).(tracer); ok {
		return false
	}

	return true
}

func (self *Scope) set(key string, value interface{}) {
	if key == `` || key == placeholderVarName {
		return
	}

	if value == nil {
		value = new(emptyValue)
	} else if v, err := exprToValue(value); err == nil {
		value = v
	} else {
		panic(fmt.Errorf("Cannot set %v: %v", key, err))
	}

	value = intIfYouCan(value)
	value = mapifyStruct(value)

	// log.Infof("SSET scope(%d)[%v] = %T(%v)", self.Level(), key, value, value)
	maputil.DeepSet(self.data, strings.Split(key, `.`), value)
}

func (self *Scope) get(key string, fallback ...interface{}) (interface{}, *Scope) {
	key = self.prepVariableName(key)

	if v := maputil.DeepGet(self.data, strings.Split(key, `.`)); v != nil {
		// return *copies* of compound types
		if typeutil.IsMap(v) {
			v = maputil.DeepCopy(v)
		} else if typeutil.IsArray(v) {
			v = sliceutil.Sliceify(v)
		}

		// log.Debugf("SGET scope(%d)[%v] -> %T(%v)", self.Level(), key, v, v)
		return v, self
	} else if self.parent != nil {
		// log.Debugf("SGET scope(%d)[%v] -> PARENT", self.Level(), key)

		if v, scope := self.parent.get(key, fallback...); v != nil {
			return v, scope
		}
	}

	if len(fallback) > 0 {
		// log.Debugf("SGET scope(%d)[%v] -> %T(%v) FALLBACK", self.Level(), key, fallback[0], fallback[0])
		return fallback[0], self
	} else {
		// log.Debugf("SGET scope(%d)[%v] -> nil FALLBACK", self.Level(), key)
		return new(emptyValue), self
	}
}

func (self *Scope) Interpolate(in string) string {
	for i := 0; i < maxInterpolateSequences; i++ {
		if match := rxutil.Match(rxInterpolate, in); match != nil {
			seq := match.Group(1)
			seq = stringutil.Unwrap(seq, `{`, `}`)
			value := self.Get(seq)

			in = match.ReplaceGroup(1, fmt.Sprintf("%v", value))
		} else {
			break
		}
	}

	return in
}

func (self *Scope) prepVariableName(key string) string {
	key = strings.TrimPrefix(key, `$`)

	return key
}
