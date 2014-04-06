package main

import (
	"path/filepath"
//	"fmt"
)

type MediaID uint

type MediaData struct {
	name string
	path string
	id   MediaID
}

type MediaFolder struct {
	name    string
	subdirs map[string]*MediaFolder
	files   []*MediaData
}

type MediaLibrary struct {
	data  map[MediaID]*MediaData
	dirs  map[string]*MediaFolder
	maxid MediaID
}

func NewMediaLibrary() *MediaLibrary {
	return &MediaLibrary{
		data: make(map[MediaID]*MediaData),
		dirs: make(map[string]*MediaFolder),
	}
}

func newMediaFolder(name string) *MediaFolder {
	return &MediaFolder{
		name:    name,
		subdirs: make(map[string]*MediaFolder),
		files:   make([]*MediaData, 0, 1),
	}
}

func splitPath(path string) []string {
	ret := make([]string, 0, 1)
	path = filepath.Clean(path)
	for path != "." {
		ret = append(ret, filepath.Base(path))
		path = filepath.Dir(path)
	}
	for i, j := 0, len(ret)-1; i < j; i, j = i+1, j-1 {
		ret[j], ret[i] = ret[i], ret[j]
	}
	return ret
}

func (lib *MediaLibrary) Add(root string, fpath string, isdir bool) MediaID {

	// Recurse into the subdirectories
	path := splitPath(fpath)
	if _, hasroot := lib.dirs[root]; !hasroot {
		lib.dirs[root] = newMediaFolder(root)
	}
	fold := lib.dirs[root]

	id := lib.maxid
	name := path[len(path)-1]
	path = path[:len(path)-1]

	for _, dir := range path {
		if _, exists := fold.subdirs[dir]; !exists {
			fold.subdirs[dir] = newMediaFolder(dir)
		}
		fold = fold.subdirs[dir]
	}

	if isdir {
		if _, exists := fold.subdirs[name]; !exists {
			fold.subdirs[name] = newMediaFolder(name)
		}
	} else {
		data := &MediaData{name: name, path: fpath, id: id}
		fold.files = append(fold.files, data)
		lib.data[id] = data
		lib.maxid++
	}

	return id
}

func (lib *MediaLibrary) Contains(id MediaID) bool {
	_, has := lib.data[id]
	return has
}

func (lib *MediaLibrary) Get(id MediaID) *MediaData {
	return lib.data[id]
}
