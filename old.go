package main

import (
	"fmt"
	"log"
	"os"
	//"gopkg.in/yaml.v3"
	"github.com/fsnotify/fsnotify"
)


type Config1 struct {
    FolderPath       string
    NumberOfWatchers int
    BannedFiles      []string
    ImportantFiles   []string
    OperationsToWatch []fsnotify.Op
}


func mainly() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)	
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("Event: ", event)
				if event.Has(fsnotify.Write) {
                    log.Println("modified file:", event.Name)
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("error:", err)
			}
		}
	} ()


	// Add a path.
    err = watcher.Add("/home/janek")
    if err != nil {
        log.Fatal(err)
    }

    // Block main goroutine forever.
    <-make(chan struct{})
}