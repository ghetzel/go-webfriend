package browser

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/dom"
	"github.com/gobwas/glob"
	"github.com/mafredri/cdp/devtool"
)

var netTrackingEvents = `Network.{requestWillBeSent,responseReceived,loadingFailed,loadingFinished}`
var domTrackingEvents = `DOM.*`
var consoleEvents = `Console.messageAdded`

var skipItem = errors.New(`skip`)

type TabID string

type NetworkRequestPattern struct {
	URL               string
	ResourceType      string
	InterceptionStage string
	Pattern           glob.Glob
}

func (self *NetworkRequestPattern) ToMap() map[string]interface{} {
	rv := make(map[string]interface{})

	if self.URL != `` {
		rv[`urlPattern`] = self.URL
	}

	if self.ResourceType != `` {
		rv[`resourceType`] = self.ResourceType
	}

	if self.InterceptionStage != `` {
		rv[`interceptionStage`] = self.InterceptionStage
	}

	return rv
}

type NetworkInterceptResponse struct {
	URL          string
	Method       string
	Body         io.Reader
	PostData     map[string]interface{}
	Header       http.Header
	Error        error
	AuthResponse string
	Username     string
	Password     string
	Autoremove   bool
}

func (self *NetworkInterceptResponse) ToMap(id string) map[string]interface{} {
	rv := map[string]interface{}{
		`interceptionId`: id,
	}

	if self.Error == nil {
		if self.Body == nil {
			if v := self.URL; v != `` {
				rv[`url`] = v
			}

			if v := self.Method; v != `` {
				rv[`method`] = v
			}

			if len(self.PostData) > 0 {
				rv[`postData`] = maputil.Join(self.PostData, `=`, `&`)
			}

			if len(self.Header) > 0 {
				rv[`headers`] = self.Header
			}
		} else if data, err := ioutil.ReadAll(self.Body); err == nil {
			raw := &http.Response{
				StatusCode:    200,
				Proto:         `HTTP/1.1`,
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        self.Header,
				Body:          ioutil.NopCloser(bytes.NewBuffer(data)),
				ContentLength: int64(len(data)),
			}

			dest := bytes.NewBuffer(nil)

			if err := raw.Write(dest); err == nil {
				rv[`rawResponse`] = base64.StdEncoding.EncodeToString(dest.Bytes())
			} else {
				log.Warningf("Failed to generate intercept response: %v", err)
				rv[`errorReason`] = `Failed`
			}
		} else {
			log.Warningf("Failed to read intercept body: %v", err)
			rv[`errorReason`] = `Failed`
		}

		switch self.AuthResponse {
		case `Cancel`:
			rv[`authChallengeResponse`] = map[string]interface{}{
				`response`: `Cancel`,
			}

		case `ProvideCredentials`:
			rv[`authChallengeResponse`] = map[string]interface{}{
				`response`: `ProvideCredentials`,
				`username`: self.Username,
				`password`: self.Password,
			}
		}
	} else {
		rv[`errorReason`] = self.Error.Error()
	}

	return rv
}

type NetworkInterceptFunc func(*Tab, *NetworkRequestPattern, *Event) *NetworkInterceptResponse

type NetworkRequest struct {
	ID         string
	Request    *Event
	Response   *Event
	Failure    *Event
	Completion *Event
}

func (self *NetworkRequest) IsCompleted() bool {
	if self.Completion != nil || self.Failure != nil {
		return true
	}

	return false
}

func (self *NetworkRequest) Error() error {
	if self.Failure != nil {
		return errors.New(self.Failure.P().String(`errorText`))
	}

	return nil
}

func (self *NetworkRequest) R() *maputil.Map {
	if self.Response != nil {
		return self.Response.P()
	} else {
		return maputil.M(nil)
	}
}

type PageInfo struct {
	URL      string `json:"url"`
	State    string `json:"state"`
	loaderId string
	frameId  string
}

type Tab struct {
	AfterCommandDelay    time.Duration
	browser              *Browser
	target               *devtool.Target
	rpc                  *RPC
	events               chan *Event
	waiters              sync.Map
	networkRequests      sync.Map
	accumulators         sync.Map
	mostRecentFrameId    int64
	mostRecentFrame      []byte
	mostRecentDimensions []int
	screencasting        bool
	castlock             sync.Mutex
	mostRecentInfo       *PageInfo
	netIntercepts        sync.Map
}

