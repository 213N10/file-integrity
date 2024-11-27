package main

import (
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

type Folder struct {
	FolderPath                 string   `yaml:"folder_path"`
	ImportantFiles             []string `yaml:"important_files"`
	OperationsToWatch          []string `yaml:"operations_to_watch"`
	OperationsToWatchProcessed []fsnotify.Op
}

type Config struct {
	LogPath   string   `yaml:"log_path"`
	LogLevel  string   `yaml:"log_level"`
	LogFormat string   `yaml:"log_format"`
	Folders   []Folder `yaml:"folders"`
}

var opMapping = map[string]fsnotify.Op{
	"create": fsnotify.Create,
	"write":  fsnotify.Write,
	"remove": fsnotify.Remove,
	"rename": fsnotify.Rename,
}

var logLevelMapping = map[string]slog.Level{
	"debug":   slog.LevelDebug,
	"info":    slog.LevelInfo,
	"warning": slog.LevelWarn,
	"error":   slog.LevelError,
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

func convertLogLevel(level string) (slog.Level, error) {
	mappedLevel, exists := logLevelMapping[level]
	if !exists {
		return slog.Level(0), fmt.Errorf("unknown log level: %s", level)
	}
	return mappedLevel, nil
}

// Funkcja przetwarzająca konfigurację i konwertująca operacje na []fsnotify.Op
func processConfig(configs []Folder) error {
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
