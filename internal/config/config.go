package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Config struct {
	//struct that represents the JSON file structure
	DbUrl    string `json:"db_url"`
	UserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	//func that reads the config file in the home directory
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("could not get config file path: %w", err)
	}
	//open the config file
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("could not open config file: %w", err)
	}
	defer configFile.Close()
	//read the content of the config file
	content, err := io.ReadAll(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config file: %w", err)
	}
	//unmarshal the content of the config file
	var cfg Config
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("could not unmarshal config file: %w", err)
	}
	//return the config
	return cfg, nil
}

func (c Config) SetUser(name string) error {
	//func that sets the current user name
	c.UserName = name
	err := write(c)
	if err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}
	return nil
}

func getConfigFilePath() (string, error) {
	//func that returns the path to the config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	return homeDir + "/" + configFileName, nil
}

func write(cfg Config) error {
	//func that writes the config to a file
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("could not get config file path: %w", err)
	}

	configFile, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("could not create config file: %w", err)
	}
	defer configFile.Close()

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("could not marshal config: %w", err)
	}

	if _, err := configFile.Write(jsonData); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}
	return nil
}

const configFileName = ".gatorconfig.json"
