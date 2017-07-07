package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Components []string

type StreamServices []string

type WebUI struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Overlay struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Twitch struct {
	ClientId   string `yaml:"client_id"`
	Username   string `yaml:"username"`
	OAuthToken string `yaml:"oauth_token"`
}

type LaMetricApp struct {
	Url         string `yaml:"url"`
	AccessToken string `yaml:"access_token"`
	Icon        string `yaml:"icon"`
}

type LaMetric struct {
	Follower   *LaMetricApp `yaml:"follower"`
	Host       *LaMetricApp `yaml:"host"`
	Bits       *LaMetricApp `yaml:"bits"`
	Subscriber *LaMetricApp `yaml:"subscriber"`
}

type Config struct {
	ListenAddr      string
	StoragePath     string
	CustomFilesPath string
	OverlayCorePath string
	MongoHosts      string         `yaml:"mongo_hosts"`
	StreamServices  StreamServices `yaml:"streams"`
	Components      Components     `yaml:"components"`
	Twitch          Twitch         `yaml:"twitch"`
	WebUI           WebUI          `yaml:"webui"`
	Overlay         Overlay        `yaml:"overlay"`
	LaMetric        *LaMetric       `yaml:"lametric"`
}

func GetTestConfig() (c *Config) {
	c = &Config{}
	c.MongoHosts = "localhost"
	return
}

func LoadConfig() (c *Config) {
	data, err := ioutil.ReadFile("settings.yaml")

	if err != nil {
		panic(err)
	}

	c = &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		panic(err)
	}

	c.ListenAddr = ":30000"
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if c.WebUI.Host == "" {
		c.WebUI.Host = "localhost"
	}

	if c.Overlay.Host == "" {
		c.Overlay.Host = "localhost"
	}

	c.OverlayCorePath = filepath.Join(wd, filepath.Join("html-overlay", "www-dist"))
	c.StoragePath = GetLocalStoragePath()
	c.CustomFilesPath = filepath.Join(c.StoragePath, "files")
	log.Printf("Components: %#v", c.Components)
	log.Printf("WebUI: %s:%d", c.WebUI.Host, c.WebUI.Port)
	log.Printf("Overlay: %s:%d", c.Overlay.Host, c.Overlay.Port)
	log.Printf("Local files will be stored at %s", c.StoragePath)
	return
}
