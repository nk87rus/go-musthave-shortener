package config

import (
	"flag"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

const defaultAddr = "localhost:8080"

type ConfigParser interface {
	ParseEnv() error
	ParseFlags(args []string) error
	MakeConfig() ConfigData
}

type Parser struct {
	flagsData ConfigData
	envData   ConfigData
}

type ConfigData struct {
	Addr        string `env:"SERVER_ADDRESS"`
	BaseAddr    string `env:"BASE_URL"`
	FileStorage string `env:"FILE_STORAGE_PATH"`
	DBDSN       string `env:"DATABASE_DSN"`
	AuditFile   string `env:"AUDIT_FILE"`
	AuditURL    string `env:"AUDIT_URL"`
	EnableTLS   bool   `env:"ENABLE_HTTPS"`
}

func InitConfig(args []string) (*ConfigData, error) {
	parser := Parser{}
	if err := parser.ParseEnv(); err != nil {
		return nil, err
	}

	if err := parser.ParseFlags(args); err != nil {
		return nil, err
	}

	newConfig := parser.MakeConfig()
	newConfig.CheckBaseURL()

	return &newConfig, nil
}

func (cd *ConfigData) CheckBaseURL() {
	if strings.TrimSpace(cd.BaseAddr) == "" {
		scheme := "http"
		if cd.EnableTLS {
			scheme += "s"
		}
		cd.BaseAddr = scheme + "://" + cd.Addr
	}
}

func (p *Parser) ParseEnv() error {
	if err := env.Parse(&p.envData); err != nil {
		return err
	}

	_, p.envData.EnableTLS = os.LookupEnv("ENABLE_HTTPS")
	return nil
}

func (p *Parser) ParseFlags(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	flags.StringVar(&p.flagsData.Addr, "a", defaultAddr, "address")
	flags.StringVar(&p.flagsData.BaseAddr, "b", "", "base address")
	flags.StringVar(&p.flagsData.FileStorage, "f", "./storage.json", "storage file path")
	flags.StringVar(&p.flagsData.DBDSN, "d", "", "database conn string")
	flags.StringVar(&p.flagsData.AuditFile, "audit-file", "", "audit file")
	flags.StringVar(&p.flagsData.AuditURL, "audit-url", "", "audit url")
	flags.BoolVar(&p.flagsData.EnableTLS, "s", false, "enable TLS")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if !p.flagsData.EnableTLS {
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "s" {
				p.flagsData.EnableTLS = true
			}
		})
	}

	return nil
}

func (p *Parser) MakeConfig() ConfigData {
	checkStringField(&p.envData.Addr, &p.flagsData.Addr)
	checkStringField(&p.envData.BaseAddr, &p.flagsData.BaseAddr)
	checkStringField(&p.envData.FileStorage, &p.flagsData.FileStorage)
	checkStringField(&p.envData.DBDSN, &p.flagsData.DBDSN)
	checkStringField(&p.envData.AuditFile, &p.flagsData.AuditFile)
	checkStringField(&p.envData.AuditURL, &p.flagsData.AuditURL)

	if p.envData.EnableTLS != p.flagsData.EnableTLS && p.flagsData.EnableTLS {
		p.envData.EnableTLS = p.flagsData.EnableTLS
	}

	return p.envData
}

func checkStringField(trgField, secondConf *string) {
	if strings.TrimSpace(*trgField) == "" {
		*trgField = *secondConf
	}
}
