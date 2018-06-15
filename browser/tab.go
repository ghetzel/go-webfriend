package browser

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/mathutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/gobwas/glob"
	"github.com/mafredri/cdp/devtool"
)

var netTrackingEvents = `Network.{requestWillBeSent,responseReceived,loadingFailed}`
var domTrackingEvents = `DOM.*`
var consoleEvents = `Console.messageAdded`

type TabID string

type PageInfo struct {
	URL      string `json:"url"`
	State    string `json:"state"`
	loaderId string
	frameId  string
}

type Tab struct {
	browser              *Browser
	target               *devtool.Target
	rpc                  *RPC
	events               chan *Event
	waiters              sync.Map
	networkRequests      []*Event
	netreqLock           sync.Mutex
	accumulators         sync.Map
	currentDocument      *Document
	mostRecentFrameId    int64
	mostRecentFrame      []byte
	mostRecentDimensions []int
	screencasting        bool
	castlock             sync.Mutex
	mostRecentInfo       *PageInfo
}

func newTabFromTarget(browser *Browser, target *devtool.Target) (*Tab, error) {
	tab := &Tab{
		browser:         browser,
		target:          target,
		events:          make(chan *Event),
		networkRequests: make([]*Event, 0),
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

func (self *Tab) DOM() *Document {
	if self.currentDocument == nil {
		self.currentDocument = NewDocument(self, nil)
	}

	return self.currentDocument
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
		log.Debugf("[event] %v", event)

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
		self.netreqLock.Lock()
		defer self.netreqLock.Unlock()
		self.networkRequests = append(self.networkRequests, event)
	})

	self.RegisterEventHandler(domTrackingEvents, func(event *Event) {
		// log.Debugf("evt %v", event)
		dom := self.DOM()

		switch event.Name {
		case `DOM.documentUpdated`:
			dom.Reset()
		case `DOM.setChildNodes`:
			for _, node := range event.Params.Slice(`nodes`) {
				dom.addElementFromResult(
					maputil.M(node),
				)
			}

		case `DOM.attributeModified`, `DOM.attributeRemoved`:
			nid := int(event.Params.Int(`nodeId`))

			if element, ok := dom.Element(nid); ok {
				element.RefreshAttributes()
			} else {
				log.Warningf("Got attribute update event for unknown node %d", nid)
			}
		}
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

func (self *Tab) ResetNetworkRequests() {
	self.netreqLock.Lock()
	defer self.netreqLock.Unlock()
	self.networkRequests = nil
}

func (self *Tab) GetLoaderRequest(id string) (request *Event, response *Event, reqerr *Event) {
	self.netreqLock.Lock()
	defer self.netreqLock.Unlock()

	for _, event := range self.networkRequests {
		reqId := event.Params.String(`loaderId`, event.Params.String(`frameId`))

		if reqId == id {
			switch event.Name {
			case `Network.requestWillBeSent`:
				if request == nil {
					request = event
				}
			case `Network.responseReceived`:
				if response == nil {
					response = event
					break
				}
			case `Network.loadingFailed`:
				if reqerr == nil {
					reqerr = event
					break
				}
			}
		}
	}

	// go through requests AGAIN, setting details that weren't found in the initital pass
	for _, event := range self.networkRequests {
		switch event.Name {
		case `Network.requestWillBeSent`:
			if request == nil {
				request = event
			}
		case `Network.responseReceived`:
			if response == nil {
				response = event
				return
			}
		case `Network.loadingFailed`:
			if reqerr == nil {
				reqerr = event
				return
			}
		}
	}

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
			remoteType := result.String(`subtype`, result.String(`type`))

			// handles compound types
			switch remoteType {
			case `array`:
				out := make([]interface{}, 0)

				// go through the results and populate the output array
				for _, elem := range maputil.M(rv.Result).Slice(`result`) {
					if elemM := maputil.M(elem); elemM.Bool(`enumerable`) {
						if elemV, err := self.getJavascriptResponse(maputil.M(elemM.Get(`value`))); err == nil {
							out = append(out, elemV)
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
								out[key] = elemV
							} else {
								return nil, fmt.Errorf("key %s: %v", key, err)
							}
						}
					}
				}

				return out, nil

			default:
				log.Dumpf("Unhandled Value: %s", result.Value())
				return nil, fmt.Errorf("Unhandled Javascript type %q", remoteType)
			}
		} else {
			return nil, err
		}
	} else {
		return result.Get(`value`), nil
	}
}

func (self *Tab) releaseObjectGroup(gid string) error {
	_, err := self.RPC(`Runtime`, `releaseObjectGroup`, map[string]interface{}{
		`objectGroup`: gid,
	})

	return err
}
