package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

const defaultAddr = "localhost:8080"

type ConfigData struct {
	Addr        string `envconfig:"SERVER_ADDRESS"`
	BaseAddr    string `envconfig:"BASE_URL"`
	FileStorage string `envconfig:"FILE_STORAGE_PATH"`
	DBDSN       string `envconfig:"DATABASE_DSN"`
}

func InitConfig(args []string) (*ConfigData, error) {
	var newConfig ConfigData

	fmt.Printf("DEBUG ARGS: %+v\nDEBUG ENVS: %+v\n", args, os.Environ())

	err := envconfig.Process("", &newConfig)
	if err != nil {
		return nil, err
	}

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var needParseFlag = false
	if strings.TrimSpace(newConfig.Addr) == "" {
		flags.StringVar(&newConfig.Addr, "a", defaultAddr, "address")
		needParseFlag = true
	}

	if strings.TrimSpace(newConfig.BaseAddr) == "" {
		flags.StringVar(&newConfig.BaseAddr, "b", "", "base address")
		needParseFlag = true
	}
	if strings.TrimSpace(newConfig.FileStorage) == "" {
		flags.StringVar(&newConfig.FileStorage, "f", "./storage.json", "storage file path")
		needParseFlag = true
	}
	if strings.TrimSpace(newConfig.DBDSN) == "" {
		flags.StringVar(&newConfig.DBDSN, "d", "", "database conn string")
		needParseFlag = true
	}

	if needParseFlag {
		if err := flags.Parse(args[1:]); err != nil {
			return nil, err
		}
	}

	newConfig.CheckBaseURL()

	return &newConfig, nil
}

func (cd *ConfigData) CheckBaseURL() {
	if strings.TrimSpace(cd.BaseAddr) == "" {
		cd.BaseAddr = "http://" + cd.Addr
	}
}
