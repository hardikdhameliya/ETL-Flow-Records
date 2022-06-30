package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var conf *viper.Viper

func init() {
	conf = viper.New()
	conf.SetConfigName("config") // name of config file (without extension)
	conf.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	conf.AddConfigPath(".")      // optionally look for config in the working directory
	conf.AddConfigPath("$HOME/") // optionally look for config in the home directory

	err := conf.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

// Get entire section of configuration
func GetSection(section ...string) (map[string]interface{}, error) {

	if len(section) == 0 {
		return nil, errors.New("Section is not specified")
	}
	s := strings.Join(section, ".")
	if conf.IsSet(s) {
		return conf.GetStringMap(s), nil
	}
	return nil, errors.New("Section is not exist")
}

// Get specific string parameter from a section
func GetSectionParamString(parameters ...string) (string, error) {

	if len(parameters) == 0 {
		return "", errors.New("Section is not specified")
	}
	s := strings.Join(parameters, ".")
	if conf.IsSet(s) {
		return conf.GetString(s), nil
	}
	return "", errors.New("Section is not exist")
}

// Get specific duration parameter from a section
func GetSectionParamDuration(parameters ...string) (time.Duration, error) {
	if len(parameters) == 0 {
		return 0, errors.New("Section is not specified")
	}
	s := strings.Join(parameters, ".")
	if conf.IsSet(s) {
		return conf.GetDuration(s), nil
	}
	return 0, errors.New("Section is not exist")
}

// Get specific integer parameter from a section
func GetSectionParamInt(parameters ...string) (int, error) {

	if len(parameters) == 0 {
		return 0, errors.New("Section is not specified")
	}
	s := strings.Join(parameters, ".")
	if conf.IsSet(s) {
		return conf.GetInt(s), nil
	}
	return 0, errors.New("Section is not exist")
}
