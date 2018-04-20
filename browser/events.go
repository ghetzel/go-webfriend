package browser

import (
	"fmt"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/gobwas/glob"
)

type EventCallbackFunc func(event *Event)

type EventWaiter struct {
	Pattern glob.Glob
	Events  chan *Event
	id      string
	tab     *Tab
}

func NewEventWaiter(tab *Tab, eventGlob string) (*EventWaiter, error) {
	if pattern, err := glob.Compile(eventGlob); err == nil {
		return &EventWaiter{
			Pattern: pattern,
			Events:  make(chan *Event, 1024),
			id:      stringutil.UUID().String(),
			tab:     tab,
		}, nil
	} else {
		return nil, err
	}
}

func (self *EventWaiter) Match(event *Event) bool {
	return self.Pattern.Match(event.Name)
}

func (self *EventWaiter) Wait(timeout time.Duration) (*Event, error) {
	select {
	case event := <-self.Events:
		log.Debugf("[rpc] Wait over; got %v", event)
		return event, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout")
	}
}

func (self *EventWaiter) Remove() {
	if self.tab != nil {
		self.tab.RemoveWaiter(self.id)
	}
}

type Event struct {
	ID        int
	Name      string
	Result    *maputil.Map
	Params    *maputil.Map
	Error     error
	Timestamp time.Time
}

func (self *Event) String() string {
	if self.Error != nil {
		return self.Error.Error()
	} else {
		return self.Name
	}
}

func eventFromRpcResponse(resp *RpcMessage) *Event {
	var err error

	if len(resp.Error) > 0 {
		eM := maputil.M(resp.Error)
		err = fmt.Errorf("Code %d: %v", eM.Int(`code`), eM.String(`message`))
	}

	return &Event{
		ID:        int(resp.ID),
		Name:      resp.Method,
		Result:    maputil.M(resp.Result),
		Params:    maputil.M(resp.Params),
		Timestamp: time.Now(),
		Error:     err,
	}
}
