package binfo

import (
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

// List returns list of ExeInfo in GOBIN.
func List(n int) ([]*ExeInfo, error) {
	dir := Gobin()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var list []*ExeInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := filepath.Join(dir, e.Name())
		bi, err := buildinfo.ReadFile(name)
		if err != nil {
			list = append(list, &ExeInfo{Name: name, Err: err})
			continue
		}
		list = append(list, &ExeInfo{Name: name, BuildInfo: *bi})
	}
	return list, nil
}
