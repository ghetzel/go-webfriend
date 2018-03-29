package core

import "fmt"

// Inject Javascript into the current page, evaluate it, and return the results.
// The script is wrapped in an anonymous function whose return value will be
// returned from this command as a native data type.
//
// Scripts will have access to all local variables in the calling script that
// are defined at the time of invocation. They are available to injected scripts
// as a plain object accessible using the "this" variable.
//
func (self *Commands) Javascript(script interface{}) (interface{}, error) {
	return nil, fmt.Errorf(`NI`)
}
