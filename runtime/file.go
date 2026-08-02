package runtime

import (
	"bufio"
	"os"
	"sync"
	"unsafe"
)

type FileObject struct {
	File       *os.File
	Path       string
	lineReader *bufio.Reader
	closed     bool
	mu         sync.Mutex
}

func (f *FileObject) bufferedReaderLocked() *bufio.Reader {
	if f == nil {
		panic("expected file")
	}
	if f.closed || f.File == nil {
		panic("expected open file")
	}
	if f.lineReader == nil {
		f.lineReader = bufio.NewReader(f.File)
	}
	return f.lineReader
}

func NewFile(file *os.File) Value {
	if file == nil {
		panic("NewFile expects non-nil file")
	}
	return Value{p: unsafe.Pointer(&FileObject{File: file, Path: file.Name()}), tag: TagFile}
}

func (v Value) FileObject() *FileObject {
	if v.tag != TagFile {
		panic("FileObject called on non-file Value")
	}
	if v.p == nil {
		panic("file Value does not contain file pointer")
	}
	return (*FileObject)(v.p)
}

func (v Value) Close() error {
	if v.tag != TagFile {
		panic("close expects a file Value")
	}
	return v.FileObject().Close()
}

func (f *FileObject) Close() error {
	if f == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	if f.File == nil {
		return nil
	}
	f.lineReader = nil
	return f.File.Close()
}
