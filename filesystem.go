package main

import (
	"os"
	"log"
	"github.com/howeyc/fsnotify"
)

func ls(filename string) []string {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if f == nil {
		log.Fatalf("ls: cannot access %s: %s\n", filename, err)
	}
	defer f.Close()

	files, err := f.Readdirnames(-1)
	if files == nil {
		log.Fatalf("ls: could not get files in %s: %s\n", filename, err)
	}

	return files
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

	var flags uint32 = fsnotify.FSN_DELETE | fsnotify.FSN_RENAME | fsnotify.FSN_CREATE
	for _, fold := range dirs {
		err = watcher.WatchFlags(fold, flags)
		log.Println("Found items: ", ls(fold))
		if err != nil {
			log.Fatal(err)
		}
	}

	<-done

	/* ... do stuff ... */
	watcher.Close()
}
