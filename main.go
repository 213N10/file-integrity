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
	configFilePath := getConfigFile()
	configFile, err := os.ReadFile(configFilePath)
	if err != nil {
		text_message := "check whether you set up env variable 'FILE-INTEGRITY_CONFIG_FILEPATH' or have '/etc/file-integrity/config.yaml'"
		log.Fatalf("cannot read file config file: %v\n%v", err, text_message)
	}

	var configs Config
	err = yaml.Unmarshal(configFile, &configs)
	if err != nil {
		log.Fatalf("cannot process data from config file: %v\nWhen in doubt check config file format on: 'https://github.com/213N10/file-integrity'", err)
	}

	logFile, err := os.OpenFile(configs.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Couldnt open the log file: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "ERROR ", log.LstdFlags)

	eventsChan := make(chan string)

	go logEvents(eventsChan, logger)

	err = processConfig(configs.Folders)
	if err != nil {
		logger.Fatalf("error processing folder data: %v", err)
	}

	var wg sync.WaitGroup

	for _, folder := range configs.Folders {

		wg.Add(1)

		go func(folder Folder) {
			defer wg.Done()
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				eventsChan <- fmt.Sprintf("Error: %v. Closing thread.", err)
				return
			}
			defer watcher.Close()

			err = watcher.Add(folder.FolderPath)
			if err != nil {
				eventsChan <- fmt.Sprintf("Error: %v in path %v. Closing thread.", err, folder.FolderPath)
				return
			}

			var scanMode int
			if len(folder.ImportantFiles) == 0 {
				scanMode = 2
			} else {
				scanMode = 1
			}

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					switch scanMode {
					case 1:
						for _, op := range folder.OperationsToWatchProcessed {
							if event.Op.Has(op) {
								for _, file := range folder.ImportantFiles {
									if filepath.Base(event.Name) == file {
										eventsChan <- fmt.Sprintf("Successfully detected %v on %v", event.Op, file)
										break
									}
								}
							}
						}
					case 2:
						for _, op := range folder.OperationsToWatchProcessed {
							if event.Op.Has(op) {
								eventsChan <- fmt.Sprintf("Successfully detected %v on %v", event.Op, event.Name)
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

	wg.Wait()
	close(eventsChan)
}

func logEvents(eventsChan chan string, logger *log.Logger) {
	for event := range eventsChan {
		logger.Println(event)
	}
}

func getConfigFile() string {
	config_path := os.Getenv("FILE_INTEGRITY_CONFIG_FILEPATH")
	if config_path == "" {
		config_path = "etc/file-integrity/config.yaml"
	}
	return config_path
}
