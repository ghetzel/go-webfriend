package core

import (
	"github.com/ghetzel/go-stockutil/sliceutil"
)

// Inject Javascript into the current page, evaluate it, and return the results.
// The script is wrapped in an anonymous function whose return value will be
// returned from this command as a native data type.
//
// Scripts will have access to all exported variables in the calling script (exported with the
// "export" command). They are available to injected script as a plain object accessible using
// the "webfriend" variable.
//
// #### Examples
//
// ##### Use the `webfriend` local variable to access Friendscript data from JavaScript, make a change, then pass it back.
// ```
// $pi = 3.0
//
// javascript "return webfriend.pi + 0.14" -> $pi
// # $pi is now 3.14
// ```
//
// ##### Inject a script that will retrieve all `<a>` tags on a page and return their "href" attributes (if present).
// ```
// javascript begin
//
//	var as = document.querySelectorAll('a');
//	var hrefs = [];
//
//	for(var i = 0; i < as.length; i++) {
//	    var a = as[i];
//	    var href = a.getAttribute('href');
//
//	    if (href && href.length) {
//	        hrefs.push(href);
//	    }
//	}
//
//	return hrefs;
//
// end -> $result
//
// # $result now contains an array of zero or more strings as returned from JavaScript.
// ```
func (self *Commands) Javascript(script string) (interface{}, error) {
	return self.browser.Tab().Evaluate(script, self.exported...)
}

// Specify a variable that should be exported to external Javascripts.
func (self *Commands) Export(name string) error {
	if !sliceutil.ContainsString(self.exported, name) {
		self.exported = append(self.exported, name)
	}

	return nil
}

// Specify a variable that should be unexported.
func (self *Commands) Unexport(name string) error {
	for i, vname := range self.exported {
		if vname == name {
			self.exported = append(self.exported[:i], self.exported[i+1:]...)
		}
	}

	return nil
}
