package settings

import (
	"gopkg.in/yaml.v2"
	"syscall"
)

type Settings struct {
	Colors    `yaml:"-"`
	Resources `yaml:"resources" json:"resources"`
	Server    `yaml:"server" json:"server"`
	Config    `yaml:"config" json:"config"`
}

type Config struct {
	Flimit uint64 `yaml:"flimit" json:"flimit"`
}

type Server struct {
	Enabled bool   `yaml:"enable" json:"enable"`
	Open    bool   `yaml:"open" json:"open"`
	Host    string `yaml:"host" json:"host"`
	Port    int    `yaml:"port" json:"port"`
}

type Resources struct {
	Config string `yaml:"-" json:"-"`
	Output string `yaml:"output" json:"output"`
	Log    string `yaml:"log" json:"log"`
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
