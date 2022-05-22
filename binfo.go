package binfo

import (
	"context"
	"debug/buildinfo"
	"go/build"
	"os"
	"path/filepath"
	"runtime/debug"
)

// Gobin returns GOBIN value.
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
func List(ctx context.Context) ([]*ExeInfo, error) {
	dir := Gobin()
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
