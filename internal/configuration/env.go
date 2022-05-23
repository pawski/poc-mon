package configuration

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var envConfigFile = "env.yml"
var envConfiguration EnvConfig

type EnvConfig struct {
	HttpServerAddress string `yaml:"http_server_address"`
}

func GetEnv() (EnvConfig, error) {
	file, err := ioutil.ReadFile(envConfigFile)
	if err != nil {
		return envConfiguration, errors.Wrap(err, "Cannot read environment configuration")
	}

	err = yaml.Unmarshal(file, &envConfiguration)
	if err != nil {
		return envConfiguration, errors.Wrap(err, "Cannot decode environment configuration")
	}

	return envConfiguration, nil
}
