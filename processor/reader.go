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
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/afero"
)

// ConcurrentAccessErr indicates that a file has been modified while the tail reader was reading it
var ConcurrentAccessErr = errors.New("file has been modified since the tail reader was opened")

// TailReader reads a file from the end to the start of the file. Once created if the file is modified, the next Read
// call will return an error.
type TailReader interface {
	io.Reader
	io.Closer
	// SeekToEnd updates the offset of the TailReader towards the end of file. It basically rewinds the reader by the given
	// offset.
	SeekToEnd(offset int64)
}

// tailReader implements TailReader interface
type tailReader struct {
	fs afero.Fs

	file    afero.File
	modTime time.Time

	offsetFromEnd int64
}

// NewTailReader creates a tailReader for the given file in parameters
func NewTailReader(fs afero.Fs, name string) (TailReader, error) {
	file, err := fs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %w", err)
	}
	stat, err := readFileStat(file)
	if err != nil {
		return nil, err
	}
	return &tailReader{
		fs:      fs,
		file:    file,
		modTime: stat.ModTime(),
	}, nil
}

func readFileStat(file afero.File) (os.FileInfo, error) {
	stat, err := file.Stat()
	if err != nil {
		return stat, fmt.Errorf("failed to read file metadata %w", err)
	}
	return stat, nil
}

// Read reads a file from the end to the start.
// Each call gets closer to the start. Once the entire file is read io.EOF is returned.
// If the file is modified since the tailReader is created an error is returned on the next call to Read.
// It implements io.Reader interface.
func (tr *tailReader) Read(buf []byte) (int, error) {
	stat, err := readFileStat(tr.file)
	if err != nil {
		return 0, err
	}
	if tr.modTime.UnixNano() != stat.ModTime().UnixNano() {
		return 0, fmt.Errorf("%w", ConcurrentAccessErr)
	}

	size := stat.Size()
	if tr.offsetFromEnd >= size {
		return 0, io.EOF
	}

	length := int64(len(buf))
	offset := tr.offsetFromEnd + length
	if offset > size {
		offset = size
		length = size - tr.offsetFromEnd
	}
	_, err = tr.file.Seek(-offset, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("failed to seek file %w", err)
	}
	n, err := tr.file.Read(buf[:length])
	tr.offsetFromEnd = offset
	return n, err
}

// SeekToEnd updates the offset of the TailReader towards the end of file. It basically rewinds the reader by the given
// offset.
func (tr *tailReader) SeekToEnd(offset int64) {
	tr.offsetFromEnd -= offset
}

// Close closes the reader and the file that it reads. Any subsequent call to Read will return an error.
func (tr *tailReader) Close() error {
	return tr.file.Close()
}
