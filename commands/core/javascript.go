package core

// Inject Javascript into the current page, evaluate it, and return the results.
// The script is wrapped in an anonymous function whose return value will be
// returned from this command as a native data type.
//
// Scripts will have access to all local variables in the calling script that
// are defined at the time of invocation. They are available to injected scripts
// as a plain object accessible using the "webfriend" variable.
//
// #### Examples
//
// ##### Use the `webfriend` local variable to access Friendscript data from JavaScript, make a change, then pass it back.
//
// ```
// $pi = 3.0
// javascript "return webfriend.pi + 0.14" -> $pi
// ```
// The `$pi` variable should be equal to 3.14 after this runs.
//
// ##### Inject a script that will retrieve all `<a>` tags on a page and return their "href" attributes (if present).
//
// ```
// javascript begin
//     var as = document.querySelectorAll('a');
//     var hrefs = [];
//
//     for(var i = 0; i < as.length; i++) {
//         var a = as[i];
//         var href = a.getAttribute('href');
//
//         if (href && href.length) {
//             hrefs.push(href);
//         }
//     }
//
//     return hrefs;
// end
// ```
//
// After running, the `$result` variable will contain an array of zero or more strings.
//
func (self *Commands) Javascript(script string) (interface{}, error) {
	return self.browser.Tab().Evaluate(script)
}
