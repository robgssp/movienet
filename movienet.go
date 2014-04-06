package main

import (
	//    "fmt"
	"flag"
)

func main() {
	flag.Parse()
	argv := flag.Args()
	library := NewMediaLibrary()
	ProcessFilesystem(library, argv)
}
