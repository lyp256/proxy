package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/lyp256/proxy/pkg/auth"
	"github.com/lyp256/proxy/pkg/registry"
)

type Config struct {
	LOG string `yaml:"log"`
	TLS struct {
		Cert string `yaml:"cert"`
		Key  string `yaml:"key"`
	} `yaml:"tls"`
	HTTP        string `yaml:"http"`
	HTTPS       string `yaml:"https"`
	EnableHTTP3 bool   `yaml:"enable-http3"`
	EnableHTTP2 bool   `yaml:"enable-http2"`
	Insecure    bool   `yaml:"insecure"`
	// auth users
	Users     []auth.User `yaml:"users"`
	Component struct {
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
		// http proxy
		Proxy struct {
			Enable bool   `yaml:"enable"`
			Auth   string `yaml:"auth"`
			// 主动要求 proxy auth
			Active bool `yaml:"active"`
		}
	} `yaml:"component"`
}

func (c *Config) LoadFile(f string) error {
	buf, err := os.ReadFile(f)
	if err != nil {
		return fmt.Errorf("open config %s:%s", f, err)
	}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return fmt.Errorf("invalid config %s:%s", f, err)
	}
	return nil
}

func (c *Config) Unmarshal(buf []byte) error {
	err := yaml.Unmarshal(buf, &c)
	if err != nil {
		return fmt.Errorf("invalid config %s", err)
	}
	return nil
}

func (c *Config) Marshal() []byte {
	buf, _ := yaml.Marshal(c)
	return buf
}
