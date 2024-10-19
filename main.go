package main

import (
    "fmt"
    "log"
    "os"
    "github.com/fsnotify/fsnotify"
    "sync"
)

type Config struct {
    FolderPath       string
    NumberOfWatchers int
    BannedFiles      []string
    ImportantFiles   []string
    OperationsToWatch []fsnotify.Op
}

func main() {

    var wg sync.WaitGroup

    configs := []Config{
        {
            FolderPath:       "/home/janek/niestudia",
            NumberOfWatchers: 1,
            BannedFiles:      []string{"file1", "file2"},
            ImportantFiles:   []string{"important1", "important2"},
            OperationsToWatch: []fsnotify.Op{fsnotify.Create, fsnotify.Write},
        },
        {
            FolderPath:       "/home/janek/studia",
            NumberOfWatchers: 1,
            BannedFiles:      []string{"file3", "file4"},
            ImportantFiles:   []string{"important3", "important4"},
            OperationsToWatch: []fsnotify.Op{fsnotify.Remove, fsnotify.Rename, fsnotify.Create, fsnotify.Write},
        },
    }

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

            var isbannedEmpty bool = len(config.BannedFiles) == 0
            var isImportantEmpty bool = len(config.ImportantFiles) == 0 
            if isbannedEmpty && isImportantEmpty {
                log.Println("elo")
            }

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
                    for _, op := range config.OperationsToWatch {
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
