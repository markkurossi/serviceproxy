/*
 * trace.go
 */

package authorizer

import (
	"fmt"
	"io"
	"time"
)

type Trace struct {
	Start  time.Time
	Last   time.Time
	Events []Event
}

type Event struct {
	When  time.Duration
	Label string
}

func NewTrace() *Trace {
	now := time.Now()
	return &Trace{
		Start: now,
		Last:  now,
	}
}

func (t *Trace) Event(label string) {
	now := time.Now()
	t.Events = append(t.Events, Event{
		When:  now.Sub(t.Last),
		Label: label,
	})
	t.Last = now
}

func (t *Trace) End(w io.Writer) {
	t.Event("End")
	for _, e := range t.Events {
		fmt.Fprintf(w, "%s\t%s\n", e.When, e.Label)
	}
}
