package util

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type SimpleConfig struct {
	data map[string]interface{}
}

var Config SimpleConfig

func createDefaultConfig() map[string]interface{} {
	var data = make(map[string]interface{})
	dir, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal("Could not generate default Config.", err)
	}
	if runtime.GOOS == "windows" {
		data["save_path"] = filepath.Join(dir, "Documents", "ScAr")
	} else {
		data["save_path"] = dir + "/ScAr"
	}
	data["moodle_url"] = ""
	data["moodle_username"] = ""
	data["moodle_password"] = ""

	data["digi4s_username"] = ""
	data["digi4s_password"] = ""
	return data
}

func getConfigFilePath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var configDir string
	if runtime.GOOS == "windows" {
		configDir = filepath.Join(dir, "AppData", "Local", "scar")
	} else {
		configDir = filepath.Join(dir, ".config", "scar")
	}
	var configFile = filepath.Join(configDir, "config.json")
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return "", err
	}
	return configFile, nil
}

func (sc *SimpleConfig) Load() {
	configFile, err := getConfigFilePath()
	if err != nil {
		logrus.Fatal("Could not load config. ", err.Error())
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		logrus.Error("Could not load config data. Maybe file does not exists. Loading a default config")
		sc.data = createDefaultConfig()
		sc.save()
		return
	}

	var result map[string]interface{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Fatalf("Could not load config. Config is not a valid json: %v", err)
		return
	}
	sc.data = result
}
func (sc *SimpleConfig) GetString(key string) string {
	if value, ok := sc.data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	logrus.Error("Could not load config key. Maybe wrong type or does not exists: ", key)
	return ""
}
func (sc *SimpleConfig) GetStringWD(key string, defaultValue string) string {
	if value, ok := sc.data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}

	sc.SaveValue(key, defaultValue)
	return defaultValue
}
func (sc *SimpleConfig) GetInt(key string, defaultValue int) int {
	if value, ok := sc.data[key]; ok {
		if floatValue, ok := value.(float64); ok {
			return int(floatValue)
		}
	}
	sc.SaveValue(key, defaultValue)
	return defaultValue
}

func (sc *SimpleConfig) GetFloat(key string, defaultValue float64) float64 {
	if value, ok := sc.data[key]; ok {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
	}
	sc.SaveValue(key, defaultValue)
	return defaultValue
}
func (sc *SimpleConfig) SaveValue(key string, value interface{}) {
	sc.data[key] = value
	sc.save()
}

func (sc *SimpleConfig) save() {
	configFile, err := getConfigFilePath()
	if err != nil {
		logrus.Error("Could not save Config file", err.Error())
		return
	}
	jsonData, err := json.Marshal(sc.data)
	if err != nil {
		logrus.Error("Could not save Config file. Failed to create Json string", err.Error())
		return
	}
	err = os.WriteFile(configFile, jsonData, os.ModePerm)
	if err != nil {
		logrus.Error("Could not write Config file.", err.Error())
		return
	}
}
