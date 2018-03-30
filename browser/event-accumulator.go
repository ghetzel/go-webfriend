package browser

import "github.com/gobwas/glob"

type eventAccumulator struct {
	id      string
	tab     *Tab
	filter  glob.Glob
	stopped bool
	Events  []*Event
}

func (self *eventAccumulator) AppendIfMatch(event *Event) bool {
	if self.stopped {
		return false
	}

	if self.filter.Match(event.Name) {
		self.Events = append(self.Events, event)
		return true
	}

	return false
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
