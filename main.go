package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	// Wczytaj plik YAML
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("cannot read file: %v", err)
	}

	// Parsuj plik YAML
	var configs Config
	err = yaml.Unmarshal(file, &configs)
	if err != nil {
		log.Fatalf("cannot unmarshal yaml: %v", err)
	}

	// Konwertuj operacje na []fsnotify.Op
	err = processConfig(configs.Folders)
	if err != nil {
		log.Fatalf("error processing config: %v", err)
	}

	var wg sync.WaitGroup

	for _, folder := range configs.Folders {

		wg.Add(1)

		go func(folder Folder) {
			defer wg.Done()
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}
			defer watcher.Close()

			// Dodaj folder do watchera
			err = watcher.Add(folder.FolderPath)
			if err != nil {
				log.Fatalf("Error: %v in path %v", err, folder.FolderPath)
			}

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					//log.Println("Event: ", event)
					for _, op := range folder.OperationsToWatchProcessed {
						if event.Op.Has(op) {
							//log.Println("Operation matched: ", op)
							// Dodaj dodatkową logikę dla banned i important files tutaj
							//event.name tp ścieżka jak wziać nazwe pliku
							for _, file := range folder.ImportantFiles {
								if filepath.Base(event.Name) == file {
									log.Printf("Successfully detected %v on %v", event.Op, file)
								}
							}
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}(folder)
	}

	// Poczekaj na zakończenie wszystkich gorutyn
	wg.Wait()
}
