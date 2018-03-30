package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/gobwas/glob"
	"github.com/mafredri/cdp/rpcc"
)

type EventCallbackFunc func(event *Event)

type eventWaiter struct {
	ID      string
	Pattern glob.Glob
	Events  chan *Event
}

func newEventWaiter(eventGlob string) (*eventWaiter, error) {
	if pattern, err := glob.Compile(eventGlob); err == nil {
		return &eventWaiter{
			ID:      stringutil.UUID().String(),
			Pattern: pattern,
			Events:  make(chan *Event),
		}, nil
	} else {
		return nil, err
	}
}

func (self *eventWaiter) Match(event *Event) bool {
	return self.Pattern.Match(event.Name)
}

type Event struct {
	ID        int
	Name      string
	Result    map[string]interface{}
	Params    map[string]interface{}
	Error     error
	Timestamp time.Time
}

func (self *Event) String() string {
	if self.Error != nil {
		return self.Error.Error()
	} else {
		return fmt.Sprintf("%v", self.Name)
	}
}

func (self *Event) Get(key string, fallbacks ...interface{}) interface{} {
	return maputil.DeepGet(self.Params, strings.Split(key, `.`), fallbacks...)
}

func (self *Event) GetString(key string, fallbacks ...string) string {
	return fmt.Sprintf("%v", self.Get(key, ``))
}

func (self *Event) GetBool(key string) bool {
	if v := self.Get(key); stringutil.IsBoolean(v) {
		return stringutil.MustBool(v)
	} else {
		return false
	}
}

func (self *Event) GetInt(key string, fallbacks ...int64) int64 {
	if v := self.Get(key); stringutil.IsInteger(v) {
		return stringutil.MustInteger(v)
	} else if len(fallbacks) > 0 {
		return fallbacks[0]
	} else {
		return -1
	}
}

func eventFromRpcResponse(resp *rpcc.Response) (*Event, error) {
	event := &Event{
		ID:        int(resp.ID),
		Name:      resp.Method,
		Result:    make(map[string]interface{}),
		Params:    make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	if len(resp.Result) > 0 {
		if err := json.Unmarshal(resp.Result, &event.Result); err != nil {
			return nil, err
		}
	}

	if len(resp.Args) > 0 {
		if err := json.Unmarshal(resp.Args, &event.Params); err != nil {
			return nil, err
		}
	}

	if resp.Error != nil {
		event.Error = fmt.Errorf(
			"Error %d: %v",
			resp.Error.Code,
			resp.Error.Message,
		)
	}

	return event, nil
}

type rpccStreamIntercept struct {
	conn   io.ReadWriter
	events chan *Event
}

func newRpccStreamIntercept(conn io.ReadWriter, events chan *Event) *rpccStreamIntercept {
	return &rpccStreamIntercept{
		conn:   conn,
		events: events,
	}
}

func (self *rpccStreamIntercept) WriteRequest(req *rpcc.Request) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		return err
	}

	if _, err := self.conn.Write(buf.Bytes()); err != nil {
		return err
	}

	// log.Debugf("[proto] WROTE: %v", buf.String())

	return nil
}

func (self *rpccStreamIntercept) ReadResponse(resp *rpcc.Response) error {
	var buf bytes.Buffer

	if err := json.NewDecoder(io.TeeReader(self.conn, &buf)).Decode(resp); err != nil {
		return err
	}

	// log.Debugf("[proto] READ: %v", buf.String())

	if event, err := eventFromRpcResponse(resp); err == nil {
		self.events <- event
	} else {
		log.Warningf("event error: %v", err)
	}

	return nil
}
