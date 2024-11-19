package src

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveToFile saves the given data to a JSON file at the specified path
func SaveToFile(data interface{}, dir string, filename string) error {
	// Create the directory if it doesn't exist
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the full path to the file
	filePath := filepath.Join(dir, filename)

	// Open the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Marshal the data into JSON and write to the file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print the JSON
	err = encoder.Encode(data)
	if err != nil {
		return fmt.Errorf("failed to write JSON data: %v", err)
	}

	return nil
}
