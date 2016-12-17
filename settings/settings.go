package settings

import (
	"gopkg.in/yaml.v2"
	"os"
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

// Read from config file
func (s *Settings) Read(out interface{}) error {
	localConfigPath := s.Resources.Config
	if _, err := os.Stat(".realize/" + s.Resources.Config); err == nil {
		localConfigPath = ".realize/" + s.Resources.Config
	}
	content, err := s.Stream(localConfigPath)
	if err == nil {
		err = yaml.Unmarshal(content, out)
		return err
	}
	return err
}

// Record create and unmarshal the yaml config file
func (s *Settings) Record(out interface{}) error {
	y, err := yaml.Marshal(out)
	if err != nil {
		return err
	}
	if _, err := os.Stat(".realize/"); os.IsNotExist(err) {
		if err = os.Mkdir(".realize/", 0770); err != nil {
			return s.Write(s.Resources.Config, y)
		}
	}
	return s.Write(".realize/"+s.Resources.Config, y)
}
