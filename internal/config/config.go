package config

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

const defaultAddr = "localhost:8080"

type ConfigParser interface {
	ReadConfigFile(fileName string) error
	ParseEnv() error
	ParseFlags(args []string) error
	MakeConfig() ConfigData
}

type Parser struct {
	fileData  ConfigData
	flagsData ConfigData
	envData   ConfigData
}

type ConfigData struct {
	Addr          string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseAddr      string `env:"BASE_URL" json:"base_url"`
	FileStorage   string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DBDSN         string `env:"DATABASE_DSN" json:"database_dsn"`
	AuditFile     string `env:"AUDIT_FILE"`
	AuditURL      string `env:"AUDIT_URL"`
	EnableTLS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigFile    string `env:"CONFIG"`
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
}

func InitConfig(args []string) (*ConfigData, error) {
	parser := Parser{}
	if err := parser.ParseEnv(); err != nil {
		return nil, err
	}

	if err := parser.ParseFlags(args); err != nil {
		return nil, err
	}

	newConfig, err := parser.MakeConfig()
	if err != nil {
		return nil, err
	}

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
	flags.StringVar(&p.flagsData.FileStorage, "f", "", "storage file path")
	flags.StringVar(&p.flagsData.DBDSN, "d", "", "database conn string")
	flags.StringVar(&p.flagsData.AuditFile, "audit-file", "", "audit file")
	flags.StringVar(&p.flagsData.AuditURL, "audit-url", "", "audit url")
	flags.BoolVar(&p.flagsData.EnableTLS, "s", false, "enable TLS")
	flags.StringVar(&p.flagsData.ConfigFile, "c", "", "config file")
	flags.StringVar(&p.flagsData.TrustedSubnet, "t", "", "trusted subnet")
	var configOpt string
	flags.StringVar(&configOpt, "config", "", "config file")

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if !p.flagsData.EnableTLS {
		flags.Visit(func(f *flag.Flag) {
			if f.Name == "s" {
				p.flagsData.EnableTLS = true
			}
		})
	}

	if p.flagsData.ConfigFile == "" && configOpt != "" {
		p.flagsData.ConfigFile = configOpt
	}

	return nil
}

func (p *Parser) ReadConfigFile(fileName string) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &p.fileData)
}

func (p *Parser) MakeConfig() (ConfigData, error) {
	checkStringField(&p.envData.Addr, &p.flagsData.Addr)
	checkStringField(&p.envData.BaseAddr, &p.flagsData.BaseAddr)
	checkStringField(&p.envData.FileStorage, &p.flagsData.FileStorage)
	checkStringField(&p.envData.DBDSN, &p.flagsData.DBDSN)
	checkStringField(&p.envData.AuditFile, &p.flagsData.AuditFile)
	checkStringField(&p.envData.AuditURL, &p.flagsData.AuditURL)
	checkStringField(&p.envData.ConfigFile, &p.flagsData.ConfigFile)
	checkStringField(&p.envData.TrustedSubnet, &p.flagsData.TrustedSubnet)
	checkTLSFlag(&p.envData.EnableTLS, &p.flagsData.EnableTLS)
	// if p.envData.EnableTLS != p.flagsData.EnableTLS && p.flagsData.EnableTLS {
	// 	p.envData.EnableTLS = p.flagsData.EnableTLS
	// }

	if p.envData.ConfigFile != "" {
		if err := p.ReadConfigFile(p.envData.ConfigFile); err != nil {
			return ConfigData{}, err
		}
		checkStringField(&p.envData.Addr, &p.fileData.Addr)
		checkStringField(&p.envData.BaseAddr, &p.fileData.BaseAddr)
		checkStringField(&p.envData.FileStorage, &p.fileData.FileStorage)
		checkStringField(&p.envData.DBDSN, &p.fileData.DBDSN)
		checkStringField(&p.envData.TrustedSubnet, &p.fileData.TrustedSubnet)
		checkTLSFlag(&p.envData.EnableTLS, &p.fileData.EnableTLS)
	}

	return p.envData, nil
}

func checkStringField(trgField, secondConf *string) {
	if strings.TrimSpace(*trgField) == "" {
		*trgField = *secondConf
	}
}

func checkTLSFlag(firstOpt, secondOpt *bool) {
	if *firstOpt != *secondOpt && *secondOpt {
		*firstOpt = *secondOpt
	}
}
