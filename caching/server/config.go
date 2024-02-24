package server

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Cache struct {
		Master struct {
			Address string `yaml:"address"`
			Port    string `yaml:"port"`
		} `yaml:"master"`
		AofFileUrl string `yaml:"aof"`
	} `yaml:"cache"`
}

func readMasterNodeConfigs(configFileName string) (string, string, error) {
	yamlFile, err := os.ReadFile(configFileName)
	if err != nil {
		return "", "", err
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return "", "", err
	}

	return config.Cache.Master.Address, config.Cache.Master.Port, nil
}

func getAofFileLocation(configFileName string) (string, error) {
	yamlFile, err := os.ReadFile(configFileName)
	if err != nil {
		return "", err
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return "", err
	}

	return config.Cache.AofFileUrl, nil
}
