package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"sync"
)

type Config struct {
	FolderPath                 string   `yaml:"folder_path"`
	NumberOfWatchers           int      `yaml:"number_of_watchers"`
	ImportantFiles             []string `yaml:"important_files"`
	OperationsToWatch          []string `yaml:"operations_to_watch"`
	OperationsToWatchProcessed []fsnotify.Op
}

var opMapping = map[string]fsnotify.Op{
	"create": fsnotify.Create,
	"write":  fsnotify.Write,
	"remove": fsnotify.Remove,
	"rename": fsnotify.Rename,
}

// Funkcja do konwersji operacji ze stringów do fsnotify.Op
func convertOps(ops []string) ([]fsnotify.Op, error) {
	var result []fsnotify.Op
	for _, op := range ops {
		mappedOp, exists := opMapping[op]
		if !exists {
			return nil, fmt.Errorf("unknown operation: %s", op)
		}
		result = append(result, mappedOp)
	}
	return result, nil
}

// Funkcja przetwarzająca konfigurację i konwertująca operacje na []fsnotify.Op
func processConfig(configs []Config) error {
	for i, config := range configs {
		ops, err := convertOps(config.OperationsToWatch)
		if err != nil {
			return err
		}
		// Przypisz wynikowe []fsnotify.Op do pola Ops
		configs[i].OperationsToWatchProcessed = ops
	}
	return nil
}

func main() {
	// Wczytaj plik YAML
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("cannot read file: %v", err)
	}

	// Parsuj plik YAML
	var configs []Config
	err = yaml.Unmarshal(file, &configs)
	if err != nil {
		log.Fatalf("cannot unmarshal yaml: %v", err)
	}

	// Konwertuj operacje na []fsnotify.Op
	err = processConfig(configs)
	if err != nil {
		log.Fatalf("error processing config: %v", err)
	}

	var wg sync.WaitGroup

	for _, config := range configs {

		wg.Add(1)

		go func(config Config) {
			defer wg.Done()
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				fmt.Println("Error: ", err)
				os.Exit(1)
			}
			defer watcher.Close()

			// Dodaj folder do watchera
			err = watcher.Add(config.FolderPath)
			if err != nil {
				log.Fatal(err)
			}

			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					log.Println("Event: ", event)
					for _, op := range config.OperationsToWatchProcessed {
						if event.Op.Has(op) {
							log.Println("Operation matched: ", op)
							// Dodaj dodatkową logikę dla banned i important files tutaj
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}(config)
	}

	// Poczekaj na zakończenie wszystkich gorutyn
	wg.Wait()
}
