package main

import (
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"path/filepath"
	"strings"
	"net"
	"encoding/json"
)

var WatchChanges uint32 = fsnotify.FSN_DELETE | fsnotify.FSN_RENAME | fsnotify.FSN_CREATE

func ScanFiles(media *MediaLibrary, root string, watcher *fsnotify.Watcher) *MediaLibrary {

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		isdir := (info.Mode() & os.ModeDir) != 0
		if isdir {
			watcher.WatchFlags(path, WatchChanges)
		}
		media.Add(root, strings.Replace(path, root, "", 1), isdir)
		return nil
	})
	return media
}

func DumpJson(media *MediaLibrary, c net.Conn) {
	enc := json.NewEncoder(c)
	lib := struct {
		Name string `json:"name"`
		Tree []MediaFolderJson `json:"tree"`
	}{Name:"bootleg", Tree:make([]MediaFolderJson,0,len(media.dirs))}
	for _, root := range media.dirs {
		lib.Tree = append(lib.Tree, root.ToJson())
	}
	if enc.Encode(lib) != nil {
		panic("Could not encode JSON properly")
	}
}

func ProcessFilesystem(library *MediaLibrary, dirs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev)
				if ev.IsCreate()  {
					library.AddFromEvent(ev.Name)
				}
				if ev.IsDelete() || ev.IsRename() {
					library.RemoveFromEvent(ev.Name)
				}
				log.Println("All items: ", library)
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	for _, fold := range dirs {
		err = watcher.WatchFlags(fold, WatchChanges)
		if err != nil {
			log.Fatal(err)
		}
		library = ScanFiles(library, fold, watcher)
	}

	log.Println("All items:", library)

	// Block here, so that the function doesn't end before the goroutine
	// (which should be "never")
	<-done
}
