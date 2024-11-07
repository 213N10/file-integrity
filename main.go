package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

func main() {
	// Wczytaj plik YAML
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("cannot read file: %v", err)
	}

	// Parsuj plik YAML
	var configs Config
	err = yaml.Unmarshal(configFile, &configs)
	if err != nil {
		log.Fatalf("cannot unmarshal yaml: %v", err)
	}

	logFile, err := os.OpenFile(configs.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Couldnt open the log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	eventsChan := make(chan string)

	go logEvents(eventsChan, logger)

	// Konwertuj operacje na []fsnotify.Op
	err = processConfig(configs.Folders)
	if err != nil {
		logger.Fatalf("error processing config: %v", err)
	}

	var wg sync.WaitGroup

	for _, folder := range configs.Folders {

		wg.Add(1)

		go func(folder Folder) {
			defer wg.Done()
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				eventsChan <- fmt.Sprintf("Error: %v", err)
				os.Exit(1)
			}
			defer watcher.Close()

			// Dodaj folder do watchera
			err = watcher.Add(folder.FolderPath)
			if err != nil {
				eventsChan <- fmt.Sprintf("Error: %v in path %v", err, folder.FolderPath)
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
							for _, file := range folder.ImportantFiles {
								if filepath.Base(event.Name) == file {
									eventsChan <- fmt.Sprintf("Successfully detected %v on %v", event.Op, file)
								}
							}
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					eventsChan <- fmt.Sprintf("error: %v", err)
				}
			}
		}(folder)
	}

	// Poczekaj na zakończenie wszystkich gorutyn
	wg.Wait()
	close(eventsChan)
}

func logEvents(eventsChan chan string, logger *log.Logger) {
	for event := range eventsChan {
		logger.Println(event)
	}
}
