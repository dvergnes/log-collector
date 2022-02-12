// Copyright (c) 2022 Denis Vergnes
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package processor

import "io"

type limitEventProcessor struct {
	delegate EventProcessor

	nbEventProcessed uint
	limit            uint
}

// WithLimit decorates an EventProcessor to apply a limit on the maximum number of events that it can return
func WithLimit(processor EventProcessor, limit uint) EventProcessor {
	return &limitEventProcessor{
		delegate: processor,
		limit:    limit,
	}
}

// Next implements EventProcessor contract
func (l *limitEventProcessor) Next() (string, error) {
	if l.nbEventProcessed >= l.limit {
		return "", io.EOF
	}
	next, err := l.delegate.Next()
	if err != nil {
		return next, err
	}
	l.nbEventProcessed++
	return next, err
}

// EventFilter verifies that an event matches a condition. It returns true if the event passes the check
type EventFilter func(string) bool

// WithFilter decorates an EventProcessor to apply a predicate that must validate the event to return it
func WithFilter(processor EventProcessor, filter EventFilter) EventProcessor {
	return &filterEventProcessor{
		delegate: processor,
		predicate: filter,
	}
}

type filterEventProcessor struct {
	delegate  EventProcessor
	predicate EventFilter
}

// Next implements EventProcessor contract
func (f *filterEventProcessor) Next() (string, error) {
	for {
		next, err := f.delegate.Next()
		if err != nil {
			return next, err
		}
		if f.predicate(next) {
			return next, err
		}
	}
}
