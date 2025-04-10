package config

import (
	"os"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl           *string `json:"db_url"`
	CurrentUserName *string `json:"current_user_name"`
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = &name
	err := write(*c)
	if err != nil {
		return err
	}
	return nil
}


func Read() (*Config, error) {
	configFile, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(configFile, os.O_RDONLY|os.O_CREATE, 0311)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dburl := "postgres://example"
	cfg := Config{
		DBUrl: &dburl,
		CurrentUserName: nil,
	}
	data, err := ioutil.ReadAll(file) 
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}



func getConfigFilePath() (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, configFileName), nil
}

func write(cfg Config) error {
	configFile, err := getConfigFilePath()
	if err != nil {
		return err
	}
	
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFile, data, 0311)
	if err != nil {
		return err
	}
	return nil
}
