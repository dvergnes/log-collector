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

type EventBreaker struct {
	buf    []byte
	offset int
	length int

	reader   TailReader
	splitter bufio.SplitFunc
}

func NewEventBreaker(reader TailReader, splitter bufio.SplitFunc, capacity int) *EventBreaker {
	return &EventBreaker{
		buf: make([]byte, capacity),
		reader:   reader,
		splitter: splitter,
	}
}

func (eb *EventBreaker) Next() (string, error) {
	// if buffer is empty, fill the buffer
	if eb.offset >= eb.length {
		_, err := eb.fillBuffer()
		if err != nil {
			return "", err
		}
	}

	// search for next event separator
	advance, token, err := eb.splitter(eb.buf[eb.offset:eb.length], true)
	if err != nil {
		return "", fmt.Errorf("failed to split event %w", err)
	}
	eb.offset += advance

	// return the event
	return string(token), nil

}

func (eb *EventBreaker) fillBuffer() (int, error) {
	eb.offset = 0
	n, err := eb.reader.Read(eb.buf)
	if err != nil {
		return n, err
	}

	// search for the line separator
	advance, token, err := eb.splitter(eb.buf[:n], true)
	if err != nil {
		return 0, fmt.Errorf("failed to find first event separator %w", err)
	}

	if len(token) == 0 {
		advance += len(token)
	}

	// rewind reader so that next read will be on boundary
	eb.reader.SeekToEnd(int64(advance))
	eb.offset += advance
	eb.length = n
	return n, nil
}
