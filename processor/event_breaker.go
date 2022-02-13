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

import (
	"bufio"
	"fmt"
)

// ReverseScanLines is similar to bufio.ScanLines except that it scans the bytes from right to left
var ReverseScanLines = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	n := len(data) - 1
	for i := n; i >= 0; i-- {
		c := data[i]
		if c == '\n' {
			// TODO: drop CR
			return advance + 1, data[n+1-advance : n+1], nil
		} else {
			advance++
		}
	}
	if atEOF {
		return advance, data[:n+1], nil
	}
	return 0, nil, nil
}

// EventProcessor defines an iterator that process an event
type EventProcessor interface {
	// Next returns an event as a string. It returns io.EOF if no more event will be returned
	Next() (string, error)
}

// EventBreaker is responsible to identify events in an array of bytes read from a io.Reader
type EventBreaker struct {
	buf    []byte
	length int
	end    int

	reader   TailReader
	splitter bufio.SplitFunc
}

// NewEventBreaker creates a new EventBreaker that reads from the passed TailReader with a buffer of the given bufferSize.
// It generates events by applying the passed bufio.SplitFunc.
func NewEventBreaker(reader TailReader, splitter bufio.SplitFunc, bufferSize int) *EventBreaker {
	return &EventBreaker{
		buf:      make([]byte, bufferSize),
		reader:   reader,
		splitter: splitter,
	}
}

// Next implements EventProcessor contract
func (eb *EventBreaker) Next() (string, error) {
	// search for next event separator
	advance, token, err := eb.nextEvent(false)
	if err != nil {
		return "", err
	}

	// if we cannot make progress, we try to continue to read the reader so that we can find an event boundary
	if advance == 0 {
		// we rewind the reader for the partial event we were reading to be on event boundary
		eb.reader.SeekToEnd(uint32(eb.end))
		if err := eb.fillBuffer(); err != nil {
			return "", err
		}
		// this time we try to find event boundary and return anything that was in the buffer
		advance, token, err = eb.nextEvent(true)
		if err != nil {
			return "", fmt.Errorf("failed to split event %w", err)
		}
	}

	// return the event
	return string(token), nil

}

func (eb *EventBreaker) nextEvent(atEOF bool) (int, []byte, error) {
	for {
		// if buffer is empty, fill the buffer
		if eb.length <= 0 {
			err := eb.fillBuffer()
			if err != nil {
				return 0, nil, err
			}
		}
		advance, token, err := eb.splitter(eb.buf[:eb.end], atEOF)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to split event %w", err)
		}
		// if the splitter cannot make progress we return immediately
		if advance == 0 {
			return advance, token, err
		}
		eb.length -= advance
		eb.end -= advance
		// if the token is empty, we want to continue to process the buffer
		if len(token) != 0 {
			return advance, token, nil
		}
	}

}

func (eb *EventBreaker) fillBuffer() error {
	n, err := eb.reader.Read(eb.buf)
	if err != nil {
		return err
	}
	eb.length = n
	eb.end = n

	return nil
}