func newTabFromTarget(browser *Browser, target *devtool.Target) (*Tab, error) {
	tab := &Tab{
		browser: browser,
		target:  target,
		events:  make(chan *Event),
		mostRecentInfo: &PageInfo{
			URL:   target.URL,
			State: `initial`,
		},
	}

	return tab, tab.connect()
}

func (self *Tab) Info() *PageInfo {
	return self.mostRecentInfo
}

func (self *Tab) ID() string {
	return self.target.ID
}

func (self *Tab) Disconnect() error {
	return self.rpc.Close()
}

func (self *Tab) Emit(method string, params map[string]interface{}) {
	self.rpc.SynthesizeEvent(RpcMessage{
		ID:     0,
		Method: method,
		Params: params,
	})
}

func (self *Tab) Navigate(url string) (*RpcMessage, error) {
	self.mostRecentInfo = &PageInfo{
		URL:   url,
		State: `initial`,
	}

	self.Emit(`Webfriend.urlChanged`, map[string]interface{}{
		`url`: url,
	})

	result, err := self.browser.Tab().RPC(`Page`, `navigate`, map[string]interface{}{
		`url`: url,
	})

	if err == nil {
		r := result.R()

		self.mostRecentInfo.loaderId = r.String(`loaderId`)
		self.mostRecentInfo.frameId = r.String(`frameId`)
	}

	return result, err
}

func (self *Tab) StartScreencast(quality int, width int, height int) error {
	self.castlock.Lock()
	defer self.castlock.Unlock()

	if self.screencasting {
		return nil
	} else {
		self.screencasting = true
	}

	return self.AsyncRPC(`Page`, `startScreencast`, map[string]interface{}{
		`format`:    `png`,
		`quality`:   int(mathutil.Clamp(float64(quality), 0, 100)),
		`maxWidth`:  width,
		`maxHeight`: height,
	})
}

func (self *Tab) IsScreencasting() bool {
	self.castlock.Lock()
	defer self.castlock.Unlock()

	return self.screencasting
}

func (self *Tab) GetMostRecentFrame() (int64, []byte, int, int) {
	self.castlock.Lock()
	defer self.castlock.Unlock()

	if self.screencasting {
		if d := self.mostRecentDimensions; len(d) == 2 {
			return self.mostRecentFrameId, self.mostRecentFrame, d[0], d[1]
		}
	}

	return 0, nil, 0, 0
}

func (self *Tab) ResetMostRecentFrame() {
	self.mostRecentFrameId = 0
	self.mostRecentFrame = nil
}

func (self *Tab) StopScreencast() error {
	self.castlock.Lock()
	defer self.castlock.Unlock()

	if self.screencasting {
		self.screencasting = false
		return self.AsyncRPC(`Page`, `stopScreencast`, nil)
	} else {
		return nil
	}
}

func (self *Tab) connect() error {
	if conn, err := NewRPC(self.target.WebSocketDebuggerURL); err == nil {
		self.rpc = conn

		return self.setupEvents()
	} else {
		return err
	}
}

func (self *Tab) CreateAccumulator(filter string) (*eventAccumulator, error) {
	if pattern, err := glob.Compile(filter); err == nil {
		acc := &eventAccumulator{
			id:     stringutil.UUID().String(),
			tab:    self,
			filter: pattern,
			Events: make([]*Event, 0),
		}

		self.accumulators.Store(acc.id, acc)
		return acc, nil
	} else {
		return nil, err
	}
}

func (self *Tab) AsyncRPC(module string, method string, args map[string]interface{}) error {
	return self.rpc.CallAsync(
		fmt.Sprintf("%s.%s", module, method),
		args,
	)
}

func (self *Tab) RPC(module string, method string, args map[string]interface{}) (*RpcMessage, error) {
	if reply, err := self.rpc.Call(fmt.Sprintf("%s.%s", module, method), args, DefaultReplyTimeout); err == nil {
		return reply, nil
	} else {
		return nil, err
	}
}

func (self *Tab) setupEvents() error {
	// do this before any events will be emitted, otherwise the event loop will block
	go self.startEventReceiver()

	// setup internal event handlers
	self.registerInternalEvents()

	if err := self.rpc.CallAsync(`Console.enable`, nil); err != nil {
		return err
	}

	if err := self.rpc.CallAsync(`Page.enable`, nil); err != nil {
		return err
	}

	if err := self.rpc.CallAsync(`DOM.enable`, nil); err != nil {
		return err
	}

	if err := self.rpc.CallAsync(`DOMSnapshot.enable`, nil); err != nil {
		return err
	}

	if err := self.rpc.CallAsync(`Network.enable`, nil); err != nil {
		return err
	}

	if err := self.rpc.CallAsync(`Overlay.enable`, nil); err != nil {
		return err
	}

	return nil
}

