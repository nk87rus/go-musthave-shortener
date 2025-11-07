package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v6"
)

const defaultAddr = "localhost:8080"

type ConfigData struct {
	Addr        string `env:"SERVER_ADDRESS"`
	BaseAddr    string `env:"BASE_URL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
}

func InitConfig(args []string) (*ConfigData, error) {
	var newConfig ConfigData

	fmt.Printf("DEBUG FLAGS: %+v\n", args)
	for _, flg := range []string{"a", "b", "f"} {
			fmt.Printf("DEBUG ALLOWED FLAG: %s = %+v\n", flg, []rune(flg))
	}

	for _, arg := range args[1:] {
		if len(arg) == 2 {
			fmt.Printf("DEBUG INCOMING FLAG: %s = %+v\n", arg, []rune(arg))
		}
	}

	if err := env.Parse(&newConfig); err != nil {
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
