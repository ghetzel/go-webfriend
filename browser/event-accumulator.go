package browser

import (
	"sync"

	"github.com/gobwas/glob"
)

type eventAccumulator struct {
	id      string
	tab     *Tab
	filter  glob.Glob
	stopped bool
	Events  []*Event
	evlock  sync.Mutex
}

func (self *eventAccumulator) AppendIfMatch(event *Event) bool {
	if self.stopped {
		return false
	}

	self.evlock.Lock()
	defer self.evlock.Unlock()

	if self.filter.Match(event.Name) {
		self.Events = append(self.Events, event)
		return true
	}

	return false
}

func (self *eventAccumulator) Range() <-chan *Event {
	events := make(chan *Event)

	go func() {
		self.evlock.Lock()
		defer self.evlock.Unlock()

		for _, event := range self.Events {
			events <- event
		}

		self.Events = nil
		close(events)
	}()

	return events
}

func (self *eventAccumulator) Stop() {
	self.stopped = true
	// log.Debugf("[%s] Stopped receiving events on accumulator", self.id)
}

func (self *eventAccumulator) Resume() {
	self.stopped = false
	// log.Debugf("[%s] Resumed receiving events on accumulator", self.id)
}

func (self *eventAccumulator) Destroy() {
	if self.tab != nil {
		self.tab.accumulators.Delete(self.id)
		// log.Debugf("[%s] Removed accumulator", self.id)
	}
}