func (self *Tab) startEventReceiver() {
	for message := range self.rpc.Messages() {
		event := eventFromRpcResponse(message)

		if name := event.Name; name != `` {
			log.Debugf("[event] %v", name)
		}

		// dispatch events to waiters
		self.waiters.Range(func(_ interface{}, waiterI interface{}) bool {
			go func() {
				if waiter, ok := waiterI.(*EventWaiter); ok {
					if waiter.Match(event) {
						waiter.Events <- event
					}
				}
			}()

			return true
		})

		// add events to matching event accumulators
		self.accumulators.Range(func(_ interface{}, accI interface{}) bool {
			if accumulator, ok := accI.(*eventAccumulator); ok {
				accumulator.AppendIfMatch(event)
			}

			return true
		})
	}

	self.browser.Stop()
}

func (self *Tab) registerInternalEvents() {
	self.RegisterEventHandler(consoleEvents, func(event *Event) {
		switch event.Name {
		case `Console.messageAdded`:
			var level log.Level

			switch event.Params.String(`message.level`) {
			case `warning`:
				level = log.WARNING
			case `error`:
				level = log.ERROR
			case `debug`:
				level = log.DEBUG
			case `info`:
				level = log.INFO
			default:
				level = log.NOTICE
			}

			log.Logf(
				level,
				"[CONSOLE %s] %v",
				event.Params.String(`message.source`),
				event.Params.String(`message.text`),
			)
		}
	})

	self.RegisterEventHandler(netTrackingEvents, func(event *Event) {
		requestId := event.P().String(`requestId`)

		request := &NetworkRequest{
			ID: requestId,
		}

		if requestI, ok := self.networkRequests.Load(requestId); ok {
			request = requestI.(*NetworkRequest)
		}

		switch event.Name {
		case `Network.requestWillBeSent`:
			request.Request = event
			log.Debugf("[tab] NetworkRequest[%v] started", requestId)

		case `Network.responseReceived`:
			request.Response = event
			log.Debugf("[tab] NetworkRequest[%v] response received (HTTP %d)", requestId, event.P().Int(`response.status`))

		case `Network.loadingFailed`:
			request.Failure = event
			log.Debugf("[tab] NetworkRequest[%v] load failed: %s", requestId, event.P().String(`errorText`))

		case `Network.loadingFinished`:
			request.Completion = event
			log.Debugf("[tab] NetworkRequest[%v] load completed (%d bytes)", requestId, event.P().Int(`encodedDataLength`))
		}

		self.networkRequests.Store(requestId, request)
	})

	// monitor page URL and load state
	self.RegisterEventHandler(`Network.requestWillBeSent`, func(event *Event) {
		if p := event.P(); p != nil {
			var oldUrl string
			var newUrl string

			if self.mostRecentInfo != nil {
				oldUrl = self.mostRecentInfo.URL
				newUrl = p.String(`documentURL`)
			}

			if newUrl == `` || newUrl == oldUrl {
				return
			}

			if p.String(`frameId`) != self.mostRecentInfo.frameId {
				return
			}

			self.Emit(`Webfriend.urlChanged`, map[string]interface{}{
				`oldUrl`: oldUrl,
				`url`:    newUrl,
			})
		}
	})

	self.RegisterEventHandler(`Network.requestIntercepted`, func(event *Event) {
		url := event.P().String(`request.url`)
		id := event.P().String(`interceptionId`)
		interceptResponse := &NetworkInterceptResponse{}

		self.netIntercepts.Range(func(key interface{}, value interface{}) bool {
			if requestPattern, ok := key.(*NetworkRequestPattern); ok {
				if fn, ok := value.(NetworkInterceptFunc); ok && fn != nil {
					if requestPattern.Pattern == nil {
						if pattern, err := glob.Compile(requestPattern.URL); err == nil {
							requestPattern.Pattern = pattern
						} else {
							return true
						}
					}

					if requestPattern.Pattern.Match(url) {
						log.Debugf("Intercepted %v: %v", id, url)

						if response := fn(self, requestPattern, event); response != nil {
							if response.Autoremove {
								self.netIntercepts.Delete(key)
							}

							interceptResponse = response
						}
					}
				}
			}

			return true
		})

		// if we receive this event, we HAVE to respond to it
		if err := self.AsyncRPC(`Network`, `continueInterceptedRequest`, interceptResponse.ToMap(id)); err != nil {
			log.Errorf("Failed to complete interception: %v", err)
		}
	})

	self.RegisterEventHandler(`Page.screencastFrame`, func(event *Event) {
		defer self.RPC(`Page`, `screencastFrameAck`, map[string]interface{}{
			`sessionId`: event.Params.Int(`sessionId`),
		})

		if data := event.Params.String(`data`); data != `` {
			if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
				self.mostRecentFrame = decoded
				self.mostRecentFrameId = time.Now().UnixNano()
				self.mostRecentDimensions = []int{
					int(event.Params.Int(`metadata.deviceWidth`)),
					int(event.Params.Int(`metadata.deviceHeight`)),
				}
			} else {
				log.Warningf("[rpc] Failed to decode screencast frame: %v", err)
			}
		}
	})
}

