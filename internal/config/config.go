package config

import (
	"flag"
	"strings"
)

type ConfigData struct {
	Addr     string
	BaseAddr string
}

func InitConfig() *ConfigData {
	newConfig := &ConfigData{}
	flag.StringVar(&newConfig.Addr, "a", "localhost:8080", "address")
	flag.StringVar(&newConfig.BaseAddr, "b", "", "base address")
	flag.Parse()

	if strings.TrimSpace(newConfig.BaseAddr) == "" {
		newConfig.BaseAddr = "http://" + newConfig.Addr
	}

	return newConfig
}
