package main

import (
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"path/filepath"
)

type MediaData struct {
	name      string
	directory bool
	id        uint
}

var WatchChanges uint32 = fsnotify.FSN_DELETE | fsnotify.FSN_RENAME | fsnotify.FSN_CREATE

func ListFiles(filename string, watcher *fsnotify.Watcher) []MediaData {

	media := make([]MediaData, 0)
	var i uint
	filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		fname, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		isdir := (info.Mode() & os.ModeDir) != 0
		if isdir {
			watcher.WatchFlags(path, WatchChanges)
			log.Println("Now tracking folder", path)
		}
		media = append(media, MediaData{
			fname, isdir, i,
		})
		i++
		return nil
	})
	return media
}

func ProcessFilesystem(dirs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev)
				if ev.IsAttrib() {
					log.Println("It's an attribute change!")
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	for _, fold := range dirs {
		err = watcher.WatchFlags(fold, WatchChanges)
		log.Println("Found items: ", ListFiles(fold, watcher))
		if err != nil {
			log.Fatal(err)
		}
	}

	<-done

	/* ... do stuff ... */
	watcher.Close()
}
