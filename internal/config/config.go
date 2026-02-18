package config

import (
	"flag"
	"strings"

	"github.com/caarlos0/env/v6"
)

const defaultAddr = "localhost:8080"

type ConfigData struct {
	Addr        string `env:"SERVER_ADDRESS"`
	BaseAddr    string `env:"BASE_URL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	DBDSN       string `env:"DATABASE_DSN"`
	AuditFile   string `env:"AUDIT_FILE"`
	AuditURL    string `env:"AUDIT_URL"`
}

func InitConfig(args []string) (*ConfigData, error) {
	var newConfig ConfigData

	if err := env.Parse(&newConfig); err != nil {
		return nil, err
	}

	fConfig, err := parseFlags(args)
	if err != nil {
		return nil, err
	}

	checkField(&newConfig.Addr, &fConfig.Addr)
	checkField(&newConfig.BaseAddr, &fConfig.BaseAddr)
	checkField(&newConfig.FileStorage, &fConfig.FileStorage)
	checkField(&newConfig.DBDSN, &fConfig.DBDSN)
	checkField(&newConfig.AuditFile, &fConfig.AuditFile)
	checkField(&newConfig.AuditURL, &fConfig.AuditURL)

	newConfig.CheckBaseURL()

	return &newConfig, nil
}

func (cd *ConfigData) CheckBaseURL() {
	if strings.TrimSpace(cd.BaseAddr) == "" {
		cd.BaseAddr = "http://" + cd.Addr
	}
}

func parseFlags(args []string) (*ConfigData, error) {
	var fConfig ConfigData
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&fConfig.Addr, "a", defaultAddr, "address")
	flags.StringVar(&fConfig.BaseAddr, "b", "", "base address")
	flags.StringVar(&fConfig.FileStorage, "f", "./storage.json", "storage file path")
	flags.StringVar(&fConfig.DBDSN, "d", "", "database conn string")
	flags.StringVar(&fConfig.AuditFile, "audit-file", "", "audit file")
	flags.StringVar(&fConfig.AuditURL, "audit-url", "", "audit url")
	if err := flags.Parse(args[1:]); err != nil {
		return nil, err
	}
	return &fConfig, nil
}

func checkField(trgField, secondConf *string) {
	if strings.TrimSpace(*trgField) == "" {
		*trgField = *secondConf
	}
}
