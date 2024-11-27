package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type LogMessage struct {
	Level   slog.Level
	Message string
}

func main() {
	configFilePath := getConfigFile()
	configFile, err := os.ReadFile(configFilePath)
	if err != nil {
		text_message := "check whether you set up env variable 'FILE_INTEGRITY_CONFIG_FILEPATH' or have '/etc/file-integrity/config.yaml'"
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

	var logLevel slog.Level
	logLevel, err = convertLogLevel(configs.LogLevel)
	if err != nil {
		log.Println("Did not find valid log level. Defaulting to INFO...")
		logLevel = slog.LevelInfo
	}
	logger := setupLogger(configs.LogPath, logLevel, configs.LogFormat)
	eventsChan := make(chan LogMessage)

	go logEvents(eventsChan, logger)

	err = processConfig(configs.Folders)
	if err != nil {
		eventsChan <- LogMessage{
			Level:   slog.LevelError,
			Message: fmt.Sprintf("Error processing folder data: %v. Closing program.", err),
		}
		os.Exit(1)
	}

	var wg sync.WaitGroup

	for _, folder := range configs.Folders {

		wg.Add(1)

		go func(folder Folder) {
			defer wg.Done()
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				eventsChan <- LogMessage{
					Level:   slog.LevelError,
					Message: fmt.Sprintf("Watcher error occurred for folder %v: %v. Closing thread.", folder.FolderPath, err),
				}
				return
			}
			defer watcher.Close()

			err = watcher.Add(folder.FolderPath)
			if err != nil {

				eventsChan <- LogMessage{
					Level:   slog.LevelError,
					Message: fmt.Sprintf("Error: %v in path %v. Closing thread.", err, folder.FolderPath),
				}

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
										eventsChan <- LogMessage{
											Level:   slog.LevelError,
											Message: fmt.Sprintf("Successfully detected %v on %v", event.Op, file),
										}
										break
									}
								}
							}
						}
					case 2:
						for _, op := range folder.OperationsToWatchProcessed {
							if event.Op.Has(op) {
								eventsChan <- LogMessage{
									Level:   slog.LevelError,
									Message: fmt.Sprintf("Successfully detected %v on %v", event.Op, event.Name),
								}
							}
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					eventsChan <- LogMessage{
						Level:   slog.LevelError,
						Message: fmt.Sprintf("Error: %v", err),
					}
				}
			}
		}(folder)
	}

	wg.Wait()
	close(eventsChan)
}

func logEvents(eventsChan chan LogMessage, logger *slog.Logger) {
	for logMsg := range eventsChan {
		switch logMsg.Level {
		case slog.LevelDebug:
			logger.Debug(logMsg.Message)
		case slog.LevelInfo:
			logger.Info(logMsg.Message)
		case slog.LevelWarn:
			logger.Warn(logMsg.Message)
		case slog.LevelError:
			logger.Error(logMsg.Message)
		default:
			logger.Info(logMsg.Message)
		}
	}
}

func getConfigFile() string {
	config_path := os.Getenv("FILE_INTEGRITY_CONFIG_FILEPATH")
	if config_path == "" {
		config_path = "etc/file-integrity/config.yaml"
	}
	return config_path
}

func setupLogger(logFilePath string, logLevel slog.Level, format string) *slog.Logger {
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Couldn't open the log file: %v\n", err)
		os.Exit(1)
	}

	var handler slog.Handler

	switch format {
	case "json":
		handler = slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: logLevel})
	case "text":
		handler = slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: logLevel})
	default:
		fmt.Printf("Unsupported log format: %v. Defaulting to JSON.\n", format)
		handler = slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: logLevel})
	}

	return slog.New(handler)
}