func (self *Tab) AddNetworkIntercept(urlPattern string, waitForHeaders bool, fn NetworkInterceptFunc) error {
	requestPattern := &NetworkRequestPattern{}

	if urlPattern == `` {
		requestPattern.URL = `*`
	} else {
		requestPattern.URL = urlPattern
	}

	if waitForHeaders {
		requestPattern.InterceptionStage = `HeadersReceived`
	} else {
		requestPattern.InterceptionStage = `Request`
	}

	patterns := []map[string]interface{}{
		requestPattern.ToMap(),
	}

	self.netIntercepts.Range(func(key interface{}, _ interface{}) bool {
		if rp, ok := key.(*NetworkRequestPattern); ok {
			patterns = append(patterns, rp.ToMap())
		}

		return true
	})

	if err := self.AsyncRPC(`Network`, `setRequestInterception`, map[string]interface{}{
		`patterns`: patterns,
	}); err == nil {
		self.netIntercepts.Store(requestPattern, fn)
		return nil
	} else {
		return err
	}
}

func (self *Tab) ClearNetworkIntercepts() error {
	self.netIntercepts = sync.Map{}

	return self.AsyncRPC(`Network`, `setRequestInterception`, map[string]interface{}{
		`patterns`: []interface{}{},
	})
}

func (self *Tab) ResetNetworkRequests() {
	self.networkRequests = sync.Map{}
}

func (self *Tab) GetLoaderRequest(id string) (netreq *NetworkRequest) {
	self.networkRequests.Range(func(key interface{}, value interface{}) bool {
		if key.(string) == id {
			netreq = value.(*NetworkRequest)
			return false
		}

		return true
	})

	return
}

func (self *Tab) CreateEventWaiter(eventGlob string) (*EventWaiter, error) {
	if waiter, err := NewEventWaiter(self, eventGlob); err == nil {
		self.waiters.Store(waiter.id, waiter)
		return waiter, nil
	} else {
		return nil, err
	}
}

func (self *Tab) RemoveWaiter(id string) {
	self.waiters.Delete(id)
}

func (self *Tab) WaitFor(eventGlob string, timeout time.Duration) (*Event, error) {
	if waiter, err := self.CreateEventWaiter(eventGlob); err == nil {
		defer self.RemoveWaiter(waiter.id)

		log.Debugf("[rpc] Waiting for %v for up to %v", eventGlob, timeout)
		return waiter.Wait(timeout)
	} else {
		return nil, err
	}
}

func (self *Tab) RegisterEventHandler(eventGlob string, callback EventCallbackFunc) (string, error) {
	if waiter, err := self.CreateEventWaiter(eventGlob); err == nil {
		log.Debugf("[rpc] Registered persistent handler for %v", eventGlob)

		go func() {
			for event := range waiter.Events {
				callback(event)
			}
		}()

		return waiter.id, nil
	} else {
		return ``, err
	}
}

