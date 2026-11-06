package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	config *Config
)

type Config struct {
	App        AppConfig
	API        APIConfig
	ServerInfo ServerInfoConfig
}

type AppConfig struct {
	RootDir     string
	Environment string
	// LogFormat is json, text or color. Empty means json in production, text otherwise.
	LogFormat string
	// LogLevel is info or debug. Empty means info.
	LogLevel string
}

type AllowedOriginsConfig struct {
	App []string `json:"app"`
}

type APIConfig struct {
	App                  int
	AllowedOriginsConfig AllowedOriginsConfig
}

type ServerInfoConfig struct {
	Hostname string
}

func init() {
	// Load configuration file
	if err := Load("/etc/app/config.json"); err == nil {
		return
	}
	if err := Load("./etc/app/config.json"); err == nil {
		return
	}
	if err := Load("../etc/app/config.json"); err == nil {
		return
	}
	panic("[main] Failed to load configuration file")
}

func Load(path string) error {
	handleErr := func(err error) error {
		fmt.Fprintf(os.Stderr, "[config] Cannot open config file in %s, reason=%+v\n", path, err)
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return handleErr(err)
	}
	defer file.Close()
	config = &Config{}
	// update server info
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	config.ServerInfo.Hostname = hostname
	// Load configuration file
	err = json.NewDecoder(file).Decode(config)
	if err != nil {
		return handleErr(err)
	}
	fmt.Printf("[config] Loaded config file %s\n", path)
	loadFromEnv(config)
	return nil
}

func mustInt(v string) int {
	i, err := strconv.Atoi(v)
	if err != nil {
		panic(err)
	}
	return i
}

func ifExist(env string, overrideFn func(value string)) {
	if v := os.Getenv(env); len(v) > 0 {
		fmt.Printf("[config] Found %s=%s, override existing config\n", env, v)
		overrideFn(v)
	}
}

func loadFromEnv(config *Config) {
	// App setting
	ifExist("APP_ROOT_DIR", func(v string) {
		config.App.RootDir = v
	})
	ifExist("APP_ENV", func(v string) {
		config.App.Environment = v
	})
	ifExist("LOG_FORMAT", func(v string) {
		config.App.LogFormat = strings.ToLower(v)
	})
	ifExist("LOG_LEVEL", func(v string) {
		config.App.LogLevel = strings.ToLower(v)
	})

	// API setting
	ifExist("API_APP_PORT", func(v string) {
		config.API.App = mustInt(v)
	})
}

func GetConfig() Config {
	return *config
}
