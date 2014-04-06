package main

import (
	"os"
	"path/filepath"
	"strings"
)

type MediaID uint

type MediaData struct {
	name string `json:"name"`
	path string `json:"-"`
	id   MediaID `json:"`
}

type MediaDataJson struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Id   uint   `json:"id"`
}

func (data *MediaData) ToJson() MediaDataJson {
	return MediaDataJson{Type:"file", Name:data.name, Id:uint(data.id)}
}

type MediaFolder struct {
	name    string `json:"name"`
	subdirs map[string]*MediaFolder
	files   []*MediaData
}

type MediaFolderJson struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Children   []interface{}   `json:"children"`
}

func (fold *MediaFolder) ToJson() MediaFolderJson {
	js := MediaFolderJson{
		Type: "dir",
		Name: fold.name,
	}
	js.Children = make([]interface{}, 0, len(fold.files)+len(fold.subdirs))
	for _, fi := range fold.files {
		js.Children = append(js.Children, fi.ToJson())
	}
	for _, dir := range fold.subdirs {
		js.Children = append(js.Children, dir.ToJson())
	}
	return js
}

type MediaLibrary struct {
	data  map[MediaID]*MediaData `json:"-"`
	dirs  map[string]*MediaFolder `json:"tree"`
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

func (lib *MediaLibrary) Add(root, fpath string, isdir bool) (MediaID, bool) {

	if fpath == "" {
		return 0, false
	}

	if strings.HasPrefix(fpath, "/") {
		fpath = fpath[1:]
	}

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
		return 0, false
	} else {
		data := &MediaData{name: name, path: fpath, id: id}
		fold.files = append(fold.files, data)
		lib.data[id] = data
		lib.maxid++
		return id, true
	}

}

func (lib *MediaLibrary) AddFromEvent(fpath string) {
	for root := range lib.dirs {
		if strings.HasPrefix(fpath, root) {
			fi, err := os.OpenFile(fpath, os.O_RDONLY, 0)
			if err != nil {
				panic(err) // Shouldn't be happening
			}
			info, err := fi.Stat()
			if err != nil {
				panic(err)
			}
			isdir := (info.Mode() & os.ModeDir) != 0
			lib.Add(root, strings.Replace(fpath, root, "", 1), isdir)
			break
		}
	}
}

func (lib *MediaLibrary) rmSubdir(dir *MediaFolder) {
	for key := range dir.subdirs {
		lib.rmSubdir(dir.subdirs[key])
	}
	for _, file := range dir.files {
		lib.data[file.id] = nil
		delete(lib.data, file.id)
	}
}

func (lib *MediaLibrary) Remove(root, fpath string) {

	// Recurse into the subdirectories
	path := splitPath(fpath)
	if _, hasroot := lib.dirs[root]; !hasroot {
		return
	}
	fold := lib.dirs[root]

	name := path[len(path)-1]
	path = path[:len(path)-1]

	for _, dir := range path {
		if _, exists := fold.subdirs[dir]; !exists {
			return // It's already not part of the directory
		}
		fold = fold.subdirs[dir]
	}

	if _, exists := fold.subdirs[name]; exists {
		// It's a directory
		if _, exists := fold.subdirs[name]; exists {
			lib.rmSubdir(fold.subdirs[name])
			delete(fold.subdirs, name)
		}
	} else {
		// It's a file
		i := -1
		for p, v := range fold.files {
			if v.name == name {
				i = p
			}
		}
		if i == -1 {
			panic("Couldn't find file")
		}
		id := fold.files[i].id
		fold.files[i] = fold.files[len(fold.files)-1]
		fold.files = fold.files[:len(fold.files)-1]
		delete(lib.data, id)
	}

}

func (lib *MediaLibrary) RemoveFromEvent(fpath string) {
	for root := range lib.dirs {
		if strings.HasPrefix(fpath, root) {
			lib.Remove(root, strings.Replace(fpath, root, "", 1)[1:])
			break
		}
	}
}

func (lib *MediaLibrary) Contains(id MediaID) bool {
	_, has := lib.data[id]
	return has
}

func (lib *MediaLibrary) Get(id MediaID) *MediaData {
	return lib.data[id]
}