// Recursively retrieve values from the RPC, turning a Javascript response into a concrete
// native type. Expects to be given a maputil.Map representing a single Runtime.RemoteObject
func (self *Tab) getJavascriptResponse(result *maputil.Map) (interface{}, error) {
	if oid := result.String(`objectId`); oid != `` {
		if rv, err := self.RPC(`Runtime`, `getProperties`, map[string]interface{}{
			`objectId`:               oid,
			`ownProperties`:          true,
			`accessorPropertiesOnly`: false,
		}); err == nil {
			rType := result.String(`type`)
			rSubtype := result.String(`subtype`, rType)

			// handles compound types
			switch rSubtype {
			case `array`:
				out := make([]interface{}, 0)

				// go through the results and populate the output array
				for _, elem := range maputil.M(rv.Result).Slice(`result`) {
					if elemM := maputil.M(elem); elemM.Bool(`enumerable`) {
						var elemV interface{}
						var err error

						if concreteValue := elemM.Get(`value`); !concreteValue.IsNil() {
							elemV, err = self.getJavascriptResponse(maputil.M(concreteValue))
						} else if objectId := elemM.String(`objectId`); objectId != `` {
							elemV, err = self.getJavascriptResponse(elemM)
						} else {
							err = fmt.Errorf("Do not know how to process response")
						}

						if err == nil {
							if elemV != skipItem {
								out = append(out, elemV)
							}
						} else {
							return nil, err
						}
					}
				}

				return out, nil

			case `object`:
				out := make(map[string]interface{})

				// go through the results and populate the output map
				for _, elem := range maputil.M(rv.Result).Slice(`result`) {
					elemM := maputil.M(elem)
					valueM := maputil.M(elemM.Get(`value`))

					if elemM.Bool(`enumerable`) {
						if key := elemM.String(`name`); key != `` {
							if elemV, err := self.getJavascriptResponse(valueM); err == nil {
								if elemV != skipItem {
									out[key] = elemV
								}
							} else {
								return nil, fmt.Errorf("key %s: %v", key, err)
							}
						}
					}
				}

				return out, nil

			case `node`:
				if node, err := self.RPC(`DOM`, `describeNode`, map[string]interface{}{
					`objectId`: result.String(`objectId`),
					`depth`:    2,
				}); err == nil {
					return self.getElementFromResult(
						maputil.M(node.R().Get(`node`)),
					), nil
				} else {
					return nil, err
				}

			default:
				log.Dumpf("Unhandled Value: %s", result.Value())
				return nil, fmt.Errorf("Unhandled Javascript type %q", rSubtype)
			}
		} else {
			return nil, err
		}
	} else {
		return result.Get(`value`), nil
	}
}

func (self *Tab) resolveNode(backendNodeId int64) (string, error) {
	if node, err := self.RPC(`DOM`, `resolveNode`, map[string]interface{}{
		`backendNodeId`: backendNodeId,
	}); err == nil {
		if objectId := node.R().String(`object.objectId`); objectId != `` {
			return objectId, nil
		} else {
			return ``, fmt.Errorf("Unable to retrieve RemoteObjectId")
		}
	} else {
		return ``, err
	}
}

func (self *Tab) releaseObjectGroup(gid string) error {
	_, err := self.RPC(`Runtime`, `releaseObjectGroup`, map[string]interface{}{
		`objectGroup`: gid,
	})

	return err
}

func (self *Tab) ElementPosition(element *dom.Element) (dom.Dimensions, error) {
	if result, err := self.EvaluateOn(element, `return Object.assign({}, this.getBoundingClientRect().toJSON())`); err == nil {
		dimensions := maputil.M(result)

		return dom.Dimensions{
			Width:  int(dimensions.Int(`width`)),
			Height: int(dimensions.Int(`height`)),
			Top:    int(dimensions.Int(`top`)),
			Left:   int(dimensions.Int(`left`)),
			Bottom: int(dimensions.Int(`bottom`)),
			Right:  int(dimensions.Int(`right`)),
		}, nil
	} else {
		return dom.Dimensions{}, err
	}
}

func (self *Tab) PageSize() (float64, float64, error) {
	if result, err := self.Evaluate(`return [document.documentElement.scrollWidth, document.documentElement.scrollHeight]`); err == nil {
		if sz := typeutil.V(result).Slice(); len(sz) == 2 {
			return sz[0].Float(), sz[1].Float(), nil
		} else {
			return 0, 0, fmt.Errorf("Invalid response while retrieving page dimensions")
		}
	} else {
		return 0, 0, err
	}
}

func (self *Tab) ElementQuery(selector dom.Selector, parent *dom.Selector) ([]*dom.Element, error) {
	psel := `document`

	if parent != nil {
		if parent.IsNone() {
			return make([]*dom.Element, 0), nil
		} else {
			psel = fmt.Sprintf("document.querySelector(%q)", parent.String())
		}
	}

	if results, err := self.Evaluate(fmt.Sprintf(
		"return Array.from(%s.querySelectorAll(%q))",
		psel,
		selector,
	)); err == nil {
		elements := make([]*dom.Element, 0)

		for _, result := range sliceutil.Sliceify(results) {
			if element, ok := result.(*dom.Element); ok {
				elements = append(elements, element)
			}
		}

		return elements, nil
	} else {
		return nil, err
	}
}

