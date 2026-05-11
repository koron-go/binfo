package binfo

import (
	"context"
	"debug/buildinfo"
	"go/build"
	"io/fs"
	"iter"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

// Gobin returns GOBIN value.
// If GOBIN is vaialbe, this returns GOPATH/bin instead of it.
func Gobin() string {
	gobin := os.Getenv("GOBIN")
	if gobin != "" {
		return gobin
	}
	list := filepath.SplitList(build.Default.GOPATH)
	if len(list) > 0 && list[0] != "" {
		return filepath.Join(list[0], "bin")
	}
	return ""
}

// ExeInfo is information of executable file.
type ExeInfo struct {
	Name string
	Err  error
	debug.BuildInfo
}

func readExeInfo(name string) *ExeInfo {
	bi, err := buildinfo.ReadFile(name)
	if err != nil {
		return &ExeInfo{Name: name, Err: err}
	}
	return &ExeInfo{Name: name, BuildInfo: *bi}
}

// List returns list of ExeInfo in GOBIN.
func List(ctx context.Context, dir string) ([]*ExeInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var list []*ExeInfo
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if e.IsDir() {
			continue
		}
		name := filepath.Join(dir, e.Name())
		list = append(list, readExeInfo(name))
	}
	return list, nil
}

// List2 enumerates all ExeInfo in the directory asynchronously.
func List2(ctx context.Context, dir string) <-chan *ExeInfo {
	ch := make(chan *ExeInfo)
	cctx, cancel := context.WithCancelCause(ctx)
	go func() {
		defer close(ch)
		entries, err := os.ReadDir(dir)
		if err != nil {
			cancel(err)
			return
		}
		for _, e := range entries {
			if err := cctx.Err(); err != nil {
				cancel(err)
				return
			}
			if e.IsDir() {
				continue
			}
			name := filepath.Join(dir, e.Name())
			ch <- readExeInfo(name)
		}
		cancel(nil)
	}()
	return ch
}

func ReadDir(dir string) (iter.Seq[ExeInfo], error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("failed to read directory entries at %q", dir)
		return nil, err
	}
	return func(yield func(ExeInfo) bool) {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := filepath.Join(dir, e.Name())
			info := readExeInfo(name)
			if !yield(*info) {
				return
			}
		}
	}, nil
}

type DirReader struct {
	path  string
	dir   *os.File
	cache []fs.DirEntry
}

func NewDirReader(path string) (*DirReader, error) {
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &DirReader{
		path: path,
		dir:  dir,
	}, nil
}

func (r *DirReader) readDir() (fs.DirEntry, error) {
	if len(r.cache) == 0 {
		entries, err := r.dir.ReadDir(64)
		if err != nil {
			return nil, err
		}
		r.cache = entries
	}
	entry := r.cache[0]
	r.cache = r.cache[1:]
	return entry, nil
}

func (r *DirReader) Read() (*debug.BuildInfo, error) {
	for {
		entry, err := r.readDir()
		if err != nil {
			return nil, err
		}
		if !entry.Type().IsRegular() {
			continue
		}
		return buildinfo.ReadFile(filepath.Join(r.path, entry.Name()))
	}
}

func (r *DirReader) Path() string {
	return r.path
}

func (r *DirReader) Close() error {
	return r.dir.Close()
}
