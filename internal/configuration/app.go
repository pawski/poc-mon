package configuration

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var appConfigFile = "app.yml"
var appConfiguration AppConfig

type AppConfig struct {
	TestUrl string `yaml:"test_url"`
}

func GetApp() (AppConfig, error) {
	file, err := ioutil.ReadFile(appConfigFile)
	if err != nil {
		return appConfiguration, errors.Wrap(err, "Cannot get application configuration")
	}

	err = yaml.Unmarshal(file, &appConfiguration)
	if err != nil {
		return appConfiguration, errors.Wrap(err, "Cannot decode application configuration")
	}

	return appConfiguration, nil
}
