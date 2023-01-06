package config

import (
	"github.com/lyp256/proxy/pkg/registry"
)

func Default() *Config {
	return &Config{
		LOG:         "warning",
		HTTP:        ":80",
		HTTPS:       ":443",
		EnableHTTP3: true,
		EnableHTTP2: true,
		Insecure:    false,
		Users:       nil,
		Component: struct {
			Static struct {
				Enable bool   `yaml:"enable"`
				Root   string `yaml:"root"`
			} `yaml:"static,omitempty"`
			Registry struct {
				Enable bool `yaml:"enable"`
				registry.Config
			} `yaml:"registry"`
			Vless struct {
				Path   string `yaml:"path"`
				Enable bool   `yaml:"enable"`
				Auth   string `yaml:"auth"`
			} `yaml:"vless"`
			Proxy struct {
				Enable bool   `yaml:"enable"`
				Auth   string `yaml:"auth"`
				Active bool   `yaml:"active"`
			}
		}{},
	}
}
