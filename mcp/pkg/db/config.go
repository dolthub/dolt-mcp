package db

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNoHostDefined = errors.New("no host defined")
var ErrNoUserDefined = errors.New("no user defined")
var ErrNoDatabaseNameDefined = errors.New("no database name defined")
var ErrNoPortDefined = errors.New("no port defined")

type Config struct {
	DSN             string `yaml:"dsn" json:"dsn"`
	Host            string `yaml:"host" json:"host"`
	User            string `yaml:"user" json:"user"`
	Password        string `yaml:"password" json:"password"`
	DatabaseName    string `yaml:"database_name" json:"database_name"`
	Port            int    `yaml:"port" json:"port"`
	ParseTime       bool   `yaml:"parse_time" json:"parse_time"`
	MultiStatements bool   `yaml:"multi_statements" json:"multi_statements"`
	TLS             string `yaml:"tls" json:"tls"`
	TLSCAFile       string `yaml:"tls_ca_file" json:"tls_ca_file"`
}

func (c *Config) getDSNOptions() string {
	options := []string{}

	if c.ParseTime {
		options = append(options, "parseTime=true")
	}

	if c.MultiStatements {
		options = append(options, "multiStatements=true")
	}

	if c.TLS != "" {
		options = append(options, fmt.Sprintf("tls=%s", c.TLS))
	}

	if len(options) > 0 {
		return "?" + strings.Join(options, "&")
	}

	return ""
}

func (c *Config) GetDSN() string {
	if c.DSN != "" {
		return c.DSN
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", c.User, c.Password, c.Host, c.Port)
	if c.DatabaseName != "" {
		dsn += c.DatabaseName
	}

	dsn += c.getDSNOptions()
	return dsn
}

func (c *Config) Validate() error {
	if c.DSN == "" {
		if c.Host == "" {
			return ErrNoHostDefined
		}
		if c.User == "" {
			return ErrNoUserDefined
		}
		if c.Port == 0 {
			return ErrNoPortDefined
		}
	}
	return nil
}
