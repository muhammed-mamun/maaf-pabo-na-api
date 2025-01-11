package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// variable start with Camel case to ensure it is accessible outside from the file
type HTTPServer struct {
	Addr string `yaml:"address" env-required:"true"`
}

type OpenAIConfig struct {
	APIKEY string `yaml:"api_key" env:"GEMINI_API_KEY" env-required:"true"`
}

// env-default:"production
type Config struct {
	Env         string `yaml:"env" env:"ENV" env-required:"true"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	OpenAIAPI   OpenAIConfig `yaml:"openai_api" env-required:"true"`
}

func MustLoad() *Config {
	var configPath string
	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" {
		flags := flag.String("config", "", "path to the configuration file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			log.Fatal("Config path is not set")
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file does not exist")
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)

	if err != nil {
		log.Fatalf("can not read config file %s", err.Error())
	}

	if cfg.OpenAIAPI.APIKEY == "" {
		cfg.OpenAIAPI.APIKEY = os.Getenv("OPENAI_API_KEY")
		if cfg.OpenAIAPI.APIKEY == "" {
			log.Fatal("openai API key is not set")
		}
	}

	return &cfg
}
