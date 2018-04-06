package browser

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/fatih/structs"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/utils"
	"github.com/gobwas/glob"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/rpcc"
)

var netTrackingEvents = `Network.{requestWillBeSent,responseReceived,loadingFailed}`
var domTrackingEvents = `DOM.*`

type TabID string

type Tab struct {
	browser         *Browser
	target          *devtool.Target
	rpcConn         *rpcc.Conn
	rpc             *cdp.Client
	events          chan *Event
	waiters         sync.Map
	networkRequests []*Event
	netreqLock      sync.Mutex
	accumulators    sync.Map
}

func newTabFromTarget(browser *Browser, target *devtool.Target) (*Tab, error) {
	tab := &Tab{
		browser:         browser,
		target:          target,
		events:          make(chan *Event),
		networkRequests: make([]*Event, 0),
	}

	return tab, tab.connect()
}

func (self *Tab) ID() string {
	return self.target.ID
}

func (self *Tab) Disconnect() error {
	return self.rpcConn.Close()
}

func (self *Tab) DOM() *Document {
	return NewDocument(self, nil)
}

func (self *Tab) connect() error {
	if conn, err := rpcc.Dial(self.target.WebSocketDebuggerURL, rpcc.WithCodec(func(conn io.ReadWriter) rpcc.Codec {
		return newRpccStreamIntercept(conn, self.events)
	})); err == nil {
		self.rpcConn = conn
		self.rpc = cdp.NewClient(self.rpcConn)

		return self.setupEvents()
	} else {
		return err
	}
}

