package main

import (
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"path/filepath"
)

var WatchChanges uint32 = fsnotify.FSN_DELETE | fsnotify.FSN_RENAME | fsnotify.FSN_CREATE

func ScanFiles(media *MediaLibrary, root string, watcher *fsnotify.Watcher) *MediaLibrary {

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		isdir := (info.Mode() & os.ModeDir) != 0
		if isdir {
			watcher.WatchFlags(path, WatchChanges)
		}
		media.Add(root, path, isdir)
		return nil
	})
	return media
}

func ProcessFilesystem(dirs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	library := NewMediaLibrary()

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev.Name)
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

	log.Println("All items: ", library)

	// Block here, so that the function doesn't end before the goroutine
	// (which should be "never")
	<-done
}
