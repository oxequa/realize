package settings

import (
	"gopkg.in/yaml.v2"
	"syscall"
)

type Settings struct {
	Colors    `yaml:"-"`
	Resources `yaml:"resources,omitempty"`
	Server    `yaml:"server,omitempty"`
	Config    `yaml:"config,omitempty"`
}

type Config struct {
	Flimit uint64 `yaml:"flimit,omitempty"`
}

type Server struct {
	Enabled bool   `yaml:"enable,omitempty"`
	Open    bool   `yaml:"open,omitempty"`
	Host    string `yaml:"host,omitempty"`
	Port    int    `yaml:"port,omitempty"`
}

type Resources struct {
	Config string `yaml:"config,omitempty"`
	Output string `yaml:"output,omitempty"`
}

// Flimit defines the max number of watched files
func (s *Settings) Flimit() {
	var rLimit syscall.Rlimit
	rLimit.Max = s.Config.Flimit
	rLimit.Cur = s.Config.Flimit
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		s.Fatal("Error Setting Rlimit", err)
	}
}

// Read from the configuration file
func (s *Settings) Read(out interface{}) error {
	content, err := s.Stream(s.Resources.Config)
	if err == nil {
		err = yaml.Unmarshal(content, out)
		return err
	}
	return err
}

// Record create and unmarshal the yaml config file
func (h *Settings) Record(out interface{}) error {
	y, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	return h.Write(h.Resources.Config, y)
}
