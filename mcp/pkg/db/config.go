package db

import (
	"errors"
	"time"
)

var ErrNoHostDefined = errors.New("no host defined")
var ErrNoUserDefined = errors.New("no user defined")
var ErrNoDatabaseNameDefined = errors.New("no database name defined")
var ErrNoPortDefined = errors.New("no port defined")
var ErrNoDatabaseFileDefined = errors.New("no database file defined")
var ErrInvalidDoltLiteBusyTimeout = errors.New("DoltLite busy timeout must be between 0 and 2147483647 milliseconds")

const DefaultDoltLiteBusyTimeout = 5 * time.Second

const maxDoltLiteBusyTimeout = time.Duration(1<<31-1) * time.Millisecond

type Config struct {
	DSN             string      `yaml:"dsn" json:"dsn"`
	Host            string      `yaml:"host" json:"host"`
	User            string      `yaml:"user" json:"user"`
	Password        string      `yaml:"password" json:"password"`
	DatabaseName    string      `yaml:"database_name" json:"database_name"`
	Port            int         `yaml:"port" json:"port"`
	ParseTime       bool        `yaml:"parse_time" json:"parse_time"`
	MultiStatements bool        `yaml:"multi_statements" json:"multi_statements"`
	TLS             string      `yaml:"tls" json:"tls"`
	TLSCAFile       string      `yaml:"tls_ca_file" json:"tls_ca_file"`
	DialectType     DialectType `yaml:"dialect_type" json:"dialect_type"`

	Path        string        `yaml:"path" json:"path"`
	CommitName  string        `yaml:"commit_name" json:"commit_name"`
	CommitEmail string        `yaml:"commit_email" json:"commit_email"`
	BusyTimeout time.Duration `yaml:"busy_timeout" json:"busy_timeout"`

	doltLiteDatabase *doltLiteDatabase
}

func (c *Config) Validate() error {
	if c.DialectType == DialectDoltLite && (c.BusyTimeout < 0 || (c.BusyTimeout > 0 && c.BusyTimeout < time.Millisecond) || c.BusyTimeout > maxDoltLiteBusyTimeout) {
		return ErrInvalidDoltLiteBusyTimeout
	}
	if c.DSN != "" {
		return nil
	}
	if c.DialectType == DialectDoltLite {
		if c.Path == "" {
			return ErrNoDatabaseFileDefined
		}
		return nil
	}
	if c.Host == "" {
		return ErrNoHostDefined
	}
	if c.User == "" {
		return ErrNoUserDefined
	}
	if c.Port == 0 {
		return ErrNoPortDefined
	}
	return nil
}