func (self *Tab) getModule(module string) (*structs.Field, bool) {
	return structs.New(self.rpc).FieldOk(module)
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

func (self *Tab) RPC(module string, method string, args map[string]interface{}) (map[string]interface{}, error) {
	if mod, ok := self.getModule(module); ok {
		if fn, err := utils.GetFunctionByName(mod.Value(), method); err == nil {
			ctx := self.browser.ctx()
			arguments := make([]reflect.Value, fn.Type().NumIn())

			for i := 0; i < len(arguments); i++ {
				argT := fn.Type().In(i)

				switch i {
				case 0:
					arguments[i] = reflect.ValueOf(ctx)
					continue
				case 1:
					if argT.Kind() == reflect.Ptr {
						argT = argT.Elem()
					}

					if argT.Kind() != reflect.Struct {
						return nil, fmt.Errorf("Expected second argument to be a configuration struct, got %v", argT)
					}

					argStruct := reflect.New(argT)

					if len(args) > 0 {
						fieldNames := make([]string, 0)

						for i := 0; i < argT.NumField(); i++ {
							fieldNames = append(fieldNames, argT.Field(i).Name)
						}

						for key, _ := range args {
							if sliceutil.ContainsString(fieldNames, key) {
								continue
							} else if _, err := utils.GetFunctionByName(argStruct, fmt.Sprintf("Set%v", key)); err == nil {
								continue
							} else {
								return nil, fmt.Errorf("Could not find field or setter function for parameter '%v'", key)
							}
						}

						// for each field in the argstruct
						for i := 0; i < argT.NumField(); i++ {
							field := argStruct.Elem().Field(i)
							keyName := argT.Field(i).Name

							// if there is a key with this name in the arg map
							if value, ok := args[keyName]; ok {
								valueV := reflect.ValueOf(value)
								didSet := false

								if field.CanSet() {
									// convert if necessary
									if !valueV.Type().AssignableTo(field.Type()) {
										if valueV.Type().ConvertibleTo(field.Type()) {
											valueV = valueV.Convert(field.Type())
										}
									}

									if valueV.Type().AssignableTo(field.Type()) {
										field.Set(valueV)
										didSet = true
									}
								}

								if !didSet {
									// if we couldn't set the value, look for a setter function
									// with the format Set<FieldName>().
									fnName := fmt.Sprintf("Set%v", keyName)

									if setter, err := utils.GetFunctionByName(argStruct, fnName); err == nil {
										setterT := setter.Type()

										// valid setters must take one argument
										if setterT.NumIn() == 1 {
											setterArgT := setterT.In(0)
											valueV := reflect.ValueOf(value)

											// if we can't assign the value directly, try to convert it
											if !valueV.Type().AssignableTo(setterArgT) {
												if valueV.Type().ConvertibleTo(setterArgT) {
													valueV = valueV.Convert(setterArgT)
												} else {
													return nil, fmt.Errorf("%v: cannot convert %T to %v", fnName, value, setterArgT)
												}
											}

											// call the setter function

											setter.Call([]reflect.Value{
												valueV,
											})
										} else {
											return nil, fmt.Errorf("Setter function %v must take one argument", fnName)
										}
									} else {
										return nil, fmt.Errorf("Could not set %v, and no setter function was found", keyName)
									}
								}
							}
						}
					}

					arguments[i] = argStruct
				default:
					return nil, fmt.Errorf("Unsupported number of arguments; expected [0,1], got %d", len(arguments))
				}
			}

			// log.Debugf("CALL %v.%v", module, method)
			results := fn.Call(arguments)

			switch len(results) {
			case 1:
				return nil, fnOutputVarToError(results[0])

			case 2:
				if err := fnOutputVarToError(results[1]); err == nil {
					rv := results[0].Interface()
					output := structs.Map(rv)

					for key, value := range output {
						output[key] = stringutil.Autotype(typeutil.ResolveValue(value))
					}

					return output, nil
				} else {
					return nil, err
				}

			default:
				return nil, fmt.Errorf("Expected [1,2] result values, got %d", len(results))
			}
		} else {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("No such RPC module '%v'", module)
	}
}

func (self *Tab) setupEvents() error {
	// do this before any events will be emitted, otherwise the event loop will block
	go self.startEventReceiver()

	// setup internal event handlers
	self.registerInternalEvents()

	if err := self.rpc.Console.Enable(self.browser.ctx()); err != nil {
		return err
	}

	if err := self.rpc.Page.Enable(self.browser.ctx()); err != nil {
		return err
	}

	if err := self.rpc.DOM.Enable(self.browser.ctx()); err != nil {
		return err
	}

	if err := self.rpc.Network.Enable(self.browser.ctx(), nil); err != nil {
		return err
	}

	return nil
}

func (self *Tab) startEventReceiver() {
	for event := range self.events {
		// log.Debugf("[event] %v", event)

		// dispatch events to waiters
		self.waiters.Range(func(_ interface{}, waiterI interface{}) bool {
			if waiter, ok := waiterI.(*eventWaiter); ok {
				if waiter.Match(event) {
					waiter.Events <- event
					return false
				}
			}

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
	self.RegisterEventHandler(netTrackingEvents, func(event *Event) {
		self.netreqLock.Lock()
		defer self.netreqLock.Unlock()
		self.networkRequests = append(self.networkRequests, event)
	})

	self.RegisterEventHandler(domTrackingEvents, func(event *Event) {
		// log.Debugf("evt %v", event)

		// switch event.Name {
		// case `DOM.documentUpdated`:
		// 	self.DOM().Reset()
		// }
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
		if fmt.Sprintf("%v", maputil.Get(event.Params, `requestId`)) == id {
			switch event.Name {
			case `Network.requestWillBeSent`:
				request = event
			case `Network.responseReceived`:
				response = event
				break
			case `Network.loadingFailed`:
				reqerr = event
				break
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

func (self *Tab) WaitFor(eventGlob string, timeout time.Duration) (*Event, error) {
	if waiter, err := newEventWaiter(eventGlob); err == nil {
		self.waiters.Store(waiter.ID, waiter)
		defer self.waiters.Delete(waiter.ID)

		log.Debugf("[rpc] Waiting for %v for up to %v", eventGlob, timeout)

		select {
		case event := <-waiter.Events:
			return event, nil
		case <-time.After(timeout):
			return nil, fmt.Errorf("timeout")
		}
	} else {
		return nil, err
	}
}

func (self *Tab) RegisterEventHandler(eventGlob string, callback EventCallbackFunc) (string, error) {
	if waiter, err := newEventWaiter(eventGlob); err == nil {
		self.waiters.Store(waiter.ID, waiter)

		log.Debugf("[rpc] Registered persistent handler for %v", eventGlob)

		go func() {
			for event := range waiter.Events {
				callback(event)
			}
		}()

		return waiter.ID, nil
	} else {
		return ``, err
	}
}
