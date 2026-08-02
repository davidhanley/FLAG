package runtime

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func OpenFile(path string) Value {
	if path == "" {
		panic("open-file expects a non-empty path")
	}
	file, err := os.Open(path)
	if err != nil {
		panic("open-file failed: " + err.Error())
	}
	return NewFile(file)
}

func OpenWriter(path string) Value {
	if path == "" {
		panic("io/writer expects a non-empty path")
	}
	file, err := os.Create(path)
	if err != nil {
		panic("io/writer failed: " + err.Error())
	}
	return NewFile(file)
}

func Write(file Value, content string) Value {
	switch file.tag {
	case TagFile:
		f := file.FileObject()
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.closed || f.File == nil {
			panic("write expects an open file")
		}
		if _, err := f.File.WriteString(content); err != nil {
			panic("write failed: " + err.Error())
		}
		return NilValue()
	default:
		panic("write expects a file")
	}
}

func LineSeq(file Value) Value {
	return FileToStrings(file)
}

func ReadLine(file Value) Value {
	switch file.tag {
	case TagFile:
		return readLineFile(file.FileObject())
	default:
		panic("io/readline expects a file")
	}
}

func ScanDirectory(path string) Value {
	if strings.TrimSpace(path) == "" {
		panic("io/scan-directory expects a non-empty path")
	}

	type fileEntry struct {
		filename string
		name     string
		size     int64
		updated  time.Time
	}

	entries := make([]fileEntry, 0, 128)
	err := filepath.WalkDir(path, func(filePath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		entries = append(entries, fileEntry{
			filename: filePath,
			name:     d.Name(),
			size:     info.Size(),
			updated:  info.ModTime(),
		})
		return nil
	})
	if err != nil {
		panic("io/scan-directory failed: " + err.Error())
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].filename < entries[j].filename
	})

	out := make([]Value, 0, len(entries))
	for _, entry := range entries {
		updated := entry.updated.UTC().Format(time.RFC3339)
		out = append(out, NewMap(
			NewKeyword("filename"), NewString(entry.filename),
			NewKeyword("name"), NewString(entry.name),
			NewKeyword("size"), NewLong(entry.size),
			NewKeyword("created-date-time"), NewString(updated),
			NewKeyword("updated-date-time"), NewString(updated),
		))
	}
	return NewArray(out...)
}

func readLineFile(file *FileObject) Value {
	if file == nil {
		panic("io/readline expects a file")
	}

	file.mu.Lock()
	defer file.mu.Unlock()
	line, err := file.bufferedReaderLocked().ReadString('\n')
	if err != nil && err != io.EOF {
		panic("io/readline failed: " + err.Error())
	}
	if err == io.EOF && line == "" {
		return NilValue()
	}
	return NewString(strings.TrimRight(line, "\r\n"))
}

func FileToStrings(file Value) Value {
	switch file.tag {
	case TagFile:
		return fileToStringsFile(file.FileObject())
	case TagString:
		return FileToStringsPath(file.StringValue())
	case TagSymbol:
		return FileToStringsPath(Name(file))
	default:
		panic("file-to-strings expects a file, string, symbol, or keyword")
	}
}

func FileToStringsPath(path string) Value {
	if path == "" {
		panic("file-to-strings expects a non-empty path")
	}

	var (
		f       *os.File
		scanner *bufio.Scanner
		opened  bool
		closed  bool
	)

	closeFile := func() {
		if closed || f == nil {
			return
		}
		closed = true
		if err := f.Close(); err != nil {
			panic("file-to-strings close failed: " + err.Error())
		}
	}

	return NewLazyList(func() (Value, bool) {
		if !opened {
			opened = true
			var err error
			f, err = os.Open(path)
			if err != nil {
				panic("file-to-strings open failed: " + err.Error())
			}
			scanner = bufio.NewScanner(f)
		}

		if scanner.Scan() {
			return NewSymbol(scanner.Text()), true
		}

		closeFile()
		if err := scanner.Err(); err != nil {
			panic("file-to-strings read failed: " + err.Error())
		}
		return Value{}, false
	})
}

func fileToStringsFile(file *FileObject) Value {
	if file == nil {
		panic("file-to-strings expects a file")
	}

	var (
		scanner *bufio.Scanner
		opened  bool
	)

	closeFile := func() {
		if err := file.Close(); err != nil {
			panic("file-to-strings close failed: " + err.Error())
		}
	}

	return NewLazyList(func() (Value, bool) {
		if !opened {
			opened = true
			scanner = bufio.NewScanner(file.File)
		}

		if scanner.Scan() {
			return NewSymbol(scanner.Text()), true
		}

		closeFile()
		if err := scanner.Err(); err != nil {
			panic("file-to-strings read failed: " + err.Error())
		}
		return Value{}, false
	})
}
