package api

import "github.com/ghodss/yaml"

// Pull this into our own function so we can sub out api machinery later
func ToYaml(api interface{}) ([]byte, error) {
	return yaml.Marshal(api)
}