func (self *Tab) Evaluate(stmt string, exposed ...string) (interface{}, error) {
	return self.evaluate(``, stmt, exposed)
}

func (self *Tab) EvaluateOn(element *dom.Element, stmt string, exposed ...string) (interface{}, error) {
	if element == nil {
		return nil, fmt.Errorf("Cannot call function on unspecified element")
	} else if element.ID == `` {
		return nil, fmt.Errorf("Cannot call function on element without an ID")
	} else {
		return self.evaluate(element.ID, stmt, exposed)
	}
}

func (self *Tab) evaluate(remoteObjectId string, stmt string, exposed []string) (interface{}, error) {
	callGroupId := stringutil.UUID().Base58()

	var rv *RpcMessage
	var err error

	exposed = sliceutil.MapString(exposed, func(i int, value string) string {
		return strings.TrimPrefix(value, `$`)
	})

	// oid <= 0 means call globally
	if remoteObjectId == `` {
		rv, err = self.RPC(`Runtime`, `evaluate`, map[string]interface{}{
			`expression`: fmt.Sprintf(
				"%s;\nvar fn_%s = function(){ %s }.bind(webfriend); fn_%s()",
				self.getEvalPrescript(exposed),
				callGroupId,
				stmt,
				callGroupId,
			),
			`returnByValue`: false,
			`awaitPromise`:  false,
			`objectGroup`:   callGroupId,
		})
	} else {
		rv, err = self.RPC(`Runtime`, `callFunctionOn`, map[string]interface{}{
			`objectId`: remoteObjectId,
			`functionDeclaration`: fmt.Sprintf(
				"function(){ %s; %s }",
				self.getEvalPrescript(exposed),
				stmt,
			),
			`returnByValue`: false,
			`userGesture`:   true,
			`awaitPromise`:  false,
			`objectGroup`:   callGroupId,
		})
	}

	if err == nil {
		defer self.releaseObjectGroup(callGroupId)
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
			return self.getJavascriptResponse(maputil.M(out.Get(`result`)))
		} else if returnValue := out.Get(`result.value`).Value; returnValue != nil {
			return returnValue, nil
		} else {
			return nil, nil
		}
	} else if strings.Contains(err.Error(), `Could not find object with given id`) {
		return nil, fmt.Errorf("Element '%v' could not be found", remoteObjectId)
	} else {
		return nil, err
	}
}

func (self *Tab) getEvalPrescript(exposed []string) string {
	if scopeable := self.browser.scopeable; scopeable != nil {
		if scope := scopeable.Scope(); scope != nil {
			out := `var webfriend = `
			jsData := make(map[string]interface{})

			maputil.Walk(scope.Data(), func(value interface{}, path []string, isLeaf bool) error {
				key := strings.Join(path, `.`)

				if sliceutil.ContainsString(exposed, key) {
					maputil.DeepSet(jsData, path, value)
					return maputil.SkipDescendants
				} else {
					return nil
				}
			})

			if data, err := json.Marshal(jsData); err == nil {
				out += string(data)
				return out
			}
		}
	}

	return ``
}

func (self *Tab) getElementFromResult(node *maputil.Map) *dom.Element {
	if remoteObjectId, err := self.resolveNode(node.Int(`backendNodeId`)); err == nil {
		var element *dom.Element
		var children = node.Slice(`children`)

		// load the various properties from the given node map into a new elements
		pair := node.String(`localName`, strings.ToLower(node.String(`nodeName`)))
		ns, name := stringutil.SplitPairRightTrailing(pair, `:`)

		element = &dom.Element{
			ID:         remoteObjectId,
			Namespace:  ns,
			Name:       name,
			Attributes: make(map[string]interface{}),
		}

		for _, pair := range sliceutil.Chunks(node.Slice(`attributes`), 2) {
			element.Attributes[typeutil.String(pair[0])] = typeutil.Auto(pair[1])
		}

		switch len(children) {
		case 0:
			element.Text = node.String(`nodeValue`)
		default:
			for _, child := range children {
				childM := maputil.M(child)

				switch childM.Int(`nodeType`) {
				case 1:
					if subel := self.getElementFromResult(childM); subel != nil {
						element.Text += subel.Text
					}
				case 3:
					element.Text += childM.String(`nodeValue`)
				}
			}
		}

		return element
	} else {
		log.Warningf("Received invalid node: %v", err)
		return nil
	}
}
