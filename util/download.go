package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func SaveStructToJSON(data interface{}, filePath string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling to JSON: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}
